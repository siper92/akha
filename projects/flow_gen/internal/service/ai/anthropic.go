package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/siper92/akha/projects/flow_gen/internal/cfg"
	"github.com/siper92/akha/projects/flow_gen/internal/debug"
)

type Anthropic struct {
	client    anthropic.Client
	model     string
	maxTokens int64
	maxRounds int
}

var _ Client = (*Anthropic)(nil)

func NewAnthropic() *Anthropic {
	return &Anthropic{
		client:    anthropic.NewClient(),
		model:     cfg.GetModel(),
		maxTokens: cfg.GetMaxTokens(),
		maxRounds: cfg.GetMaxToolRounds(),
	}
}

func (a *Anthropic) Run(ctx context.Context, req Request) (Response, error) {
	tools := make(map[string]Tool, len(req.Tools))
	for _, tool := range req.Tools {
		tools[tool.Name()] = tool
	}
	return conversation{
		api:   a,
		req:   req,
		tools: tools,
		messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
		},
	}.exec(ctx)
}

type conversation struct {
	api      *Anthropic
	req      Request
	tools    map[string]Tool
	messages []anthropic.MessageParam
}

func (c conversation) exec(ctx context.Context) (Response, error) {
	out := Response{}
	for round := 0; round < c.api.maxRounds; round++ {
		msg, err := c.api.client.Messages.New(ctx, c.params())
		if err != nil {
			return out, fmt.Errorf("ai: message: %w", err)
		}
		debug.Service().Log("ai: round %d stop=%s", round+1, msg.StopReason)

		switch msg.StopReason {
		case anthropic.StopReasonRefusal:
			return out, ErrRefused
		case anthropic.StopReasonMaxTokens:
			return out, ErrTruncated
		}

		c.messages = append(c.messages, msg.ToParam())
		text, results := c.process(ctx, msg)
		out.Text = text
		out.ToolCalls += len(results)

		if msg.StopReason != anthropic.StopReasonToolUse {
			return out, nil
		}
		c.messages = append(c.messages, anthropic.NewUserMessage(results...))
	}
	return out, ErrMaxRounds
}

func (c conversation) params() anthropic.MessageNewParams {
	params := anthropic.MessageNewParams{
		Model:     c.api.model,
		MaxTokens: c.api.maxTokens,
		Messages:  c.messages,
		Tools:     c.toolParams(),
	}
	if c.req.System != "" {
		params.System = []anthropic.TextBlockParam{{Text: c.req.System}}
	}

	return params
}

func (c conversation) toolParams() []anthropic.ToolUnionParam {
	params := make([]anthropic.ToolUnionParam, 0, len(c.req.Tools))
	for _, tool := range c.req.Tools {
		schema := tool.Schema()
		param := anthropic.ToolParam{
			Name:        tool.Name(),
			Description: anthropic.String(tool.Description()),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: schema.Properties,
				Required:   schema.Required,
			},
		}
		params = append(params, anthropic.ToolUnionParam{OfTool: &param})
	}
	return params
}

func (c conversation) process(ctx context.Context, msg *anthropic.Message) (string, []anthropic.ContentBlockParamUnion) {
	var text strings.Builder
	results := []anthropic.ContentBlockParamUnion{}
	for _, block := range msg.Content {
		switch variant := block.AsAny().(type) {
		case anthropic.TextBlock:
			text.WriteString(variant.Text)
		case anthropic.ToolUseBlock:
			results = append(results, c.call(ctx, variant))
		}
	}
	return text.String(), results
}

func (c conversation) call(ctx context.Context, use anthropic.ToolUseBlock) anthropic.ContentBlockParamUnion {
	tool, ok := c.tools[use.Name]
	if !ok {
		debug.Service().Log("ai: tool %s unknown", use.Name)
		return anthropic.NewToolResultBlock(use.ID, "unknown tool: "+use.Name, true)
	}
	result, err := tool.Call(ctx, json.RawMessage(use.JSON.Input.Raw()))
	if err != nil {
		debug.Service().Log("ai: tool %s failed: %v", use.Name, err)
		return anthropic.NewToolResultBlock(use.ID, err.Error(), true)
	}
	debug.Service().Log("ai: tool %s ok: %s", use.Name, result)
	return anthropic.NewToolResultBlock(use.ID, result, false)
}

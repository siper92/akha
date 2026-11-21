package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var _ Agent = (*ClaudeAgent)(nil)

type ClaudeAgent struct {
	client        anthropic.Client
	session       *mcp.ClientSession
	model         string
	maxIterations int
	log           *slog.Logger
}

func NewClaudeAgent(client anthropic.Client, session *mcp.ClientSession, model string, log *slog.Logger) *ClaudeAgent {
	if model == "" {
		model = DefaultModel
	}
	if log == nil {
		log = slog.Default()
	}
	return &ClaudeAgent{client: client, session: session, model: model, maxIterations: 10, log: log}
}

func (a *ClaudeAgent) Run(ctx context.Context, prompt string) (string, error) {
	tools, err := a.tools(ctx)
	if err != nil {
		return "", err
	}

	messages := []anthropic.BetaMessageParam{
		anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(prompt)),
	}

	for i := 0; i < a.maxIterations; i++ {
		resp, err := a.client.Beta.Messages.New(ctx, anthropic.BetaMessageNewParams{
			Model:     a.model,
			MaxTokens: 16000,
			Messages:  messages,
			Tools:     tools,
			Fallbacks: anthropic.BetaFallbacksParamOfDefault(),
			Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
		})
		if err != nil {
			return "", fmt.Errorf("messages.new: %w", err)
		}
		if resp.StopReason == anthropic.BetaStopReasonRefusal {
			return "", ErrRefused
		}

		messages = append(messages, resp.ToParam())

		var text strings.Builder
		var results []anthropic.BetaContentBlockParamUnion
		for _, block := range resp.Content {
			switch b := block.AsAny().(type) {
			case anthropic.BetaTextBlock:
				text.WriteString(b.Text)
			case anthropic.BetaToolUseBlock:
				results = append(results, a.callTool(ctx, b.ID, b.Name, b.JSON.Input.Raw()))
			}
		}

		if resp.StopReason != anthropic.BetaStopReasonToolUse {
			return text.String(), nil
		}
		messages = append(messages, anthropic.NewBetaUserMessage(results...))
	}

	return "", ErrMaxIterations
}

func (a *ClaudeAgent) tools(ctx context.Context) ([]anthropic.BetaToolUnionParam, error) {
	res, err := a.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	out := make([]anthropic.BetaToolUnionParam, 0, len(res.Tools))
	for _, t := range res.Tools {
		schema, err := toInputSchema(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("tool %s schema: %w", t.Name, err)
		}
		tool := anthropic.BetaToolParam{
			Name:        t.Name,
			Description: anthropic.String(t.Description),
			InputSchema: schema,
		}
		out = append(out, anthropic.BetaToolUnionParam{OfTool: &tool})
	}
	return out, nil
}

func (a *ClaudeAgent) callTool(ctx context.Context, id, name, rawInput string) anthropic.BetaContentBlockParamUnion {
	a.log.InfoContext(ctx, "tool call", "name", name)
	res, err := a.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      name,
		Arguments: json.RawMessage(rawInput),
	})
	if err != nil {
		return toolResult(id, err.Error(), true)
	}
	var sb strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	return toolResult(id, sb.String(), res.IsError)
}

func toolResult(id, text string, isError bool) anthropic.BetaContentBlockParamUnion {
	return anthropic.BetaContentBlockParamUnion{OfToolResult: &anthropic.BetaToolResultBlockParam{
		ToolUseID: id,
		IsError:   anthropic.Bool(isError),
		Content: []anthropic.BetaToolResultBlockParamContentUnion{
			{OfText: &anthropic.BetaTextBlockParam{Text: text}},
		},
	}}
}

func toInputSchema(s any) (anthropic.BetaToolInputSchemaParam, error) {
	var out anthropic.BetaToolInputSchemaParam
	raw, err := json.Marshal(s)
	if err != nil {
		return out, err
	}

	var parsed struct {
		Properties map[string]any `json:"properties"`
		Required   []string       `json:"required"`
	}
	if err = json.Unmarshal(raw, &parsed); err != nil {
		return out, err
	}

	if parsed.Properties == nil {
		parsed.Properties = map[string]any{}
	}
	out.Properties = parsed.Properties
	out.Required = parsed.Required

	return out, nil
}

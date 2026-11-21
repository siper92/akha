package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

var _ Agent = (*ClaudeAgent)(nil)

type ClaudeAgent struct {
	client        anthropic.Client
	tools         ToolService
	model         string
	maxIterations int
	log           *slog.Logger
}

func NewClaudeAgent(client anthropic.Client, tools ToolService, model string, log *slog.Logger) *ClaudeAgent {
	if model == "" {
		model = DefaultModel
	}
	if log == nil {
		log = slog.Default()
	}
	return &ClaudeAgent{client: client, tools: tools, model: model, maxIterations: 10, log: log}
}

func (a *ClaudeAgent) Run(ctx context.Context, prompt string) (string, error) {
	messages := []anthropic.BetaMessageParam{
		anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(prompt)),
	}

	for i := 0; i < a.maxIterations; i++ {
		resp, err := a.client.Beta.Messages.New(ctx, anthropic.BetaMessageNewParams{
			Model:     a.model,
			MaxTokens: 16000,
			Betas:     []anthropic.AnthropicBeta{"server-side-fallback-2026-07-01"},
			Fallbacks: anthropic.BetaFallbacksParamOfDefault(),
			Messages:  messages,
			Tools:     a.tools.Params(),
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
				results = append(results, a.call(ctx, b.ID, b.Name, json.RawMessage(b.JSON.Input.Raw())))
			}
		}

		if resp.StopReason != anthropic.BetaStopReasonToolUse {
			return text.String(), nil
		}
		messages = append(messages, anthropic.NewBetaUserMessage(results...))
	}

	return "", ErrMaxIterations
}

func (a *ClaudeAgent) call(ctx context.Context, id, name string, input json.RawMessage) anthropic.BetaContentBlockParamUnion {
	a.log.InfoContext(ctx, "tool call", "name", name)
	out, err := a.tools.Call(ctx, name, input)
	if err != nil {
		return toolResult(id, err.Error(), true)
	}
	return toolResult(id, out, false)
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

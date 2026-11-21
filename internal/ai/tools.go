package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

const maxToolRounds = 10

var ErrToolLoop = errors.New("tool loop limit reached")

type Tool struct {
	Name        string
	Description string
	InputSchema map[string]any
}

type ToolRunner interface {
	Tools(ctx context.Context) ([]Tool, error)
	Call(ctx context.Context, name string, input json.RawMessage) (string, error)
}

func (s *Service) AskWithTools(ctx context.Context, model anthropic.Model, prompt string, tools ToolRunner) (string, error) {
	defs, err := tools.Tools(ctx)
	if err != nil {
		return "", fmt.Errorf("list tools: %w", err)
	}

	params := make([]anthropic.ToolUnionParam, 0, len(defs))
	for _, d := range defs {
		p := anthropic.ToolParam{
			Name:        d.Name,
			Description: anthropic.String(d.Description),
			InputSchema: anthropic.ToolInputSchemaParam{
				Properties: d.InputSchema["properties"],
				Required:   requiredFields(d.InputSchema["required"]),
			},
		}
		params = append(params, anthropic.ToolUnionParam{OfTool: &p})
	}

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
	}

	for range maxToolRounds {
		message, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
			Model:     model,
			MaxTokens: 4096,
			Messages:  messages,
			Tools:     params,
		})
		if err != nil {
			return "", fmt.Errorf("messages: %w", err)
		}
		messages = append(messages, message.ToParam())

		var text strings.Builder
		var results []anthropic.ContentBlockParamUnion
		for _, block := range message.Content {
			switch b := block.AsAny().(type) {
			case anthropic.TextBlock:
				text.WriteString(b.Text)
			case anthropic.ToolUseBlock:
				out, callErr := tools.Call(ctx, b.Name, b.Input)
				if callErr != nil {
					results = append(results, anthropic.NewToolResultBlock(b.ID, callErr.Error(), true))
					continue
				}
				results = append(results, anthropic.NewToolResultBlock(b.ID, out, false))
			}
		}

		if message.StopReason != anthropic.StopReasonToolUse {
			return text.String(), nil
		}
		messages = append(messages, anthropic.NewUserMessage(results...))
	}

	return "", ErrToolLoop
}

func requiredFields(v any) []string {
	switch r := v.(type) {
	case []string:
		return r
	case []any:
		out := make([]string, 0, len(r))
		for _, item := range r {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

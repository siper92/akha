package site

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/siper92/akha/internal/ai"
)

type OutputFormat string

const (
	OutputText OutputFormat = "text"
	OutputJSON OutputFormat = "json"
)

var ErrOutputFormat = errors.New("answer does not match output format")

type OutputConfig struct {
	Format   OutputFormat
	Schema   string
	Validate func(answer string) error
}

func (c OutputConfig) Prompt(prompt string) string {
	if c.Format != OutputJSON {
		return prompt
	}

	return prompt + "\n\nrespond only with a valid JSON object, no markdown fences and no text outside the JSON, in this exact format:\n" + c.Schema
}

func (c OutputConfig) Check(answer string) (string, error) {
	if c.Format != OutputJSON {
		return answer, nil
	}

	answer = trimJSONFence(answer)
	if !json.Valid([]byte(answer)) {
		return "", fmt.Errorf("%w: invalid json", ErrOutputFormat)
	}

	if c.Validate != nil {
		if err := c.Validate(answer); err != nil {
			return "", fmt.Errorf("%w: %w", ErrOutputFormat, err)
		}
	}

	return answer, nil
}

func trimJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}

	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")

	return strings.TrimSpace(s)
}

type asker interface {
	AskLv1(ctx context.Context, prompt string, out OutputConfig) (string, error)
	AskLv2(ctx context.Context, prompt string, out OutputConfig) (string, error)
	AskLv3(ctx context.Context, prompt string, out OutputConfig) (string, error)
}

var _ asker = (*aiAsker)(nil)

type aiAsker struct {
	s *ai.Service
}

func newAiAsker(s *ai.Service) *aiAsker {
	return &aiAsker{s: s}
}

func (a *aiAsker) AskLv1(ctx context.Context, prompt string, out OutputConfig) (string, error) {
	return a.s.AskLv1(ctx, out.Prompt(prompt))
}

func (a *aiAsker) AskLv2(ctx context.Context, prompt string, out OutputConfig) (string, error) {
	return a.s.AskLv2(ctx, out.Prompt(prompt))
}

func (a *aiAsker) AskLv3(ctx context.Context, prompt string, out OutputConfig) (string, error) {
	return a.s.AskLv3(ctx, out.Prompt(prompt))
}

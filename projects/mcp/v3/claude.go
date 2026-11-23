package v3

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go/v2"

	"github.com/siper92/akha/internal/ai"
)

var _ Asker = (*Claude)(nil)

type Claude struct {
	service *ai.Service
	model   anthropic.Model
}

func NewClaude(apiKey string, model string) (*Claude, error) {
	service, err := ai.NewAnthropic(apiKey)
	if err != nil {
		return nil, err
	}

	return &Claude{service: service, model: anthropic.Model(model)}, nil
}

func (c *Claude) Ask(ctx context.Context, prompt string, tools ai.ToolRunner) (string, error) {
	return c.service.AskWithTools(ctx, c.model, prompt, tools)
}

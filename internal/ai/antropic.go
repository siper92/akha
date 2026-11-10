package ai

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
)

type Service struct {
	client  anthropic.Client
	mu      sync.Mutex
	history []anthropic.MessageParam
}

func NewAnthropic(apiKey string) (*Service, error) {
	if apiKey == "" {
		return nil, errors.New("claude: API key is empty")
	}

	return &Service{
		client: anthropic.NewClient(
			option.WithAPIKey(apiKey),
		),
	}, nil
}

func (s *Service) Ask(ctx context.Context, question string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages := append(s.history, anthropic.NewUserMessage(anthropic.NewTextBlock(question)))

	message, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeOpus4_5,
		MaxTokens: 180000,
		System: []anthropic.TextBlockParam{
			{Text: "You are a helpful assistant. Answer concisely."},
		},
		Messages: messages,
	})
	if err != nil {
		return "", err
	}

	var answer strings.Builder
	for _, block := range message.Content {
		if block.Type == "text" {
			answer.WriteString(block.Text)
		}
	}

	s.history = append(messages, message.ToParam())
	return answer.String(), nil
}

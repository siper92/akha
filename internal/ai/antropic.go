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

const MaxTokenCount = 150000

func (s *Service) AskLv1(ctx context.Context, question string) (string, error) {
	return s.ask(ctx, question, anthropic.ModelClaudeHaiku4_5, MaxTokenCount)
}

func (s *Service) AskLv2(ctx context.Context, question string) (string, error) {
	return s.ask(ctx, question, anthropic.ModelClaudeSonnet5, MaxTokenCount)
}

func (s *Service) AskLv3(ctx context.Context, question string) (string, error) {
	return s.ask(ctx, question, anthropic.ModelClaudeOpus5, MaxTokenCount)
}

func (s *Service) ask(ctx context.Context, question string, model anthropic.Model, tokenCount int64) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	messages := append(s.history, anthropic.NewUserMessage(anthropic.NewTextBlock(question)))

	stream := s.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     model,
		MaxTokens: tokenCount,
		System: []anthropic.TextBlockParam{
			{Text: "You are a helpful assistant. Answer concisely."},
		},
		Messages: messages,
	})

	message := anthropic.Message{}
	for stream.Next() {
		if err := message.Accumulate(stream.Current()); err != nil {
			return "", err
		}
	}

	if err := stream.Err(); err != nil {
		return "", err
	}

	var answer strings.Builder
	for _, block := range message.Content {
		if block.Type == "text" {
			answer.WriteString(block.Text)
		}
	}

	return answer.String(), nil
}

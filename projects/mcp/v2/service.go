package v2

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

var _ ToolService = (*Service)(nil)

type Service struct {
	tools map[string]Tool
	order []string
}

func NewService(tools ...Tool) (*Service, error) {
	s := &Service{tools: make(map[string]Tool, len(tools))}
	if err := s.Add(tools...); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) Add(tools ...Tool) error {
	for _, t := range tools {
		name := t.Name()
		if name == "" {
			return ErrToolName
		}
		if _, ok := s.tools[name]; ok {
			return fmt.Errorf("%w: %s", ErrToolExists, name)
		}

		s.tools[name] = t
		s.order = append(s.order, name)
	}

	return nil
}

func (s *Service) Params() []anthropic.BetaToolUnionParam {
	out := make([]anthropic.BetaToolUnionParam, 0, len(s.order))
	for _, name := range s.order {
		p := s.tools[name].Param()
		out = append(out, anthropic.BetaToolUnionParam{OfTool: &p})
	}

	return out
}

func (s *Service) Call(ctx context.Context, name string, input json.RawMessage) (string, error) {
	t, ok := s.tools[name]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownTool, name)
	}

	return t.Call(ctx, input)
}

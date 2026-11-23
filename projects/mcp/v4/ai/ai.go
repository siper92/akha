package ai

import (
	"context"
	"encoding/json"
	"errors"
)

var (
	ErrRefused   = errors.New("ai: request refused")
	ErrTruncated = errors.New("ai: response hit max tokens")
	ErrMaxRounds = errors.New("ai: max tool rounds reached")
	ErrBadInput  = errors.New("ai: bad tool input")
)

type Client interface {
	Run(ctx context.Context, req Request) (Response, error)
}

type Tool interface {
	Name() string
	Description() string
	Schema() Schema
	Call(ctx context.Context, input json.RawMessage) (string, error)
}

type Schema struct {
	Properties map[string]any
	Required   []string
}

type Request struct {
	System string
	Prompt string
	Tools  []Tool
}

type Response struct {
	Text      string
	ToolCalls int
}

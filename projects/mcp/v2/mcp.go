package v2

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

const DefaultModel = "claude-opus-5-5"

var (
	ErrRefused       = errors.New("model refused the request")
	ErrMaxIterations = errors.New("max tool iterations reached")
	ErrUnknownTool   = errors.New("unknown tool")
	ErrBadInput      = errors.New("invalid tool input")
)

type Agent interface {
	Run(ctx context.Context, prompt string) (string, error)
}

type Tool interface {
	Name() string
	Param() anthropic.BetaToolParam
	Call(ctx context.Context, input json.RawMessage) (string, error)
}

type ToolService interface {
	Params() []anthropic.BetaToolUnionParam
	Call(ctx context.Context, name string, input json.RawMessage) (string, error)
}

package ai

import (
	"context"
	"encoding/json"
	"fmt"
)

type FuncTool[In any] struct {
	name        string
	description string
	schema      Schema
	fn          func(ctx context.Context, in In) (string, error)
}

var _ Tool = (*FuncTool[struct{}])(nil)

func NewTool[In any](name string, description string, schema Schema, fn func(ctx context.Context, in In) (string, error)) *FuncTool[In] {
	return &FuncTool[In]{name: name, description: description, schema: schema, fn: fn}
}

func (t *FuncTool[In]) Name() string { return t.name }

func (t *FuncTool[In]) Description() string { return t.description }

func (t *FuncTool[In]) Schema() Schema { return t.schema }

func (t *FuncTool[In]) Call(ctx context.Context, input json.RawMessage) (string, error) {
	var in In
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("%w: %s: %v", ErrBadInput, t.name, err)
	}
	return t.fn(ctx, in)
}

package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

var _ Tool = (*FuncTool[struct{}])(nil)

type FuncTool[In any] struct {
	spec ToolSpec
	fn   ToolFunc[In]
}

func NewTool[In any](spec ToolSpec, fn ToolFunc[In]) *FuncTool[In] {
	return &FuncTool[In]{spec: spec, fn: fn}
}

func (t *FuncTool[In]) Name() string { return t.spec.Name }

func (t *FuncTool[In]) Param() anthropic.BetaToolParam {
	props := make(map[string]any, len(t.spec.Props))
	var required []string
	for _, name := range slices.Sorted(maps.Keys(t.spec.Props)) {
		p := t.spec.Props[name]
		props[name] = map[string]any{"type": p.Type, "description": p.Description}
		if p.Required {
			required = append(required, name)
		}
	}

	return anthropic.BetaToolParam{
		Name:        t.spec.Name,
		Description: anthropic.String(t.spec.Description),
		InputSchema: anthropic.BetaToolInputSchemaParam{
			Properties: props,
			Required:   required,
		},
	}
}

func (t *FuncTool[In]) Call(ctx context.Context, input json.RawMessage) (string, error) {
	var in In
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBadInput, err)
	}

	return t.fn(ctx, in)
}

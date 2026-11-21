package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go/v2"
)

var _ Tool = (*CreateFile)(nil)

type CreateFile struct {
	dir string
}

func NewCreateFile(dir string) *CreateFile {
	return &CreateFile{dir: dir}
}

func (t *CreateFile) Name() string { return "createFile" }

func (t *CreateFile) Param() anthropic.BetaToolParam {
	return anthropic.BetaToolParam{
		Name:        t.Name(),
		Description: anthropic.String("creates a file with the given name and content"),
		InputSchema: anthropic.BetaToolInputSchemaParam{
			Properties: map[string]any{
				"name":    map[string]any{"type": "string", "description": "file name, e.g. part1.txt"},
				"content": map[string]any{"type": "string", "description": "file content"},
			},
			Required: []string{"name", "content"},
		},
	}
}

func (t *CreateFile) Call(_ context.Context, input json.RawMessage) (string, error) {
	var in struct {
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBadInput, err)
	}
	if in.Name == "" {
		return "", fmt.Errorf("%w: empty name", ErrBadInput)
	}
	if err := os.MkdirAll(t.dir, 0o755); err != nil {
		return "", err
	}
	root, err := os.OpenRoot(t.dir)
	if err != nil {
		return "", err
	}
	defer root.Close()
	if err := root.WriteFile(in.Name, []byte(in.Content), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("created %s (%d bytes)", in.Name, len(in.Content)), nil
}

package v3

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/siper92/akha/internal/ai"
)

var _ ai.ToolRunner = (*FileTools)(nil)

type FileTools struct {
	fs    FS
	debug Debug
}

type fileInput struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

func NewFileTools(fs FS, debug Debug) *FileTools {
	return &FileTools{fs: fs, debug: debug}
}

func (t *FileTools) Tools(_ context.Context) ([]ai.Tool, error) {
	return []ai.Tool{
		fileTool("createFile", "creates or replaces a file with the given content", "fileName", "content"),
		fileTool("readFile", "returns the content of a file", "fileName"),
		fileTool("deleteFile", "deletes a file", "fileName"),
	}, nil
}

func (t *FileTools) Call(ctx context.Context, name string, input json.RawMessage) (string, error) {
	var in fileInput
	if err := json.Unmarshal(input, &in); err != nil {
		return "", fmt.Errorf("%w: %v", ErrBadInput, err)
	}
	if in.FileName == "" {
		return "", fmt.Errorf("%w: empty fileName", ErrBadInput)
	}
	t.debug.Log("tool %s: %s", name, in.FileName)

	switch name {
	case "createFile":
		return "created " + in.FileName, t.fs.CreateFile(ctx, in.FileName, in.Content)
	case "readFile":
		return t.fs.ReadFile(ctx, in.FileName)
	case "deleteFile":
		return "deleted " + in.FileName, t.fs.DeleteFile(ctx, in.FileName)
	}

	return "", fmt.Errorf("%w: %s", ErrUnknownTool, name)
}

func fileTool(name, description string, props ...string) ai.Tool {
	properties := make(map[string]any, len(props))
	for _, prop := range props {
		properties[prop] = map[string]any{"type": "string"}
	}

	return ai.Tool{
		Name:        name,
		Description: description,
		InputSchema: map[string]any{"properties": properties, "required": props},
	}
}

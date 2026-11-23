package tools

import (
	"context"

	"aiflow/ai"
	"aiflow/vfs"
)

type createFileInput struct {
	FileName string `json:"fileName"`
	Content  string `json:"content"`
}

type fileInput struct {
	FileName string `json:"fileName"`
}

var stringType = map[string]any{"type": "string"}

func Files(fs vfs.FS) []ai.Tool {
	return []ai.Tool{
		ai.NewTool(
			"createFile",
			"Create or overwrite a text file with the given content.",
			ai.Schema{
				Properties: map[string]any{"fileName": stringType, "content": stringType},
				Required:   []string{"fileName", "content"},
			},
			func(ctx context.Context, in createFileInput) (string, error) {
				return "created " + in.FileName, fs.WriteFile(ctx, in.FileName, in.Content)
			},
		),
		ai.NewTool(
			"readFile",
			"Read a text file and return its content.",
			ai.Schema{
				Properties: map[string]any{"fileName": stringType},
				Required:   []string{"fileName"},
			},
			func(ctx context.Context, in fileInput) (string, error) {
				return fs.ReadFile(ctx, in.FileName)
			},
		),
		ai.NewTool(
			"deleteFile",
			"Delete a file.",
			ai.Schema{
				Properties: map[string]any{"fileName": stringType},
				Required:   []string{"fileName"},
			},
			func(ctx context.Context, in fileInput) (string, error) {
				return "deleted " + in.FileName, fs.DeleteFile(ctx, in.FileName)
			},
		),
	}
}

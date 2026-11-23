package v2

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type createFileIn struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type readFileIn struct {
	Name string `json:"name"`
}

func NewCreateFile(dir string) Tool {
	return NewTool(ToolSpec{
		Name:        "createFile",
		Description: "creates a file with the given name and content",
		Props: map[string]Prop{
			"name":    {Type: "string", Description: "file name relative to the work dir, e.g. part1.txt", Required: true},
			"content": {Type: "string", Description: "file content", Required: true},
		},
	}, func(_ context.Context, in createFileIn) (string, error) {
		return inRoot(dir, in.Name, func(r *os.Root) (string, error) {
			if err := r.MkdirAll(filepath.Dir(in.Name), 0o755); err != nil {
				return "", err
			}
			if err := r.WriteFile(in.Name, []byte(in.Content), 0o644); err != nil {
				return "", err
			}

			return fmt.Sprintf("created %s (%d bytes)", in.Name, len(in.Content)), nil
		})
	})
}

func NewReadFile(dir string) Tool {
	return NewTool(ToolSpec{
		Name:        "readFile",
		Description: "reads a file from the work dir",
		Props: map[string]Prop{
			"name": {Type: "string", Description: "file name relative to the work dir", Required: true},
		},
	}, func(_ context.Context, in readFileIn) (string, error) {
		return inRoot(dir, in.Name, func(r *os.Root) (string, error) {
			b, err := r.ReadFile(in.Name)
			if err != nil {
				return "", err
			}

			return string(b), nil
		})
	})
}

func inRoot(dir, name string, fn func(r *os.Root) (string, error)) (string, error) {
	if name == "" {
		return "", fmt.Errorf("%w: empty name", ErrBadInput)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return "", err
	}
	defer root.Close()

	return fn(root)
}

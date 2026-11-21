package v1

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type CreateFileArgs struct {
	Name    string `json:"name" jsonschema:"file name, e.g. part1.txt"`
	Content string `json:"content" jsonschema:"full file content"`
}

type ListFilesArgs struct{}

type fileTools struct {
	root string
}

func NewServer(root string) *mcp.Server {
	ft := &fileTools{root: root}
	s := mcp.NewServer(&mcp.Implementation{Name: ServerName, Version: ServerVersion}, nil)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "createFile",
		Description: "Creates a file with the given name and content in the output directory",
	}, ft.createFile)
	mcp.AddTool(s, &mcp.Tool{
		Name:        "listFiles",
		Description: "Lists the files in the output directory",
	}, ft.listFiles)
	return s
}

func (f *fileTools) createFile(ctx context.Context, _ *mcp.CallToolRequest, in CreateFileArgs) (*mcp.CallToolResult, any, error) {
	name := filepath.Base(strings.TrimSpace(in.Name))
	if name == "" || name == "." || name == ".." || name == string(filepath.Separator) {
		return nil, nil, fmt.Errorf("%w: %q", ErrBadFileName, in.Name)
	}
	if err := os.MkdirAll(f.root, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create root: %w", err)
	}
	path := filepath.Join(f.root, name)
	if err := os.WriteFile(path, []byte(in.Content), 0o644); err != nil {
		return nil, nil, fmt.Errorf("write %s: %w", name, err)
	}
	return textResult(fmt.Sprintf("created %s (%d bytes)", name, len(in.Content))), nil, nil
}

func (f *fileTools) listFiles(ctx context.Context, _ *mcp.CallToolRequest, _ ListFilesArgs) (*mcp.CallToolResult, any, error) {
	entries, err := os.ReadDir(f.root)
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("read root: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return textResult("no files"), nil, nil
	}
	return textResult(strings.Join(names, "\n")), nil, nil
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}
}

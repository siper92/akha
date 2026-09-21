package fs

import "context"

const ModuleName = "FS"

type Sandbox interface {
	Root() string
	Resolve(path string) (string, error)
}

type FS interface {
	ReadFile(ctx context.Context, path string) (string, error)
	WriteFile(ctx context.Context, path, content string) error
	UpdateFile(ctx context.Context, path, content string) error
	ListFiles(ctx context.Context, dir string) ([]string, error)
}

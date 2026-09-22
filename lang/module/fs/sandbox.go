package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrEscape   = errors.New("path escapes root")
	ErrNotExist = errors.New("file does not exist")
)

type sandbox struct {
	root string
}

var _ Sandbox = (*sandbox)(nil)

func NewSandbox(root string) (Sandbox, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &sandbox{root: abs}, nil
}

func (s *sandbox) Root() string { return s.root }

func (s *sandbox) Resolve(path string) (string, error) {
	full := filepath.Join(s.root, path)
	rel, err := filepath.Rel(s.root, full)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrEscape, path)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: %s", ErrEscape, path)
	}
	return full, nil
}

type fileSystem struct {
	sb Sandbox
}

var _ FS = (*fileSystem)(nil)

func New(sb Sandbox) FS {
	return &fileSystem{sb: sb}
}

func (f *fileSystem) ReadFile(ctx context.Context, path string) (string, error) {
	full, err := f.sb.Resolve(path)
	if err != nil {
		return "", err
	}
	b, err := os.ReadFile(full)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("%w: %s", ErrNotExist, path)
	}
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (f *fileSystem) WriteFile(ctx context.Context, path, content string) error {
	full, err := f.sb.Resolve(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

func (f *fileSystem) UpdateFile(ctx context.Context, path, content string) error {
	full, err := f.sb.Resolve(path)
	if err != nil {
		return err
	}
	if _, err := os.Stat(full); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrNotExist, path)
	} else if err != nil {
		return err
	}
	return os.WriteFile(full, []byte(content), 0o644)
}

func (f *fileSystem) ListFiles(ctx context.Context, dir string) ([]string, error) {
	full, err := f.sb.Resolve(dir)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(full)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotExist, dir)
	}
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			name += "/"
		}
		out = append(out, name)
	}
	return out, nil
}

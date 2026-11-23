package v3

import (
	"context"
	"fmt"
	"os"
)

var (
	_ FS = (*MemFS)(nil)
	_ FS = (*DirFS)(nil)
)

type MemFS struct {
	files map[string]string
}

func NewMemFS() *MemFS {
	return &MemFS{files: map[string]string{}}
}

func (m *MemFS) CreateFile(_ context.Context, fileName string, content string) error {
	m.files[fileName] = content
	return nil
}

func (m *MemFS) ReadFile(_ context.Context, fileName string) (string, error) {
	content, ok := m.files[fileName]
	if !ok {
		return "", fmt.Errorf("%w: %s", os.ErrNotExist, fileName)
	}

	return content, nil
}

func (m *MemFS) DeleteFile(_ context.Context, fileName string) error {
	if _, ok := m.files[fileName]; !ok {
		return fmt.Errorf("%w: %s", os.ErrNotExist, fileName)
	}
	delete(m.files, fileName)

	return nil
}

type DirFS struct {
	root *os.Root
}

func OpenDirFS(dir string) (*DirFS, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}

	return &DirFS{root: root}, nil
}

func (d *DirFS) Close() error { return d.root.Close() }

func (d *DirFS) CreateFile(_ context.Context, fileName string, content string) error {
	return d.root.WriteFile(fileName, []byte(content), 0o644)
}

func (d *DirFS) ReadFile(_ context.Context, fileName string) (string, error) {
	b, err := d.root.ReadFile(fileName)
	return string(b), err
}

func (d *DirFS) DeleteFile(_ context.Context, fileName string) error {
	return d.root.Remove(fileName)
}

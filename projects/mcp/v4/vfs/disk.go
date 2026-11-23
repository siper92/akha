package vfs

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

type Disk struct {
	root *os.Root
}

var _ FS = (*Disk)(nil)

func NewDisk(dir string) (*Disk, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("vfs: create %s: %w", dir, err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("vfs: open %s: %w", dir, err)
	}
	return &Disk{root: root}, nil
}

func (d *Disk) WriteFile(_ context.Context, name string, content string) error {
	return d.root.WriteFile(name, []byte(content), 0o644)
}

func (d *Disk) ReadFile(_ context.Context, name string) (string, error) {
	content, err := d.root.ReadFile(name)
	if err != nil {
		return "", notFound(name, err)
	}
	return string(content), nil
}

func (d *Disk) DeleteFile(_ context.Context, name string) error {
	return notFound(name, d.root.Remove(name))
}

func notFound(name string, err error) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return err
}

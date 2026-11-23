package vfs

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("vfs: file not found")

type FS interface {
	WriteFile(ctx context.Context, name string, content string) error
	ReadFile(ctx context.Context, name string) (string, error)
	DeleteFile(ctx context.Context, name string) error
}

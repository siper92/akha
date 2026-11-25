package vfs

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("file service: file not found")

type FSService interface {
	WriteFile(ctx context.Context, name string, content string) error
	ReadFile(ctx context.Context, name string) (string, error)
	DeleteFile(ctx context.Context, name string) error
}

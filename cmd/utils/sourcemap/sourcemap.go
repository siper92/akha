package sourcemap

import (
	"context"
	"errors"
)

var (
	ErrNotDir   = errors.New("sourcemap: path is not a directory")
	ErrNoModule = errors.New("sourcemap: go.mod not found")
)

type File struct {
	Module  string
	Path    string
	Imports []string
	Types   []string
	Exports []string
	Private []string
}

type Options struct {
	WithPrivate bool
}

type Parser interface {
	Parse(ctx context.Context, root string, opts Options) ([]*File, error)
}

type Encoder interface {
	Encode(ctx context.Context, files []*File) ([]byte, error)
}

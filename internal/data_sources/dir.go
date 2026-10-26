package data_sources

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	ErrNoHandler = errors.New("no handler for extension")
	ErrNotDir    = errors.New("not a directory")
	ErrFileType  = errors.New("file data type mismatch")
)

type File interface {
	Path() string
	Ext() string
}

type FileDataSource[T any] interface {
	File
	Init(ext string) error
	Process(ctx context.Context) (T, error)
}

type DirDataSource interface {
	Path() string
	Files(ctx context.Context) ([]File, error)
}

type FileHandler func(path string) File

type Handlers map[string]FileHandler

func As[T any](f File) (FileDataSource[T], error) {
	s, ok := f.(FileDataSource[T])
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFileType, f.Path())
	}
	if err := s.Init(f.Ext()); err != nil {
		return nil, err
	}
	return s, nil
}

func Read[T any](ctx context.Context, f File) (T, error) {
	s, err := As[T](f)
	if err != nil {
		var zero T
		return zero, err
	}
	return s.Process(ctx)
}

var _ DirDataSource = (*Folder)(nil)

type Folder struct {
	path     string
	handlers Handlers
}

func NewFolder(path string, handlers Handlers) (*Folder, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrNotDir, path)
	}
	return &Folder{path: path, handlers: handlers}, nil
}

func (d *Folder) Path() string { return d.path }

func (d *Folder) Files(ctx context.Context) ([]File, error) {
	entries, err := os.ReadDir(d.path)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.TrimPrefix(filepath.Ext(e.Name()), ".")
		h, ok := d.handlers[ext]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrNoHandler, e.Name())
		}
		files = append(files, h(filepath.Join(d.path, e.Name())))
	}
	slices.SortFunc(files, func(a, b File) int { return strings.Compare(a.Path(), b.Path()) })
	return files, nil
}

type FolderProvider struct {
	root     string
	handlers Handlers
}

func NewFolderProvider(root string, handlers Handlers) *FolderProvider {
	return &FolderProvider{root: root, handlers: handlers}
}

func (p *FolderProvider) Dir(_ context.Context, path string) (DirDataSource, error) {
	return NewFolder(filepath.Join(p.root, path), p.handlers)
}

package data_sources

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
)

type Kind string

const (
	KindDB    Kind = "db"
	KindFS    Kind = "fs"
	KindAPI   Kind = "api"
	KindOther Kind = "other"
)

var (
	ErrSourceNotFound  = errors.New("source not found")
	ErrDuplicateSource = errors.New("duplicate source")
	ErrSourceType      = errors.New("source type mismatch")
)

type Query struct {
	Target string
	Params map[string]any
}

type Descriptor interface {
	Name() string
	Kind() Kind
}

type Source[T any] interface {
	Descriptor
	Fetch(ctx context.Context, q Query) (T, error)
}

type Provider interface {
	Register(src Descriptor) error
	Lookup(name string) (Descriptor, error)
	Names() []string
}

func Get[T any](p Provider, name string) (Source[T], error) {
	d, err := p.Lookup(name)
	if err != nil {
		return nil, err
	}
	s, ok := d.(Source[T])
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSourceType, name)
	}
	return s, nil
}

func Fetch[T any](ctx context.Context, p Provider, name string, q Query) (T, error) {
	s, err := Get[T](p, name)
	if err != nil {
		var zero T
		return zero, err
	}
	return s.Fetch(ctx, q)
}

var _ Source[struct{}] = (*FuncSource[struct{}])(nil)

type FuncSource[T any] struct {
	name string
	kind Kind
	fn   func(ctx context.Context, q Query) (T, error)
}

func NewSource[T any](name string, kind Kind, fn func(ctx context.Context, q Query) (T, error)) *FuncSource[T] {
	return &FuncSource[T]{name: name, kind: kind, fn: fn}
}

func (s *FuncSource[T]) Name() string { return s.name }

func (s *FuncSource[T]) Kind() Kind { return s.kind }

func (s *FuncSource[T]) Fetch(ctx context.Context, q Query) (T, error) {
	return s.fn(ctx, q)
}

var _ Provider = (*Registry)(nil)

type Registry struct {
	mu      sync.RWMutex
	sources map[string]Descriptor
}

func NewRegistry() *Registry {
	return &Registry{sources: map[string]Descriptor{}}
}

func (r *Registry) Register(src Descriptor) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.sources[src.Name()]; ok {
		return fmt.Errorf("%w: %s", ErrDuplicateSource, src.Name())
	}
	r.sources[src.Name()] = src
	return nil
}

func (r *Registry) Lookup(name string) (Descriptor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sources[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrSourceNotFound, name)
	}
	return s, nil
}

func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.sources))
	for n := range r.sources {
		names = append(names, n)
	}
	slices.Sort(names)
	return names
}

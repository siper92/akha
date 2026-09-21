package eval

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"
)

var (
	ErrDuplicateModule = errors.New("duplicate module")
	ErrNilModule       = errors.New("nil module")
)

type registry struct {
	mu   sync.RWMutex
	mods map[string]Module
}

var _ Registry = (*registry)(nil)

func NewRegistry() Registry {
	return &registry{mods: map[string]Module{}}
}

func (r *registry) Register(m Module) error {
	if m == nil {
		return ErrNilModule
	}
	name := m.Name()
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.mods[name]; ok {
		return fmt.Errorf("%w: %s", ErrDuplicateModule, name)
	}
	r.mods[name] = m
	return nil
}

func (r *registry) Lookup(name string) (Module, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.mods[name]
	return m, ok
}

func (r *registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Sorted(maps.Keys(r.mods))
}

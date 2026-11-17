package pages

import (
	"context"
	"fmt"
)

type memStore struct {
	byPath  map[string]Page
	ordered []Page
	nav     []NavItem
}

var _ Store = (*memStore)(nil)

func Build(ctx context.Context, src Source, r Renderer) (Store, error) {
	docs, err := src.List(ctx)
	if err != nil {
		return nil, err
	}
	s := &memStore{
		byPath:  make(map[string]Page, len(docs)),
		ordered: make([]Page, 0, len(docs)),
		nav:     NavFrom(docs),
	}
	for _, d := range docs {
		p, err := r.Render(ctx, d)
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", d.Slug, err)
		}
		s.byPath[p.Path] = p
		s.ordered = append(s.ordered, p)
	}
	return s, nil
}

func (s *memStore) Get(path string) (Page, bool) {
	p, ok := s.byPath[path]
	return p, ok
}

func (s *memStore) Pages() []Page {
	out := make([]Page, len(s.ordered))
	copy(out, s.ordered)
	return out
}

func (s *memStore) Nav() []NavItem {
	out := make([]NavItem, len(s.nav))
	copy(out, s.nav)
	return out
}

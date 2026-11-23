package v3

import "sync"

var _ Store[Data] = (*MemStore[Data])(nil)

type MemStore[D any] struct {
	mu sync.RWMutex
	d  D
}

func NewStore[D any](d D) *MemStore[D] {
	return &MemStore[D]{d: d}
}

func (s *MemStore[D]) Data() D {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.d
}

func (s *MemStore[D]) Update(fn func(d D) D) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.d = fn(s.d)
}

package vfs

import (
	"context"
	"fmt"
	"sync"
)

type Memory struct {
	mu    sync.RWMutex
	files map[string]string
}

var _ FSService = (*Memory)(nil)

func NewMemory() *Memory {
	return &Memory{files: map[string]string{}}
}

func (m *Memory) WriteFile(_ context.Context, name string, content string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[name] = content
	return nil
}

func (m *Memory) ReadFile(_ context.Context, name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	content, ok := m.files[name]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return content, nil
}

func (m *Memory) DeleteFile(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.files[name]; !ok {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	delete(m.files, name)
	return nil
}

package flow

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
)

var ErrTaskNotFound = errors.New("flow: task not found")

type Status string

const (
	StatusPending Status = "pending"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type TaskRecord struct {
	ID          string
	Title       string
	Description string
	Status      Status
	Result      string
}

type TaskRepository interface {
	NextPending(ctx context.Context) (TaskRecord, bool, error)
	Save(ctx context.Context, record TaskRecord) error
	List(ctx context.Context) ([]TaskRecord, error)
}

type MemoryRepository struct {
	mu      sync.RWMutex
	records []TaskRecord
}

var _ TaskRepository = (*MemoryRepository)(nil)

func NewMemoryRepository(records ...TaskRecord) *MemoryRepository {
	stored := slices.Clone(records)
	for i := range stored {
		if stored[i].Status == "" {
			stored[i].Status = StatusPending
		}
	}
	return &MemoryRepository{records: stored}
}

func (r *MemoryRepository) NextPending(_ context.Context) (TaskRecord, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, record := range r.records {
		if record.Status == StatusPending {
			return record, true, nil
		}
	}
	return TaskRecord{}, false, nil
}

func (r *MemoryRepository) Save(_ context.Context, record TaskRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.records {
		if r.records[i].ID == record.ID {
			r.records[i] = record
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrTaskNotFound, record.ID)
}

func (r *MemoryRepository) List(_ context.Context) ([]TaskRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return slices.Clone(r.records), nil
}

package v3

import (
	"context"
	"slices"
)

var _ TaskRepository = (*MemTaskRepository)(nil)

type MemTaskRepository struct {
	tasks []TaskRecord
}

func NewTaskRepository(tasks ...TaskRecord) *MemTaskRepository {
	return &MemTaskRepository{tasks: tasks}
}

func (r *MemTaskRepository) List(_ context.Context) ([]TaskRecord, error) {
	return slices.Clone(r.tasks), nil
}

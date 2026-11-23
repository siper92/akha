package v3

import (
	"context"
	"fmt"
	"slices"
)

var (
	_ Step[Data] = LoadTasks{}
	_ Step[Data] = RunTasks{}
)

type LoadTasks struct{}

func (LoadTasks) Name() string { return "load_tasks" }

func (LoadTasks) Exec(ctx context.Context, sp FlowServices, store Store[Data]) error {
	tasks, err := sp.Tasks().List(ctx)
	if err != nil {
		return err
	}

	store.Update(func(d Data) Data {
		d.Tasks = tasks
		return d
	})
	sp.DebugService().Log("loaded %d tasks", len(tasks))

	return nil
}

type RunTasks struct {
	Task Task
}

func (s RunTasks) Name() string { return "run_tasks" }

func (s RunTasks) Exec(ctx context.Context, sp FlowServices, store Store[Data]) error {
	task := s.Task
	if task == nil {
		task = AITask
	}

	for _, record := range store.Data().Tasks {
		out, err := task(ctx, sp.TaskServices(), TaskIn{Record: record, Done: store.Data().Results})
		if err != nil {
			return fmt.Errorf("task %s: %w", record.ID, err)
		}

		store.Update(func(d Data) Data {
			d.Results = append(slices.Clone(d.Results), out)
			return d
		})
	}

	return nil
}

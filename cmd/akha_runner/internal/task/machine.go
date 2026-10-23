package task

import (
	"context"
	"errors"

	"github.com/siper92/akha/cmd/akha_runner/internal/types"
	"github.com/siper92/akha/internal/flow"
	"github.com/siper92/akha/internal/state_machine"
)

var ErrEmptyContent = errors.New("task content is empty")

type Task struct {
	UID           string
	ProjectID     int64
	Status        types.TaskStatus
	ExecutionPlan flow.ExecutionPlan // code of the execution plan
	WorkerID      int64
	Content       string
	Output        string
	Error         string
}

type Processor interface {
	Load(ctx context.Context, t *Task) error
	Clarify(ctx context.Context, t *Task) (clear bool, err error)
	Plan(ctx context.Context, t *Task) error
	Run(ctx context.Context, t *Task) (subs bool, err error)
	RunSub(ctx context.Context, t *Task) error
}

type Machine = state_machine.Machine[Task, types.TaskStatus]

type step = state_machine.StepFunc[Task, types.TaskStatus]

func BuildNewMachine(p Processor) *Machine {
	return state_machine.NewMachine[Task, types.TaskStatus](types.Loading).
		Step(types.Loading, track(load(p)), types.Clarifying).
		Step(types.Clarifying, track(clarify(p)), types.Clarifying, types.Planning).
		Step(types.Planning, track(plan(p)), types.Running).
		Step(types.Running, track(run(p)), types.RunningSub, types.Done).
		Step(types.RunningSub, track(runSub(p)), types.Running).
		Final(types.Done, types.Failed)
}

func Process(ctx context.Context, m *Machine, t *Task) error {
	from := t.Status
	if from == "" {
		from = types.Loading
	}

	if _, err := m.RunFrom(ctx, from, t); err != nil {
		t.Status = types.Failed
		t.Error = err.Error()
		return err
	}
	return nil
}

func track(fn step) step {
	return func(ctx context.Context, t *Task) (types.TaskStatus, error) {
		next, err := fn(ctx, t)
		if err == nil {
			t.Status = next
		}
		return next, err
	}
}

func load(p Processor) step {
	return func(ctx context.Context, t *Task) (types.TaskStatus, error) {
		if t.Content == "" {
			return types.Failed, ErrEmptyContent
		}
		if err := p.Load(ctx, t); err != nil {
			return types.Failed, err
		}
		return types.Clarifying, nil
	}
}

func clarify(p Processor) step {
	return func(ctx context.Context, t *Task) (types.TaskStatus, error) {
		ok, err := p.Clarify(ctx, t)
		if err != nil {
			return types.Failed, err
		}
		if !ok {
			return types.Clarifying, nil
		}
		return types.Planning, nil
	}
}

func plan(p Processor) step {
	return func(ctx context.Context, t *Task) (types.TaskStatus, error) {
		if err := p.Plan(ctx, t); err != nil {
			return types.Failed, err
		}
		return types.Running, nil
	}
}

func run(p Processor) step {
	return func(ctx context.Context, t *Task) (types.TaskStatus, error) {
		subs, err := p.Run(ctx, t)
		if err != nil {
			return types.Failed, err
		}
		if subs {
			return types.RunningSub, nil
		}
		return types.Done, nil
	}
}

func runSub(p Processor) step {
	return func(ctx context.Context, t *Task) (types.TaskStatus, error) {
		if err := p.RunSub(ctx, t); err != nil {
			return types.Failed, err
		}
		return types.Running, nil
	}
}

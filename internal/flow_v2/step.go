package flow_v2

import (
	"context"
	"fmt"
)

type Loader[T any] func(ctx context.Context, rt Runtime) (T, error)

type Decider[R any] func(results []R) Outcome

var _ TaskStep[struct{}, struct{}] = (*BaseStep[struct{}, struct{}])(nil)

type BaseStep[T, R any] struct {
	id     string
	load   Loader[T]
	decide Decider[R]
	tasks  []Task[T, R]
	plan   ExecutionPlan
}

func NewStep[T, R any](id string, load Loader[T], decide Decider[R]) *BaseStep[T, R] {
	return &BaseStep[T, R]{id: id, load: load, decide: decide}
}

func (s *BaseStep[T, R]) ID() string { return s.id }

func (s *BaseStep[T, R]) Init(tasks []Task[T, R], plan ExecutionPlan) error {
	s.tasks, s.plan = tasks, plan
	return nil
}

func (s *BaseStep[T, R]) Data(ctx context.Context, rt Runtime) (T, error) {
	return s.load(ctx, rt)
}

func (s *BaseStep[T, R]) Execute(ctx context.Context, rt Runtime) (Outcome, error) {
	if s.plan == nil {
		return "", fmt.Errorf("%w: step %s", ErrNotInit, s.id)
	}
	data, err := s.Data(ctx, rt)
	if err != nil {
		return "", fmt.Errorf("step %s: %w", s.id, err)
	}
	results := make([]R, 0, len(s.tasks))
	for _, t := range s.tasks {
		t.Init(data)
		if err := t.Validate(); err != nil {
			return "", fmt.Errorf("task %s: %w", t.ID(), err)
		}
		r, err := t.Process(ctx)
		if err != nil {
			return "", fmt.Errorf("task %s: %w", t.ID(), err)
		}
		results = append(results, r)
	}
	rt.State().Set(s.id, results)
	if s.decide == nil {
		return OutcomeOK, nil
	}
	return s.decide(results), nil
}

func Results[R any](st State, id string) ([]R, error) {
	v, ok := st.Get(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoResult, id)
	}
	r, ok := v.([]R)
	if !ok {
		return nil, fmt.Errorf("%w: %s type %T", ErrNoResult, id, v)
	}
	return r, nil
}

type FuncTask[T, R any] struct {
	id       string
	data     T
	validate func(T) error
	process  func(context.Context, T) (R, error)
}

var _ Task[struct{}, struct{}] = (*FuncTask[struct{}, struct{}])(nil)

func NewTask[T, R any](id string, validate func(T) error, process func(context.Context, T) (R, error)) *FuncTask[T, R] {
	return &FuncTask[T, R]{id: id, validate: validate, process: process}
}

func (t *FuncTask[T, R]) ID() string { return t.id }

func (t *FuncTask[T, R]) Init(data T) { t.data = data }

func (t *FuncTask[T, R]) Validate() error {
	if t.validate == nil {
		return nil
	}
	return t.validate(t.data)
}

func (t *FuncTask[T, R]) Process(ctx context.Context) (R, error) {
	return t.process(ctx, t.data)
}

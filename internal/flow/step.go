package flow

import (
	"context"
	"fmt"

	"github.com/siper92/akha/internal/data_sources"
)

type DataGetterFunc[D any] func(ctx context.Context, scope StepScope) ([]D, error)

func (f DataGetterFunc[D]) Get(ctx context.Context, scope StepScope) ([]D, error) {
	return f(ctx, scope)
}

var _ DataGetter[struct{}] = sourceGetter[struct{}]{}

type sourceGetter[D any] struct {
	name string
	q    data_sources.Query
}

func FromSource[D any](name string, q data_sources.Query) DataGetter[D] {
	return sourceGetter[D]{name: name, q: q}
}

func (g sourceGetter[D]) Get(ctx context.Context, scope StepScope) ([]D, error) {
	return data_sources.Fetch[[]D](ctx, scope.Sources(), g.name, g.q)
}

var _ Step = (*TaskStep[struct{}, struct{}])(nil)

type TaskStep[D, R any] struct {
	id      ID
	getter  DataGetter[D]
	plan    Plan
	tasks   map[ID]Task[D, R]
	proc    TaskProcessor[D, R]
	workers int
}

func NewStep[D, R any](id ID, getter DataGetter[D], tasks ...Task[D, R]) (*TaskStep[D, R], error) {
	ids := make([]ID, 0, len(tasks))
	byID := make(map[ID]Task[D, R], len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.ID())
		byID[t.ID()] = t
	}
	plan := NewPlan(ids...)
	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("step %s: %w", id, err)
	}
	return &TaskStep[D, R]{
		id:      id,
		getter:  getter,
		plan:    plan,
		tasks:   byID,
		proc:    ProcessorFunc[TaskRun[D, R], R](processTask[D, R]),
		workers: 1,
	}, nil
}

func (s *TaskStep[D, R]) Workers(n int) *TaskStep[D, R] {
	if n > 0 {
		s.workers = n
	}
	return s
}

func (s *TaskStep[D, R]) Use(p TaskProcessor[D, R]) *TaskStep[D, R] {
	s.proc = p
	return s
}

func (s *TaskStep[D, R]) ID() ID { return s.id }

func (s *TaskStep[D, R]) Process(ctx context.Context, scope StepScope) (any, error) {
	items, err := s.getter.Get(ctx, scope)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrDataGet, s.id, err)
	}
	return ProcessAll[D, R](ctx, ProcessorFunc[D, R](s.chain), items, s.workers)
}

func (s *TaskStep[D, R]) chain(ctx context.Context, data D) (R, error) {
	sc := &taskScope[D, R]{step: s, data: data, results: map[ID]R{}, cur: -1}
	var out R
	for i, id := range s.plan.Order() {
		r, err := sc.run(ctx, i, id)
		if err != nil {
			return out, err
		}
		out = r
	}
	return out, nil
}

var _ Scope[struct{}] = (*taskScope[struct{}, struct{}])(nil)

type taskScope[D, R any] struct {
	step    *TaskStep[D, R]
	data    D
	results map[ID]R
	cur     int
}

func (sc *taskScope[D, R]) Result(id ID) (R, bool) {
	r, ok := sc.results[id]
	return r, ok
}

func (sc *taskScope[D, R]) Call(ctx context.Context, id ID) (R, error) {
	i, ok := sc.step.plan.Index(id)
	if !ok {
		var zero R
		return zero, fmt.Errorf("%w: %s", ErrUnknownTask, id)
	}
	if i >= sc.cur {
		var zero R
		return zero, fmt.Errorf("%w: %s", ErrNotPrevious, id)
	}
	return sc.run(ctx, i, id)
}

func (sc *taskScope[D, R]) run(ctx context.Context, i int, id ID) (R, error) {
	prev := sc.cur
	sc.cur = i
	defer func() { sc.cur = prev }()

	r, err := sc.step.proc.Process(ctx, TaskRun[D, R]{Task: sc.step.tasks[id], Data: sc.data, Scope: sc})
	if err != nil {
		return r, fmt.Errorf("%w: %s: %w", ErrTaskFailed, id, err)
	}
	sc.results[id] = r
	return r, nil
}

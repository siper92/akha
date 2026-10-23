package state_machine

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type MachineState string

const (
	Done   MachineState = "done"
	Failed MachineState = "failed"
	Error  MachineState = "error"
)

var (
	ErrUnknownState      = errors.New("unknown state")
	ErrDuplicateStep     = errors.New("duplicate step")
	ErrNilStep           = errors.New("nil step func")
	ErrInvalidTransition = errors.New("invalid transition")
	ErrStepFailed        = errors.New("step failed")
	ErrStepPanic         = errors.New("step panic")
	ErrMaxSteps          = errors.New("max steps reached")
)

type StepFunc[T any, S comparable] func(ctx context.Context, data *T) (S, error)

type Transition[S comparable] struct {
	From, To S
	Took     time.Duration
	Err      error
}

type Runner[T any, S comparable] interface {
	Validate() error
	Run(ctx context.Context, data *T) (S, error)
	RunFrom(ctx context.Context, from S, data *T) (S, error)
}

var _ Runner[struct{}, string] = (*Machine[struct{}, string])(nil)

type step[T any, S comparable] struct {
	fn StepFunc[T, S]
	to map[S]struct{}
}

type Machine[T any, S comparable] struct {
	steps    map[S]step[T, S]
	final    map[S]struct{}
	start    S
	maxSteps int
	errs     []error
	OnStep   func(t Transition[S])
}

func NewMachine[T any, S comparable](start S) *Machine[T, S] {
	return &Machine[T, S]{
		steps:    map[S]step[T, S]{},
		final:    map[S]struct{}{},
		start:    start,
		maxSteps: 100,
	}
}

func (m *Machine[T, S]) Step(id S, fn StepFunc[T, S], to ...S) *Machine[T, S] {
	if _, ok := m.steps[id]; ok {
		m.errs = append(m.errs, fmt.Errorf("%w: %v", ErrDuplicateStep, id))
		return m
	}
	if fn == nil {
		m.errs = append(m.errs, fmt.Errorf("%w: %v", ErrNilStep, id))
		return m
	}

	next := make(map[S]struct{}, len(to))
	for _, s := range to {
		next[s] = struct{}{}
	}
	m.steps[id] = step[T, S]{fn: fn, to: next}
	return m
}

func (m *Machine[T, S]) Final(states ...S) *Machine[T, S] {
	for _, s := range states {
		m.final[s] = struct{}{}
	}
	return m
}

func (m *Machine[T, S]) MaxSteps(n int) *Machine[T, S] {
	if n > 0 {
		m.maxSteps = n
	}
	return m
}

func (m *Machine[T, S]) Validate() error {
	errs := append([]error{}, m.errs...)
	if !m.known(m.start) {
		errs = append(errs, fmt.Errorf("%w: start %v", ErrUnknownState, m.start))
	}
	for id, s := range m.steps {
		for to := range s.to {
			if !m.known(to) {
				errs = append(errs, fmt.Errorf("%w: %v -> %v", ErrUnknownState, id, to))
			}
		}
	}
	return errors.Join(errs...)
}

func (m *Machine[T, S]) Run(ctx context.Context, data *T) (S, error) {
	return m.RunFrom(ctx, m.start, data)
}

func (m *Machine[T, S]) RunFrom(ctx context.Context, from S, data *T) (S, error) {
	if err := m.Validate(); err != nil {
		return from, err
	}
	if !m.known(from) {
		return from, fmt.Errorf("%w: %v", ErrUnknownState, from)
	}

	cur := from
	for i := 0; ; i++ {
		if _, ok := m.final[cur]; ok {
			return cur, nil
		}
		if i >= m.maxSteps {
			return cur, fmt.Errorf("%w: %d at %v", ErrMaxSteps, m.maxSteps, cur)
		}
		if err := ctx.Err(); err != nil {
			return cur, err
		}

		next, err := m.exec(ctx, cur, data)
		if err != nil {
			return cur, err
		}
		cur = next
	}
}

func (m *Machine[T, S]) exec(ctx context.Context, cur S, data *T) (S, error) {
	s := m.steps[cur]
	began := time.Now()

	next, err := call(ctx, s.fn, data)
	if err != nil {
		err = fmt.Errorf("%w: %v: %w", ErrStepFailed, cur, err)
	} else if _, ok := s.to[next]; !ok {
		err = fmt.Errorf("%w: %v -> %v", ErrInvalidTransition, cur, next)
	}

	if m.OnStep != nil {
		m.OnStep(Transition[S]{From: cur, To: next, Took: time.Since(began), Err: err})
	}
	return next, err
}

func (m *Machine[T, S]) known(s S) bool {
	if _, ok := m.steps[s]; ok {
		return true
	}
	_, ok := m.final[s]
	return ok
}

func call[T any, S comparable](ctx context.Context, fn StepFunc[T, S], data *T) (next S, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w: %v", ErrStepPanic, r)
		}
	}()
	return fn(ctx, data)
}

package flow

import (
	"context"
	"errors"
	"fmt"

	"aiflow/debug"
)

var (
	ErrNoSteps       = errors.New("flow: no steps")
	ErrDuplicateStep = errors.New("flow: duplicate step")
	ErrUnknownStep   = errors.New("flow: unknown step")
)

type Transition struct {
	next string
	end  bool
}

func Next(step string) Transition { return Transition{next: step} }

func End() Transition { return Transition{end: true} }

type Step[SP ServiceProvider, T any] interface {
	Name() string
	Exec(ctx context.Context, sp SP, data T) (Transition, error)
}

type Runner[T any] interface {
	Run(ctx context.Context, data T) error
}

type Flow[SP ServiceProvider, T any] struct {
	sp    SP
	start string
	steps map[string]Step[SP, T]
}

var _ Runner[*State] = (*Flow[ServiceProvider, *State])(nil)

func New[SP ServiceProvider, T any](sp SP, steps ...Step[SP, T]) (*Flow[SP, T], error) {
	if len(steps) == 0 {
		return nil, ErrNoSteps
	}
	byName := make(map[string]Step[SP, T], len(steps))
	for _, step := range steps {
		if _, ok := byName[step.Name()]; ok {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateStep, step.Name())
		}
		byName[step.Name()] = step
	}
	return &Flow[SP, T]{sp: sp, start: steps[0].Name(), steps: byName}, nil
}

func (f *Flow[SP, T]) Run(ctx context.Context, data T) error {
	name := f.start
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		step, ok := f.steps[name]
		if !ok {
			return fmt.Errorf("%w: %s", ErrUnknownStep, name)
		}
		debug.Service().Log("flow: step %s", name)
		transition, err := step.Exec(ctx, f.sp, data)
		if err != nil {
			return fmt.Errorf("flow: step %s: %w", name, err)
		}
		if transition.end {
			return nil
		}
		name = transition.next
	}
}

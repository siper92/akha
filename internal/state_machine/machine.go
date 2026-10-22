package state_machine

import (
	"context"
	"time"
)

type MachineState string

const (
	Done   MachineState = "done"
	Failed MachineState = "failed"
	Error  MachineState = "error"
)

type StepFunc[T any, S comparable] func(ctx context.Context, data *T) (S, error)

type Transition[S comparable] struct {
	From, To S
	Took     time.Duration
	Err      error
}

type Machine[T any, S comparable] struct {
	steps    map[S]StepFunc[T, S]
	start    S
	maxSteps int
	OnStep   func(t Transition[S])
}

func NewMachine[T any, S comparable](start S) *Machine[T, S] {
	return &Machine[T, S]{steps: map[S]StepFunc[T, S]{}, start: start, maxSteps: 100}
}

func (m *Machine[T, S]) Step(id S, fn StepFunc[T, S]) *Machine[T, S] {
	m.steps[id] = fn
	return m
}

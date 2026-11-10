package flow_v3

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	DoneTransition    Transition = "done"
	ErrorTransition   Transition = "error"
	FinishTransition  Transition = "finish"
	UnknownTransition Transition = "unknown"
)

const FinishStep StepName = "finish"

var (
	ErrNoSteps           = errors.New("no steps provided")
	ErrStepExists        = errors.New("step already exists")
	ErrStepNotFound      = errors.New("step not found")
	ErrNoStepExec        = errors.New("no exec for step")
	ErrTransitionMissing = errors.New("transition not found")
	ErrUnknownTransition = errors.New("unknown transition")
	ErrNoActiveStep      = errors.New("no active step")
	ErrRetriesExhausted  = errors.New("retries exhausted")
	ErrInvalidGraph      = errors.New("invalid graph")
)

var _ DataServiceProvide[any] = (*StepData[any])(nil)

type StepData[D any] struct {
	mu sync.RWMutex
	d  D
}

func NewStepData[D any](d D) *StepData[D] {
	return &StepData[D]{d: d}
}

func (s *StepData[D]) ServiceProvide() ServiceProvide {
	return s
}

func (s *StepData[D]) Data() D {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.d
}

func (s *StepData[D]) SetData(d D) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.d = d
}

func (s *StepData[D]) Update(fn func(d D) D) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.d = fn(s.d)
}

var _ Step[DataServiceProvide[any]] = (*MachineStep[any])(nil)

type MachineStep[D any] struct {
	name        StepName
	transitions map[Transition]nexStep
	exec        StepExec[D]
}

func (m *MachineStep[D]) process(ctx context.Context, s DataServiceProvide[D]) (Transition, error) {
	if m.exec == nil {
		return UnknownTransition, fmt.Errorf("%w: %s", ErrNoStepExec, m.name)
	}

	return m.exec(ctx, s)
}

func (m *MachineStep[D]) Name() StepName {
	return m.name
}

func (m *MachineStep[D]) Retry() int {
	if e, ok := m.transitions[ErrorTransition]; ok && e.Next == m.name {
		return e.Retry
	}

	return 0
}

func (m *MachineStep[D]) On(t Transition) (StepName, error) {
	if e, ok := m.transitions[t]; ok {
		return e.Next, nil
	}

	return "", fmt.Errorf("%w: %s on %s", ErrTransitionMissing, t, m.name)
}

type nexStep struct {
	Next  StepName
	Retry int
}

func (n *nexStep) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		n.Next = StepName(node.Value)
		n.Retry = 1

		return nil
	case yaml.MappingNode:
		var raw struct {
			Next  string `yaml:"next"`
			Retry *int   `yaml:"retry"`
		}
		if err := node.Decode(&raw); err != nil {
			return err
		}
		if raw.Next == "" {
			return fmt.Errorf("line %d: transition is missing 'next'", node.Line)
		}

		n.Next = StepName(raw.Next)
		n.Retry = 1

		if raw.Retry != nil {
			n.Retry = *raw.Retry
		}

		return nil
	default:
		return fmt.Errorf("line %d: transition must be a step name or {next, retry}", node.Line)
	}
}

var _ Flow[DataServiceProvide[any]] = (*MachineFlow[any])(nil)

type MachineFlow[D any] struct {
	start   Step[DataServiceProvide[D]]
	active  Step[DataServiceProvide[D]]
	steps   map[StepName]Step[DataServiceProvide[D]]
	attempt int
}

func (m *MachineFlow[D]) Init(steps ...Step[DataServiceProvide[D]]) error {
	if len(steps) == 0 {
		return ErrNoSteps
	}

	if m.steps == nil {
		m.steps = make(map[StepName]Step[DataServiceProvide[D]], len(steps))
	}

	for _, step := range steps {
		if _, ok := m.steps[step.Name()]; ok {
			return fmt.Errorf("%w: %s", ErrStepExists, step.Name())
		}
		m.steps[step.Name()] = step
	}

	m.start = steps[0]
	m.active = steps[0]

	return nil
}

func (m *MachineFlow[D]) Execute(ctx context.Context, s DataServiceProvide[D]) error {
	m.active = m.start
	m.attempt = 0

	for m.current() != nil {
		if err := ctx.Err(); err != nil {
			return err
		}

		transition, stepErr := m.process(ctx, s)
		if transition == FinishTransition {
			return stepErr
		}

		if stepErr != nil && transition != ErrorTransition {
			return stepErr
		}

		if err := m.move(transition); err != nil {
			return errors.Join(stepErr, err)
		}
	}

	return nil
}

func (m *MachineFlow[D]) current() Step[DataServiceProvide[D]] {
	return m.active
}

func (m *MachineFlow[D]) process(ctx context.Context, s DataServiceProvide[D]) (Transition, error) {
	step := m.current()
	if step == nil {
		return UnknownTransition, ErrNoActiveStep
	}

	transition, err := step.process(ctx, s)
	if err != nil {
		return transition, fmt.Errorf("step %s: %w", step.Name(), err)
	}

	if transition == "" || transition == UnknownTransition {
		return UnknownTransition, fmt.Errorf("step %s: %w", step.Name(), ErrUnknownTransition)
	}

	return transition, nil
}

func (m *MachineFlow[D]) move(transition Transition) error {
	step := m.current()
	if step == nil {
		return ErrNoActiveStep
	}

	nextName, err := step.On(transition)
	if err != nil {
		return err
	}

	if nextName == FinishStep {
		m.active = nil
		m.attempt = 0

		return nil
	}

	next, ok := m.steps[nextName]
	if !ok {
		return fmt.Errorf("%w: %s", ErrStepNotFound, nextName)
	}

	if transition == ErrorTransition && nextName == step.Name() {
		m.attempt++
		if m.attempt > step.Retry() {
			return fmt.Errorf("%w: step %s after %d attempts", ErrRetriesExhausted, step.Name(), m.attempt)
		}
	} else {
		m.attempt = 0
	}

	m.active = next

	return nil
}

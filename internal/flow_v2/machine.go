package flow_v2

import (
	"context"
	"fmt"
	"sync"

	ds "github.com/siper92/akha/internal/data_sources"
)

var (
	_ Flow            = (*Machine)(nil)
	_ Runtime         = (*Machine)(nil)
	_ State           = (*MemState)(nil)
	_ ServiceProvider = (*ds.FolderProvider)(nil)
)

const DefaultMaxTransitions = 1000

type Machine struct {
	plan     ExecutionPlan
	steps    map[string]Step
	services ServiceProvider
	state    State
	max      int
}

func NewMachine(max int) *Machine {
	if max <= 0 {
		max = DefaultMaxTransitions
	}
	return &Machine{max: max}
}

func (m *Machine) Init(plan ExecutionPlan, steps []Step, sp ServiceProvider) error {
	m.steps = make(map[string]Step, len(steps))
	for _, s := range steps {
		if _, ok := m.steps[s.ID()]; ok {
			return fmt.Errorf("%w: %s", ErrDuplicateStep, s.ID())
		}
		m.steps[s.ID()] = s
	}
	m.plan, m.services, m.state = plan, sp, NewMemState()
	return nil
}

func (m *Machine) Services() ServiceProvider { return m.services }

func (m *Machine) State() State { return m.state }

func (m *Machine) Run(ctx context.Context) error {
	if m.plan == nil {
		return ErrNotInit
	}
	id := m.plan.Start()
	for n := 0; id != End; n++ {
		if n >= m.max {
			return fmt.Errorf("%w: %d", ErrMaxTransitions, m.max)
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		step, ok := m.steps[id]
		if !ok {
			return fmt.Errorf("%w: %s", ErrUnknownStep, id)
		}
		outcome, err := step.Execute(ctx, m)
		if err != nil {
			return err
		}
		if id, err = m.plan.Next(id, outcome); err != nil {
			return err
		}
	}
	return nil
}

type MemState struct {
	mu sync.RWMutex
	m  map[string]any
}

func NewMemState() *MemState { return &MemState{m: map[string]any{}} }

func (s *MemState) Get(id string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[id]
	return v, ok
}

func (s *MemState) Set(id string, v any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[id] = v
}

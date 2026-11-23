package v3

import (
	"context"
	"fmt"
)

var _ Flow = (*StepFlow[Data])(nil)

type StepFlow[D any] struct {
	sp    FlowServices
	store Store[D]
	steps []Step[D]
	state State
}

func NewFlow[D any](sp FlowServices, store Store[D], steps ...Step[D]) *StepFlow[D] {
	return &StepFlow[D]{sp: sp, store: store, steps: steps, state: StateIdle}
}

func (f *StepFlow[D]) State() State { return f.state }

func (f *StepFlow[D]) Run(ctx context.Context) error {
	if f.state != StateIdle {
		return fmt.Errorf("%w: %s", ErrFlowState, f.state)
	}

	debug := f.sp.DebugService()
	f.state = StateRunning
	for _, step := range f.steps {
		debug.Log("step %s: start", step.Name())
		if err := step.Exec(ctx, f.sp, f.store); err != nil {
			f.state = StateFailed
			debug.Log("step %s: failed: %v", step.Name(), err)
			return fmt.Errorf("step %s: %w", step.Name(), err)
		}
		debug.Log("step %s: done", step.Name())
	}
	f.state = StateDone

	return nil
}

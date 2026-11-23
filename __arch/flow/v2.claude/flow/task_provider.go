package flow

import "fmt"

// ---------------------------------------------------------------------------
// Task service provider (SP)
//
// Each step builds one TaskProvider per execution from the flow's Provider.
// It is a scoped view:
//   - its own local Registry (step-level overrides, e.g. a different model)
//   - falls back to the flow SP for everything else
//   - carries execution identity (flow id, step name, attempt)
//   - exposes the step's scratch space so tasks can pass data along
//
// Tasks depend only on TaskProvider, never on Flow or Step internals.
// ---------------------------------------------------------------------------

type TaskProvider struct {
	local   *Registry
	FlowID  string
	Step    string
	Attempt int
	Scratch *Data // step-local; discarded after the step
	Shared  *Data // flow-level; visible to all steps
}

func newTaskProvider(flowSP Provider, sc *StepContext, overrides func(*Registry)) *TaskProvider {
	local := NewRegistry()
	local.parent = flowSP
	if overrides != nil {
		overrides(local)
	}
	return &TaskProvider{
		local:   local,
		FlowID:  sc.FlowID,
		Step:    sc.Step,
		Attempt: sc.Attempt,
		Scratch: NewData(),
		Shared:  sc.Data,
	}
}

// Lookup makes TaskProvider itself a Provider, so Resolve[T] works on it.
func (tp *TaskProvider) Lookup(key string) (any, bool) { return tp.local.Lookup(key) }

// API returns the API client visible to this task.
func (tp *TaskProvider) API() (APIClient, error) { return Resolve[APIClient](tp, KeyAPI) }

// Tool returns one tool by name from the visible toolbox.
func (tp *TaskProvider) Tool(name string) (Tool, error) {
	tb, err := Resolve[Toolbox](tp, KeyTools)
	if err != nil {
		return nil, err
	}
	t, ok := tb[name]
	if !ok {
		return nil, fmt.Errorf("task provider: tool %q not available in step %q", name, tp.Step)
	}
	return t, nil
}

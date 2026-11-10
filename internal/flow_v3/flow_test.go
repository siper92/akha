package flow_v3

import (
	"context"
	"testing"
)

var def = `
graph:
  clarify:
    done: plan
    error:
      next: clarify
      retry: 2
`

func TestStepDefToTransitionMap(t *testing.T) {
	mapKey, err := stepDefToTransitionMap(def)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	transitions := mapKey["clarify"]
	if transitions == nil {
		t.Errorf("expected 'clarify' step to be present in the graph")
	}

	if transitions["done"].Next != "plan" {
		t.Errorf("expected 'done' transition to lead to 'plan', got '%s'", transitions["done"].Next)
	}

	if transitions["error"].Next != "clarify" {
		t.Errorf("expected 'error' transition to lead to 'clarify', got '%s'", transitions["error"].Next)
	}

	if transitions["error"].Retry != 2 {
		t.Errorf("expected 'error' transition to have retry count of 2, got %d", transitions["error"].Retry)
	}
}

func TestNexStepUnmarshalYAML(t *testing.T) {
	machineFlow, err := NewMachineFlow(def, map[StepName]StepExec{
		"clarify": func(ctx context.Context, d any) (Transition, any, error) {
			return DoneTransition, nil, nil
		},
		"plan": func(ctx context.Context, d any) (Transition, any, error) {
			return FinishTransition, nil, nil
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if machineFlow == nil {
		t.Errorf("expected machineFlow to be non-nil")
	}

	process, err := machineFlow.process()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if process != DoneTransition {
		t.Errorf("expected process to return 'done', got '%s'", process)
	}

	machineFlow.move(process)
	if machineFlow.current().Name() != "plan" {
		t.Errorf("expected current step to be 'plan', got '%s'", machineFlow.current().Name())
	}

	process, err = machineFlow.process()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if process != FinishTransition {
		t.Errorf("expected process to return 'finish', got '%s'", process)
	}

	if process == FinishTransition {
		return
	}
}

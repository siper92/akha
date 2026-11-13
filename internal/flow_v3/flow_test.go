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
	machineFlow, err := NewMachineFlow(def, map[StepName]StepExec[map[string]string]{
		"clarify": func(ctx context.Context, d DataServiceProvide[map[string]string]) (Transition, error) {
			d.SetData(map[string]string{"clarify": "some data"})
			return DoneTransition, nil
		},
		"plan": func(ctx context.Context, d DataServiceProvide[map[string]string]) (Transition, error) {
			d.Update(func(data map[string]string) map[string]string {
				data["plan"] = "some data"
				return data
			})
			return FinishTransition, nil
		},
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if machineFlow == nil {
		t.Errorf("expected machineFlow to be non-nil")
	}

	ctx := context.Background()
	dataProvide := NewStepData(map[string]string{})

	process, err := machineFlow.process(ctx, dataProvide)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if process != DoneTransition {
		t.Errorf("expected process to return 'done', got '%s'", process)
	}

	dataCheck := dataProvide.Data()
	if dataCheck["clarify"] != "some data" {
		t.Errorf("expected data to contain 'clarify' key with value 'some data', got '%s'", dataCheck["clarify"])
	}

	machineFlow.move(process)
	if machineFlow.current().Name() != "plan" {
		t.Errorf("expected current step to be 'plan', got '%s'", machineFlow.current().Name())
	}

	process, err = machineFlow.process(ctx, dataProvide)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if process != FinishTransition {
		t.Errorf("expected process to return 'finish', got '%s'", process)
	}

	dataCheck = dataProvide.Data()
	if dataCheck["plan"] != "some data" {
		t.Errorf("expected data to contain 'plan' key with value 'some data', got '%s'", dataCheck["plan"])
	}

	if process == FinishTransition {
		return
	}
}

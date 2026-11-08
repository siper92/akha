package flow_v3

import "testing"

const simpleFlowSchema = `
graph:
  clarify:
    done: plan
    error:
      next: clarify
      retry: 2
  plan:
    done: finish
    error: plan
`

func TestSimpleFlow(t *testing.T) {
	graph, err := NewSimpleFlowGraph(simpleFlowSchema)
	if err != nil {
		t.Fatalf("Failed to create graph: %v", err)
	}

	current := graph.Current()
	if current.Name() != "clarify" {
		t.Errorf("Expected current step to be 'clarify', got '%s'", current.Name())
	}

	next, err := graph.On(NewStep("done", 0))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if next.Name() != "plan" {
		t.Errorf("Expected next step to be 'plan', got '%s'", next.Name())
	}

	graph.current = "plan"
	next, err = graph.On(NewStep("done", 0))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if next.Name() != "finish" {
		t.Errorf("Expected next step to be 'finish', got '%s'", next.Name())
	}

	graph.current = "clarify"
	next, err = graph.On(NewStep("error", 0))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if next.Name() != "clarify" {
		t.Errorf("Expected next step to be 'clarify', got '%s'", next.Name())
	}

	if graph.Attempt() != 1 {
		t.Errorf("Expected attempt to be 1, got %d", graph.Attempt())
	}
}

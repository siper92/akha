package flow_v3

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var _ Step = (*SimpleStep)(nil)

type SimpleStep struct {
	name  string
	retry int
}

func (s *SimpleStep) Name() string {
	return s.name
}

func (s *SimpleStep) Retry() int {
	return s.retry
}

func NewStep(name string, retry int) *SimpleStep {
	return &SimpleStep{
		name:  name,
		retry: retry,
	}
}

// ####################################

var (
	ErrStepNotFound       = errors.New("step not found")
	ErrTransitionNotFound = errors.New("transition not found")
	ErrRetriesExhausted   = errors.New("retries exhausted")
	ErrInvalidGraph       = errors.New("invalid graph")
)

var _ FlowGraph = (*SimpleFlowGraph)(nil)

type flowNode struct {
	active      Step
	transitions map[string]Step
}

type SimpleFlowGraph struct {
	current string
	attempt int
	steps   map[string]flowNode
}

func (g *SimpleFlowGraph) Current() Step {
	return g.steps[g.current].active
}

func (g *SimpleFlowGraph) Attempt() int {
	return g.attempt
}

func (g *SimpleFlowGraph) On(event Step) (Step, error) {
	if event == nil {
		return nil, fmt.Errorf("%w: nil event", ErrTransitionNotFound)
	}
	node, ok := g.steps[g.current]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrStepNotFound, g.current)
	}

	next, ok := node.transitions[event.Name()]
	if !ok {
		return nil, fmt.Errorf("%w: %q on step %q", ErrTransitionNotFound, event.Name(), g.current)
	}

	if _, ok := g.steps[next.Name()]; !ok {
		return nil, fmt.Errorf("%w: %q", ErrStepNotFound, next.Name())
	}

	if next.Name() == g.current && next.Retry() > 0 {
		if g.attempt >= next.Retry() {
			return nil, fmt.Errorf("%w: step %q after %d retries", ErrRetriesExhausted, g.current, g.attempt)
		}
		g.attempt++
	} else {
		g.attempt = 0
	}

	g.current = next.Name()
	return next, nil
}

// NewSimpleFlowGraph builds a graph from YAML of the form
//
//	graph:
//	  <step>:
//	    <event>: <next step>          # short form, no retry limit
//	    <event>:                      # long form
//	      next: <next step>
//	      retry: <n>
//
// The YAML is decoded into plain maps and then checked by hand, with no
// custom unmarshaller. A leaf value is either a string (short form) or a
// map (long form), so the inner map holds `any` rather than `string`.
//
// Go maps have no order, so the start step is the one step that no other
// step points to (self-loops don't count). Targets that are not declared
// as steps (e.g. "finish") become terminal steps.
func NewSimpleFlowGraph(s string) (*SimpleFlowGraph, error) {
	var doc struct {
		Graph map[string]map[string]any `yaml:"graph"`
	}
	if err := yaml.Unmarshal([]byte(strings.TrimSpace(s)), &doc); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidGraph, err)
	}
	if len(doc.Graph) == 0 {
		return nil, fmt.Errorf("%w: \"graph\" must be a non-empty mapping", ErrInvalidGraph)
	}

	g := &SimpleFlowGraph{steps: make(map[string]flowNode, len(doc.Graph))}
	incoming := make(map[string]bool) // steps reached from a different step

	for name, events := range doc.Graph {
		if name == "" {
			return nil, fmt.Errorf("%w: empty step name", ErrInvalidGraph)
		}
		node := flowNode{
			active:      NewStep(name, 0),
			transitions: make(map[string]Step, len(events)),
		}
		for event, raw := range events {
			next, retry, err := parseTransition(raw)
			if err != nil {
				return nil, fmt.Errorf("%w: step %q, event %q: %v", ErrInvalidGraph, name, event, err)
			}
			node.transitions[event] = NewStep(next, retry)
			if next != name {
				incoming[next] = true
			}
		}
		g.steps[name] = node
	}

	// Undeclared targets become terminal steps.
	for target := range incoming {
		if _, ok := g.steps[target]; !ok {
			g.steps[target] = flowNode{active: NewStep(target, 0), transitions: map[string]Step{}}
		}
	}

	// Find the start step.
	var starts []string
	for name := range doc.Graph {
		if !incoming[name] {
			starts = append(starts, name)
		}
	}
	if len(starts) != 1 {
		sort.Strings(starts)
		return nil, fmt.Errorf("%w: need exactly one start step (no incoming transitions), found %v",
			ErrInvalidGraph, starts)
	}
	g.current = starts[0]

	return g, nil
}

func parseTransition(raw any) (string, int, error) {
	switch v := raw.(type) {
	case string:
		if v == "" {
			return "", 0, errors.New("empty next step")
		}
		return v, 0, nil

	case map[string]any:
		var next string
		var retry int
		for k, val := range v {
			switch k {
			case "next":
				s, ok := val.(string)
				if !ok || s == "" {
					return "", 0, fmt.Errorf("next must be a non-empty string, got %v", val)
				}
				next = s
			case "retry":
				n, ok := val.(int)
				if !ok {
					return "", 0, fmt.Errorf("retry must be an integer, got %v", val)
				}
				if n < 0 {
					return "", 0, fmt.Errorf("retry must be >= 0, got %d", n)
				}
				retry = n
			default:
				return "", 0, fmt.Errorf("unknown key %q", k)
			}
		}
		if next == "" {
			return "", 0, errors.New("missing next")
		}
		return next, retry, nil

	default:
		return "", 0, fmt.Errorf("must be a step name or {next, retry}, got %v", raw)
	}
}

package flow_v3

import "errors"

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

// ####################################

var ErrStepNotFound = errors.New("step not found")

var _ FlowGraph = (*SimpleFlowGraph)(nil)

type SimpleFlowGraph struct {
	current Step
	retry   int
	on      map[Step]Step
}

func (s *SimpleFlowGraph) Current() Step {
	return s.current
}

func (s *SimpleFlowGraph) On(step Step) (Step, error) {
	if next, ok := s.on[step]; ok {
		if next.Retry() > s.retry {
			s.retry++
			return step, nil
		}

		s.current = next
		s.retry = 0

		return next, nil
	}

	return nil, ErrStepNotFound
}

func NewSimpleFlowGraph(schema string) *SimpleFlowGraph {
	return &SimpleFlowGraph{
		current: &SimpleStep{name: "clarify", retry: 0},
		retry:   0,
		on: map[Step]Step{
			&SimpleStep{name: "clarify", retry: 0}:   &SimpleStep{name: "plan", retry: 0},
			&SimpleStep{name: "plan", retry: 0}:      &SimpleStep{name: "task", retry: 0},
			&SimpleStep{name: "task", retry: 0}:      &SimpleStep{name: "implement", retry: 0},
			&SimpleStep{name: "implement", retry: 0}: &SimpleStep{name: "finish", retry: 0},
		},
	}
}

package flow

import "fmt"

// FlowState is the state of a flow instance (the state machine itself).
type FlowState int

const (
	FlowCreated FlowState = iota
	FlowRunning
	FlowCompleted
	FlowFailed
	FlowCancelled
)

func (s FlowState) String() string {
	switch s {
	case FlowCreated:
		return "created"
	case FlowRunning:
		return "running"
	case FlowCompleted:
		return "completed"
	case FlowFailed:
		return "failed"
	case FlowCancelled:
		return "cancelled"
	}
	return fmt.Sprintf("FlowState(%d)", int(s))
}

// Terminal reports whether no further transitions are possible.
func (s FlowState) Terminal() bool {
	return s == FlowCompleted || s == FlowFailed || s == FlowCancelled
}

// flowTransitions is the single source of truth for legal flow transitions.
var flowTransitions = map[FlowState][]FlowState{
	FlowCreated: {FlowRunning, FlowCancelled},
	FlowRunning: {FlowCompleted, FlowFailed, FlowCancelled},
}

func canTransition(from, to FlowState) bool {
	for _, s := range flowTransitions[from] {
		if s == to {
			return true
		}
	}
	return false
}

// StepState is the lifecycle state of one step execution.
type StepState int

const (
	StepPending StepState = iota
	StepRunning
	StepSucceeded
	StepFailed
	StepSkipped
)

func (s StepState) String() string {
	switch s {
	case StepPending:
		return "pending"
	case StepRunning:
		return "running"
	case StepSucceeded:
		return "succeeded"
	case StepFailed:
		return "failed"
	case StepSkipped:
		return "skipped"
	}
	return fmt.Sprintf("StepState(%d)", int(s))
}

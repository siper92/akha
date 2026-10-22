package types

import (
	"github.com/siper92/akha/internal/flow"
	"github.com/siper92/akha/internal/state_machine"
)

type TaskStatus state_machine.MachineState

const (
	Loading    TaskStatus = "loading"
	Clarifying TaskStatus = "clarifying"
	Planning   TaskStatus = "planning"
	Running    TaskStatus = "running"
	RunningSub TaskStatus = "running_sub"
)

type TaskPlan interface {
	Flow() flow.Flow
}

type TaskRun interface {
	Status() TaskStatus
	GetStep() flow.FlowStep
	Dir() string // dir for output
	Plan() TaskPlan
}

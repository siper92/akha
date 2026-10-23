package task

import (
	"github.com/siper92/akha/cmd/akha_runner/internal/types"
	"github.com/siper92/akha/internal/flow"
	"github.com/siper92/akha/internal/state_machine"
)

type Task struct {
	UID           string
	ProjectID     int64
	Status        types.TaskStatus
	ExecutionPlan flow.ExecutionPlan // code of the execution plan
	WorkerID      int64
	Content       string
	Output        string
	Error         string
}

func BuildNewMachine() *state_machine.Machine[*Task, types.TaskStatus] {
	m := state_machine.NewMachine[*Task, types.TaskStatus](types.Loading)
	return m
}

package task

import (
	"github.com/siper92/akha/cmd/akha_runner/internal/types"
	"github.com/siper92/akha/internal/state_machine"
)

type Task struct {
	UID       string
	ProjectID int64
	Status    types.TaskStatus
}

func BuildNewMachine() *state_machine.Machine[*Task, types.TaskStatus] {
	m := state_machine.NewMachine[*Task, types.TaskStatus](types.Loading)
	return m
}

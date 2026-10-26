package flow

import (
	"context"

	"github.com/siper92/akha/internal/state_machine"
)

var _ Task[struct{}, struct{}] = (*MachineTask[struct{}, struct{}, string])(nil)

type TaskState[D, R any] struct {
	Data   D
	Scope  Scope[R]
	Result R
}

type MachineTask[D, R any, S comparable] struct {
	id      ID
	machine state_machine.Runner[TaskState[D, R], S]
}

func NewTask[D, R any, S comparable](id ID, m state_machine.Runner[TaskState[D, R], S]) *MachineTask[D, R, S] {
	return &MachineTask[D, R, S]{id: id, machine: m}
}

func (t *MachineTask[D, R, S]) ID() ID { return t.id }

func (t *MachineTask[D, R, S]) Run(ctx context.Context, data D, scope Scope[R]) (R, error) {
	st := &TaskState[D, R]{Data: data, Scope: scope}
	if _, err := t.machine.Run(ctx, st); err != nil {
		var zero R
		return zero, err
	}
	return st.Result, nil
}

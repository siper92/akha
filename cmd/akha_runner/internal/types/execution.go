package types

import "github.com/siper92/akha/internal/flow"

type TaskRunner interface {
	Load(task string) error
	Clarify() error
	Plan() error
	Run() error
}

var _ flow.FlowStep = (*TaskStep)(nil)

type TaskStep struct {
	content string
}

func (t TaskStep) ID() string {
	return "task"
}

func (t TaskStep) Process() error {
	return nil
}

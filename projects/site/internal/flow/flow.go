package flow

import (
	"context"
	"errors"
	"io"

	flow "github.com/siper92/akha/internal/flow_v3"
)

const (
	ClarifyStepName   flow.StepName = "clarify"
	PlanStepName      flow.StepName = "plan"
	TaskStepName      flow.StepName = "task"
	ImplementStepName flow.StepName = "implement"
)

const TaskDoneTransition flow.Transition = "task_done"

const (
	inputKey             = "input"
	dbFileName           = "site.db"
	implementDirName     = "implement"
	implementMaxAttempts = 3
)

var (
	ErrMissingData = errors.New("missing step data")
	ErrTaskFailed  = errors.New("task failed")
)

var textOutput = OutputConfig{Format: OutputText}

type SiteData struct {
	Input    string
	Clarify  string
	Plan     string
	Tasks    TaskPlan
	TasksRaw string
	StoryIDs []int64
}

type siteStepRunner interface {
	Name() flow.StepName
	run(ctx context.Context, s SiteService) (flow.Transition, error)
}

type AskVia func(s SiteService) AskFunc

type StoreFunc func(ctx context.Context, s SiteService, answer string) error

func newSteps(steps ...siteStepRunner) flow.Steps[SiteData] {
	out := make(flow.Steps[SiteData], len(steps))
	for _, st := range steps {
		out[st.Name()] = siteStep(st.run)
	}

	return out
}

type SiteFlow struct {
	machine *flow.MachineFlow[SiteData]
	service SiteService
	closer  io.Closer
}

func (s *SiteFlow) Close() error {
	if s.closer == nil {
		return nil
	}

	return s.closer.Close()
}

func (s *SiteFlow) Run(ctx context.Context, input string) error {
	s.service.SetData(SiteData{Input: input})

	return s.machine.Execute(ctx, s.service)
}

func (s *SiteFlow) Execute(ctx context.Context) error {
	input, _ := ctx.Value(inputKey).(string)

	return s.Run(ctx, input)
}

func (s *SiteFlow) Data() SiteData {
	return s.service.Data()
}

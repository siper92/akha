package site

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/siper92/akha/internal/ai"
	"github.com/siper92/akha/internal/cache"
	flow "github.com/siper92/akha/internal/flow_v3"
	"github.com/siper92/akha/internal/parse"
	"github.com/siper92/akha/projects/site/sdk/db"
	dbsdk "github.com/siper92/akha/projects/site/sdk/db-sdk"
)

const (
	ClarifyStepName   flow.StepName = "clarify"
	PlanStepName      flow.StepName = "plan"
	TaskStepName      flow.StepName = "task"
	ImplementStepName flow.StepName = "implement"
)

const (
	inputKey             = "input"
	dbFileName           = "site.db"
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

type SiteFlow struct {
	machine *flow.MachineFlow[SiteData]
	service *siteService
	closer  io.Closer
}

func NewSiteBuildSteps(key string, cacheDir string) (*SiteFlow, error) {
	aiClient, err := ai.NewAnthropic(key)
	if err != nil {
		return nil, err
	}

	fileCache, err := cache.NewFile(cacheDir)
	if err != nil {
		return nil, err
	}

	conn, err := db.Init(context.Background(), filepath.Join(cacheDir, dbFileName))
	if err != nil {
		return nil, err
	}

	f, err := NewSiteFlow(Providers{
		AI:    newAiAsker(aiClient),
		Cache: fileCache,
		Tasks: db.NewRepository(conn),
	}, conn)
	if err != nil {
		return nil, errors.Join(err, conn.Close())
	}

	return f, nil
}

func NewSiteFlow(p Providers, closer io.Closer) (*SiteFlow, error) {
	machine, err := flow.NewMachineFlow(flowDef, siteSteps())
	if err != nil {
		return nil, err
	}

	return &SiteFlow{
		machine: machine,
		service: newSiteService(SiteData{}, p),
		closer:  closer,
	}, nil
}

func siteSteps() flow.Steps[SiteData] {
	return flow.Steps[SiteData]{
		ClarifyStepName: askStep{
			name:  ClarifyStepName,
			def:   GetClarifyDef,
			out:   textOutput,
			input: func(d SiteData) string { return d.Input },
			store: func(_ context.Context, s SiteService, answer string) error {
				d := s.Data()
				d.Clarify = answer
				s.SetData(d)
				return nil
			},
		}.exec(),
		PlanStepName: askStep{
			name:  PlanStepName,
			def:   GetPlanDef,
			out:   textOutput,
			input: func(d SiteData) string { return d.Clarify },
			store: func(_ context.Context, s SiteService, answer string) error {
				d := s.Data()
				d.Plan = answer
				s.SetData(d)
				return nil
			},
		}.exec(),
		TaskStepName: askStep{
			name:  TaskStepName,
			def:   GetTaskDef,
			out:   tasksOutput,
			input: func(d SiteData) string { return d.Plan },
			store: storeTasks,
		}.exec(),
		ImplementStepName: siteStep(implement),
	}
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

type askStep struct {
	name  flow.StepName
	def   func(SiteData) (*parse.MDFlowStep, error)
	out   OutputConfig
	input func(SiteData) string
	store func(ctx context.Context, s SiteService, answer string) error
}

func (a askStep) exec() flow.StepExec[SiteData] {
	return siteStep(a.run)
}

func (a askStep) run(ctx context.Context, s SiteService) (flow.Transition, error) {
	answer, t, err := a.answer(ctx, s)
	if err != nil {
		return t, fmt.Errorf("%s: %w", a.name, err)
	}

	var storeErr error
	s.Update(func(d SiteData) SiteData {
		next := d
		if storeErr = a.store(ctx, s, answer); storeErr != nil {
			return d
		}

		return next
	})

	if storeErr != nil {
		return flow.ErrorTransition, fmt.Errorf("%s: store: %w", a.name, storeErr)
	}

	return flow.DoneTransition, nil
}

func (a askStep) answer(ctx context.Context, s SiteService) (string, flow.Transition, error) {
	key := string(a.name)

	cached, ok, err := s.Cache().Get(ctx, key)
	if err != nil {
		return "", flow.UnknownTransition, fmt.Errorf("read cache: %w", err)
	}

	if ok {
		answer, err := a.out.Check(string(cached))
		if err != nil {
			return "", flow.ErrorTransition, fmt.Errorf("cached response: %w", err)
		}

		return answer, flow.DoneTransition, nil
	}

	d := s.Data()
	input := a.input(d)
	if input == "" {
		return "", flow.UnknownTransition, ErrMissingData
	}

	step, err := a.def(d)
	if err != nil {
		return "", flow.UnknownTransition, fmt.Errorf("load step def: %w", err)
	}

	answer, err := s.AI().AskLv3(ctx, getAiPrompt(input, step), a.out)
	if err == nil {
		answer, err = a.out.Check(answer)
	}

	if err != nil {
		return "", flow.ErrorTransition, fmt.Errorf("ask: %w", err)
	}

	if err = s.Cache().Put(ctx, key, []byte(answer), cache.NoExpiry); err != nil {
		return "", flow.UnknownTransition, fmt.Errorf("save response: %w", err)
	}

	return answer, flow.DoneTransition, nil
}

func storeTasks(ctx context.Context, s SiteService, answer string) error {
	tasks, err := parseTasks(answer)
	if err != nil {
		return err
	}

	ids, err := saveTasks(ctx, s, tasks)
	if err != nil {
		return err
	}

	d := s.Data()
	d.TasksRaw, d.Tasks, d.StoryIDs = answer, tasks, ids
	s.SetData(d)

	return nil
}

func implement(ctx context.Context, s SiteService) (flow.Transition, error) {
	d := s.Data()
	if d.TasksRaw == "" || len(d.StoryIDs) == 0 {
		return flow.UnknownTransition, fmt.Errorf("%s: %w", ImplementStepName, ErrMissingData)
	}

	for _, storyID := range d.StoryIDs {
		tasks, err := s.Tasks().ListTasks(ctx, storyID)
		if err != nil {
			return flow.ErrorTransition, fmt.Errorf("%s: list tasks: %w", ImplementStepName, err)
		}

		for _, t := range tasks {
			if db.TaskStatus(t.Status) == db.TaskStatusCompleted {
				continue
			}

			if err = implementTask(ctx, s, d.Plan, t); err != nil {
				return flow.ErrorTransition, fmt.Errorf("%s: task %q: %w", ImplementStepName, t.TaskKey, err)
			}
		}
	}

	return flow.DoneTransition, nil
}

func implementTask(ctx context.Context, s SiteService, plan string, t dbsdk.FlowTask) error {
	setStatus := func(status db.TaskStatus, attempt int64, msg string) error {
		if err := s.Tasks().UpdateTaskStatus(ctx, t.ID, status, attempt, msg); err != nil {
			return fmt.Errorf("update status: %w", err)
		}

		return nil
	}

	if err := setStatus(db.TaskStatusInProgress, 0, ""); err != nil {
		return err
	}

	prompt := plan + "\n\n# task " + t.TaskKey + ": " + t.Label + "\n\n" + t.Content
	key := fmt.Sprintf("%s_%s_%d", ImplementStepName, t.TaskKey, t.ID)

	var lastErr error
	for attempt := int64(1); attempt <= implementMaxAttempts; attempt++ {
		answer, err := s.AI().AskLv3(ctx, prompt, textOutput)
		if err == nil {
			if err = s.Cache().Put(ctx, key, []byte(answer), cache.NoExpiry); err != nil {
				return fmt.Errorf("save response: %w", err)
			}

			return setStatus(db.TaskStatusCompleted, attempt, "")
		}

		lastErr = err
		if err = setStatus(db.TaskStatusInProgress, attempt, lastErr.Error()); err != nil {
			return err
		}
	}

	if err := setStatus(db.TaskStatusFailed, implementMaxAttempts, lastErr.Error()); err != nil {
		return err
	}

	return fmt.Errorf("%w: %w", ErrTaskFailed, lastErr)
}

func getAiPrompt(input string, d *parse.MDFlowStep) string {
	prompt := input + "\n\n"
	for _, name := range []string{"idea", "questions", "answers"} {
		if section, ok := d.Section(name); ok && section != nil {
			prompt += "\n" + section.String()
		}
	}

	return prompt
}

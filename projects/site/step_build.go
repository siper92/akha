package site

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/siper92/akha/internal/ai"
	cache2 "github.com/siper92/akha/internal/cache"
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
	data    *siteService
	conn    *sql.DB
}

func NewSiteBuildSteps(key string, cacheDir string) (*SiteFlow, error) {
	aiClient, err := ai.NewAnthropic(key)
	if err != nil {
		return nil, err
	}
	aiService := newAiAsker(aiClient)

	cache, err := cache2.NewFile(cacheDir)
	if err != nil {
		return nil, err
	}

	machine, err := flow.NewMachineFlow(flowDef, map[flow.StepName]flow.StepExec[SiteData]{
		ClarifyStepName: askStep(aiService, ClarifyStepName, GetClarifyDef, OutputConfig{Format: OutputText},
			func(d SiteData) string { return d.Input },
			func(ds SiteService, d SiteData, answer string) SiteData {
				d.Clarify = answer
				return d
			},
		),
		PlanStepName: askStep(aiService, PlanStepName, GetPlanDef, OutputConfig{Format: OutputText},
			func(d SiteData) string { return d.Clarify },
			func(ds SiteService, d SiteData, answer string) SiteData {
				d.Plan = answer
				return d
			},
		),
		TaskStepName: askStep(aiService, TaskStepName, GetTaskDef, tasksOutput,
			func(d SiteData) string { return d.Plan },
			func(ds SiteService, d SiteData, answer string) SiteData {
				d.TasksRaw = answer
				tasks, err := parseTasks(answer)
				if err != nil {
					_ = fmt.Errorf("%s: %w", TaskStepName, err)
					return d
				}
				d.Tasks = tasks

				s, err := asSiteService(ds)
				if err != nil {
					_ = fmt.Errorf("%s: %w", TaskStepName, err)
					return d
				}

				storyIDs, err := saveTasks(context.Background(), s, tasks)
				if err != nil {
					_ = fmt.Errorf("%s: %w", TaskStepName, err)
					return d
				}
				d.StoryIDs = storyIDs

				return d
			},
		),
		ImplementStepName: implementStep(aiService),
	})

	if err != nil {
		return nil, err
	}

	conn, err := db.Init(context.Background(), filepath.Join(cacheDir, dbFileName))
	if err != nil {
		return nil, err
	}

	return &SiteFlow{
		machine: machine,
		data:    newSiteService(SiteData{}, cache, db.NewRepository(conn)),
		conn:    conn,
	}, nil
}

func (s *SiteFlow) Close() error {
	return s.conn.Close()
}

func (s *SiteFlow) Run(ctx context.Context, input string) error {
	s.data.SetData(SiteData{Input: input})

	return s.machine.Execute(ctx, s.data)
}

func (s *SiteFlow) Execute(ctx context.Context) error {
	input, _ := ctx.Value(inputKey).(string)

	return s.Run(ctx, input)
}

func (s *SiteFlow) Data() SiteData {
	return s.data.Data()
}

func askStep(
	a asker,
	name flow.StepName,
	def func(SiteData) (*parse.MDFlowStep, error),
	out OutputConfig,
	get func(SiteData) string,
	set func(SiteService, SiteData, string) SiteData,
) flow.StepExec[SiteData] {
	return func(ctx context.Context, ds flow.DataServiceProvide[SiteData]) (flow.Transition, error) {
		s, err := asSiteService(ds)
		if err != nil {
			return flow.UnknownTransition, fmt.Errorf("%s: %w", name, err)
		}

		var ok bool
		var val []byte
		if val, ok, err = s.Cache().Get(ctx, string(name)); err != nil {
			return flow.UnknownTransition, fmt.Errorf("%s: save response: %w", name, err)
		}

		var answer string
		if ok {
			if answer, err = out.Check(string(val)); err != nil {
				return flow.ErrorTransition, fmt.Errorf("%s: cached response: %w", name, err)
			}
		} else {
			input := get(s.Data())
			if input == "" {
				return flow.UnknownTransition, fmt.Errorf("%s: %w", name, ErrMissingData)
			}

			d, err := def(s.Data())
			if err != nil {
				return "", fmt.Errorf("load step def: %w", err)
			}

			prompt := getAiPrompt(input, d)
			answer, err = a.AskLv3(ctx, prompt, out)
			if err != nil {
				return flow.ErrorTransition, fmt.Errorf("%s: ask: %w", name, err)
			}

			if answer, err = out.Check(answer); err != nil {
				return flow.ErrorTransition, fmt.Errorf("%s: %w", name, err)
			}

			if err = s.Cache().Put(
				ctx, string(name), []byte(answer), cache2.NoExpiry,
			); err != nil {
				return flow.UnknownTransition, fmt.Errorf("%s: save response: %w", name, err)
			}
		}

		s.Update(func(d SiteData) SiteData {
			return set(s, d, answer)
		})

		return flow.DoneTransition, nil
	}
}

func implementStep(a asker) flow.StepExec[SiteData] {
	return func(ctx context.Context, ds flow.DataServiceProvide[SiteData]) (flow.Transition, error) {
		s, err := asSiteService(ds)
		if err != nil {
			return flow.UnknownTransition, fmt.Errorf("%s: %w", ImplementStepName, err)
		}

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

				if err = implementTask(ctx, a, s, d.Plan, t); err != nil {
					return flow.ErrorTransition, fmt.Errorf("%s: task %q: %w", ImplementStepName, t.TaskKey, err)
				}
			}
		}

		return flow.DoneTransition, nil
	}
}

func implementTask(ctx context.Context, a asker, s SiteService, plan string, t dbsdk.FlowTask) error {
	repo := s.Tasks()
	if err := repo.UpdateTaskStatus(ctx, t.ID, db.TaskStatusInProgress, 0, ""); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	prompt := plan + "\n\n# task " + t.TaskKey + ": " + t.Label + "\n\n" + t.Content
	out := OutputConfig{Format: OutputText}

	var lastErr error
	for attempt := int64(1); attempt <= implementMaxAttempts; attempt++ {
		answer, err := a.AskLv3(ctx, prompt, out)
		if err == nil {
			answer, err = out.Check(answer)
		}

		if err == nil {
			key := fmt.Sprintf("%s_%s_%d", ImplementStepName, t.TaskKey, t.ID)
			if err = s.Cache().Put(ctx, key, []byte(answer), cache2.NoExpiry); err != nil {
				return fmt.Errorf("save response: %w", err)
			}

			if err = repo.UpdateTaskStatus(ctx, t.ID, db.TaskStatusCompleted, attempt, ""); err != nil {
				return fmt.Errorf("update status: %w", err)
			}

			return nil
		}

		lastErr = err
		if err = repo.UpdateTaskStatus(ctx, t.ID, db.TaskStatusInProgress, attempt, lastErr.Error()); err != nil {
			return fmt.Errorf("update status: %w", err)
		}
	}

	if err := repo.UpdateTaskStatus(ctx, t.ID, db.TaskStatusFailed, implementMaxAttempts, lastErr.Error()); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return fmt.Errorf("%w: %w", ErrTaskFailed, lastErr)
}

func getAiPrompt(input string, d *parse.MDFlowStep) string {
	var aiPrompt = input + "\n\n"
	for _, section := range []string{"idea", "questions", "answers"} {
		idea, ok := d.Section(section)
		if !ok {
			continue
		}

		if idea != nil {
			aiPrompt += "\n" + idea.String()
		}
	}

	return aiPrompt
}

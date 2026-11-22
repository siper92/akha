package flow

import (
	"context"
	_ "embed"
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/siper92/akha/internal/ai"
	"github.com/siper92/akha/internal/cache"
	flow "github.com/siper92/akha/internal/flow_v3"
	"github.com/siper92/akha/internal/parse"
	"github.com/siper92/akha/projects/site"
	"github.com/siper92/akha/projects/site/internal/db"
)

func GetFlowDef() string {
	return site.FlowDef
}

func GetTaskFlow() (fs.FS, error) {
	task, err := fs.Sub(site.Steps, "steps/task_proto")
	if err != nil {
		return nil, err
	}

	return task, nil
}

func GetClarifyDef(s SiteData) (*parse.MDFlowStep, error) {
	clarify, err := fs.ReadFile(site.Steps, "steps/task_proto/1.clarify/clarify.md")
	if err != nil {
		return nil, err
	}

	return parse.Parse(string(clarify))
}

func GetPlanDef(s SiteData) (*parse.MDFlowStep, error) {
	plan, err := fs.ReadFile(site.Steps, "steps/task_proto/2.plan/plan.md")
	if err != nil {
		return nil, err
	}

	return parse.Parse(string(plan))
}

func GetTaskDef(s SiteData) (*parse.MDFlowStep, error) {
	task, err := fs.ReadFile(site.Steps, "steps/task_proto/3.tasks/tasks.md")
	if err != nil {
		return nil, err
	}

	return parse.Parse(string(task))
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
	machine, err := flow.NewMachineFlow(site.FlowDef, siteSteps())
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

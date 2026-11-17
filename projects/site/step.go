package site

import (
	"embed"
	_ "embed"
	"io/fs"

	"github.com/siper92/akha/internal/parse"
)

//go:embed flow.yaml
var flowDef string

//go:embed all:steps
var steps embed.FS

func GetFlowDef() string {
	return flowDef
}

func GetTaskFlow() (fs.FS, error) {
	task, err := fs.Sub(steps, "steps/task_proto")
	if err != nil {
		return nil, err
	}

	return task, nil
}

func GetClarifyDef() (*parse.MDFlowStep, error) {
	clarify, err := fs.ReadFile(steps, "steps/task_proto/1.clarify/clarify.md")
	if err != nil {
		return nil, err
	}

	return parse.Parse(string(clarify))
}

func GetPlanDef() (*parse.MDFlowStep, error) {
	plan, err := fs.ReadFile(steps, "steps/task_proto/2.plan/plan.md")
	if err != nil {
		return nil, err
	}

	return parse.Parse(string(plan))
}

func GetTaskDef() (*parse.MDFlowStep, error) {
	task, err := fs.ReadFile(steps, "steps/task_proto/3.tasks/tasks.md")
	if err != nil {
		return nil, err
	}

	return parse.Parse(string(task))
}

package site

import (
	"context"
	"errors"
	"fmt"

	"github.com/siper92/akha/internal/ai"
	cache2 "github.com/siper92/akha/internal/cache"
	flow "github.com/siper92/akha/internal/flow_v3"
	"github.com/siper92/akha/internal/parse"
)

const (
	ClarifyStepName   flow.StepName = "clarify"
	PlanStepName      flow.StepName = "plan"
	TaskStepName      flow.StepName = "task"
	ImplementStepName flow.StepName = "implement"
)

const (
	inputKey = "input"
)

var ErrMissingData = errors.New("missing step data")

type SiteData struct {
	Input   string
	Clarify string
	Plan    string
	Tasks   string
}

type asker interface {
	AskLv1(ctx context.Context, prompt string) (string, error)
	AskLv2(ctx context.Context, prompt string) (string, error)
	AskLv3(ctx context.Context, prompt string) (string, error)
}

var _ asker = (*ai.Service)(nil)

type SiteFlow struct {
	machine *flow.MachineFlow[SiteData]
	data    *siteService
}

func NewSiteBuildSteps(key string, cacheDir string) (*SiteFlow, error) {
	aiService, err := ai.NewAnthropic(key)
	if err != nil {
		return nil, err
	}

	cache, err := cache2.NewFile(cacheDir)
	if err != nil {
		return nil, err
	}

	machine, err := flow.NewMachineFlow(flowDef, map[flow.StepName]flow.StepExec[SiteData]{
		ClarifyStepName: askStep(aiService, ClarifyStepName, GetClarifyDef,
			func(d SiteData) string { return d.Input },
			func(d SiteData, answer string) SiteData {
				d.Clarify = answer
				return d
			},
		),
		PlanStepName: askStep(aiService, PlanStepName, GetPlanDef,
			func(d SiteData) string { return d.Clarify },
			func(d SiteData, answer string) SiteData {
				d.Plan = answer
				return d
			},
		),
		TaskStepName: askStep(aiService, TaskStepName, GetTaskDef,
			func(d SiteData) string { return d.Plan },
			func(d SiteData, answer string) SiteData {
				d.Tasks = answer
				return d
			},
		),
		ImplementStepName: implementStep,
	})

	if err != nil {
		return nil, err
	}

	return &SiteFlow{
		machine: machine,
		data:    newSiteService(SiteData{}, cache),
	}, nil
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
	get func(SiteData) string,
	set func(SiteData, string) SiteData,
) flow.StepExec[SiteData] {
	return func(ctx context.Context, ds flow.DataServiceProvide[SiteData]) (flow.Transition, error) {
		s, err := asSiteService(ds)
		if err != nil {
			return flow.UnknownTransition, fmt.Errorf("%s: %w", name, err)
		}

		input := get(s.Data())
		if input == "" {
			return flow.UnknownTransition, fmt.Errorf("%s: %w", name, ErrMissingData)
		}

		d, err := def(s.Data())
		if err != nil {
			return "", fmt.Errorf("load step def: %w", err)
		}

		prompt := getAiPrompt(input, d)

		answer, err := a.AskLv2(ctx, prompt)
		if err != nil {
			return flow.ErrorTransition, fmt.Errorf("%s: ask: %w", name, err)
		}

		s.Update(func(d SiteData) SiteData {
			return set(d, answer)
		})

		if err = s.Cache().Put(ctx,
			string(name), []byte(answer), cache2.NoExpiry,
		); err != nil {
			return flow.UnknownTransition, fmt.Errorf("%s: save response: %w", name, err)
		}

		return flow.DoneTransition, nil
	}
}

func implementStep(_ context.Context, s flow.DataServiceProvide[SiteData]) (flow.Transition, error) {
	if s.Data().Tasks == "" {
		return flow.UnknownTransition, fmt.Errorf("%s: %w", ImplementStepName, ErrMissingData)
	}

	return flow.DoneTransition, nil
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

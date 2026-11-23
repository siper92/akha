package flow

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/anthropics/anthropic-sdk-go/v2"
	"github.com/anthropics/anthropic-sdk-go/v2/option"
	"github.com/siper92/akha/internal/ai"
	"github.com/siper92/akha/internal/cache"
	flow "github.com/siper92/akha/internal/flow_v3"
	"github.com/siper92/akha/internal/parse"
	akmcp "github.com/siper92/akha/projects/mcp/v2"
	"github.com/siper92/akha/projects/site"
	"github.com/siper92/akha/projects/site/internal/db"
)

var ErrAsk = errors.New("ask failed")

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
	return readStepDef("steps/task_proto/1.clarify/clarify.md")
}

func GetPlanDef(s SiteData) (*parse.MDFlowStep, error) {
	return readStepDef("steps/task_proto/2.plan/plan.md")
}

func GetTaskDef(s SiteData) (*parse.MDFlowStep, error) {
	return readStepDef("steps/task_proto/3.tasks/tasks.md")
}

func readStepDef(path string) (*parse.MDFlowStep, error) {
	b, err := fs.ReadFile(site.Steps, path)
	if err != nil {
		return nil, err
	}

	return parse.Parse(string(b))
}

func NewSiteFlow(key string, cacheDir string) (*SiteFlow, error) {
	aiClient, err := ai.NewAnthropic(key)
	if err != nil {
		return nil, err
	}

	fileCache, err := cache.NewFile(cacheDir)
	if err != nil {
		return nil, err
	}

	tools, err := akmcp.NewService(implementTools(filepath.Join(cacheDir, implementDirName))...)
	if err != nil {
		return nil, err
	}

	conn, err := db.Init(context.Background(), filepath.Join(cacheDir, dbFileName))
	if err != nil {
		return nil, err
	}

	f, err := newSiteFlow(Providers{
		AI:    newAiAsker(aiClient),
		Agent: akmcp.NewClaudeAgent(anthropic.NewClient(option.WithAPIKey(key)), tools, akmcp.DefaultModel, nil),
		Cache: fileCache,
		Tasks: db.NewRepository(conn),
	}, conn)
	if err != nil {
		return nil, errors.Join(err, conn.Close())
	}

	return f, nil
}

func implementTools(dir string) []akmcp.Tool {
	return []akmcp.Tool{
		akmcp.NewCreateFile(dir),
		akmcp.NewReadFile(dir),
	}
}

func newSiteFlow(p Providers, closer io.Closer) (*SiteFlow, error) {
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
	return newSteps(
		askStep{
			name:  ClarifyStepName,
			def:   GetClarifyDef,
			out:   textOutput,
			input: func(d SiteData) string { return d.Input },
			store: storeText(func(d *SiteData, v string) { d.Clarify = v }),
		},
		askStep{
			name:  PlanStepName,
			def:   GetPlanDef,
			out:   textOutput,
			input: func(d SiteData) string { return d.Clarify },
			store: storeText(func(d *SiteData, v string) { d.Plan = v }),
		},
		askStep{
			name:  TaskStepName,
			def:   GetTaskDef,
			out:   tasksOutput,
			input: func(d SiteData) string { return d.Plan },
			store: storeTasks,
		},
		implementStep{
			name:     ImplementStepName,
			out:      textOutput,
			attempts: implementMaxAttempts,
		},
	)
}

func viaLv3(s SiteService) AskFunc {
	return s.AI().AskLv3
}

func viaAgent(s SiteService) AskFunc {
	return agentAsk(s.Agent())
}

func askWith(via, fallback AskVia, s SiteService) AskFunc {
	if via == nil {
		return fallback(s)
	}

	return via(s)
}

var _ siteStepRunner = askStep{}

type askStep struct {
	name  flow.StepName
	def   func(SiteData) (*parse.MDFlowStep, error)
	out   OutputConfig
	input func(SiteData) string
	store StoreFunc
	via   AskVia
}

func (a askStep) Name() flow.StepName {
	return a.name
}

func (a askStep) run(ctx context.Context, s SiteService) (flow.Transition, error) {
	answer, err := cachedAsk(ctx, s.Cache(), string(a.name), a.out, a.prompt(s.Data()), askWith(a.via, viaLv3, s))
	if err != nil {
		return askTransition(err), err
	}

	if err = a.store(ctx, s, answer); err != nil {
		return flow.ErrorTransition, fmt.Errorf("store: %w", err)
	}

	return flow.DoneTransition, nil
}

func (a askStep) prompt(d SiteData) func() (string, error) {
	return func() (string, error) {
		input := a.input(d)
		if input == "" {
			return "", ErrMissingData
		}

		step, err := a.def(d)
		if err != nil {
			return "", fmt.Errorf("load step def: %w", err)
		}

		return getAiPrompt(input, step), nil
	}
}

func cachedAsk(ctx context.Context, c cache.Cache, key string, out OutputConfig, prompt func() (string, error), ask AskFunc) (string, error) {
	cached, ok, err := c.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("read cache: %w", err)
	}

	if ok {
		answer, err := out.Check(string(cached))
		if err != nil {
			return "", fmt.Errorf("cached response: %w", err)
		}

		return answer, nil
	}

	p, err := prompt()
	if err != nil {
		return "", err
	}

	answer, err := ask(ctx, p, out)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrAsk, err)
	}

	if answer, err = out.Check(answer); err != nil {
		return "", err
	}

	if err = c.Put(ctx, key, []byte(answer), cache.NoExpiry); err != nil {
		return "", fmt.Errorf("save response: %w", err)
	}

	return answer, nil
}

func askTransition(err error) flow.Transition {
	if errors.Is(err, ErrAsk) || errors.Is(err, ErrOutputFormat) {
		return flow.ErrorTransition
	}

	return flow.UnknownTransition
}

func storeText(set func(d *SiteData, v string)) StoreFunc {
	return func(_ context.Context, s SiteService, answer string) error {
		s.Update(func(d SiteData) SiteData {
			set(&d, answer)
			return d
		})

		return nil
	}
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

	s.Update(func(d SiteData) SiteData {
		d.TasksRaw, d.Tasks, d.StoryIDs = answer, tasks, ids
		return d
	})

	return nil
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

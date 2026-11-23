package flow

import (
	"context"
	"fmt"
	"time"
)

// StepContext is what the flow hands to a step for one execution.
type StepContext struct {
	FlowID   string
	Step     string
	Attempt  int      // 1-based
	Provider Provider // the flow SP
	Data     *Data    // flow-level shared data
}

// Result is what a step returns to the state machine. Besides the output,
// it can steer the machine: jump to a named step, or stop the flow early.
type Result struct {
	Output any
	Next   string // optional: name of the next step (branching / loops)
	Stop   bool   // optional: finish the flow successfully after this step
}

// Step is the contract the state machine knows about. Anything that
// implements it can be placed in any flow.
type Step interface {
	Name() string
	Execute(ctx context.Context, sc *StepContext) (Result, error)
}

// ErrorPolicy decides what the flow does when a step ultimately fails.
type ErrorPolicy int

const (
	FailFlow     ErrorPolicy = iota // default: flow -> failed
	ContinueFlow                    // record failure, move to next step
)

// StepConfig holds per-step execution policy, enforced by the flow, not by
// the step. This keeps step logic free of retry/timeout plumbing.
type StepConfig struct {
	Retries   int           // extra attempts after the first
	Backoff   time.Duration // wait between attempts
	Timeout   time.Duration // per attempt; 0 = none
	OnError   ErrorPolicy
	When      func(sc *StepContext) bool // skip the step when it returns false
	Overrides func(*Registry)            // step-scoped services for the task SP
}

// StepOption configures a StepConfig.
type StepOption func(*StepConfig)

func WithRetry(n int, backoff time.Duration) StepOption {
	return func(c *StepConfig) { c.Retries, c.Backoff = n, backoff }
}
func WithTimeout(d time.Duration) StepOption { return func(c *StepConfig) { c.Timeout = d } }
func WithOnError(p ErrorPolicy) StepOption   { return func(c *StepConfig) { c.OnError = p } }
func WithCondition(fn func(*StepContext) bool) StepOption {
	return func(c *StepConfig) { c.When = fn }
}

// WithServices registers step-local services (visible only to this step's
// tasks) on top of the flow SP, e.g. swap the API client for one step.
func WithServices(fn func(*Registry)) StepOption {
	return func(c *StepConfig) { c.Overrides = fn }
}

// configured is implemented by steps that carry a StepConfig.
type configured interface{ config() StepConfig }

// ---------------------------------------------------------------------------
// TaskStep: the reusable, general-purpose step. It runs an ordered task
// pipeline with a TaskProvider derived from the flow SP.
// ---------------------------------------------------------------------------

type TaskStep struct {
	name  string
	tasks []Task
	cfg   StepConfig
	route func(tp *TaskProvider) (Result, error) // optional custom result/route
}

// NewStep creates a step from tasks. Steps are values: define once, reuse
// in as many flows as needed.
func NewStep(name string, tasks []Task, opts ...StepOption) *TaskStep {
	s := &TaskStep{name: name, tasks: tasks}
	for _, o := range opts {
		o(&s.cfg)
	}
	return s
}

// Route lets the step decide the next step from its scratch data.
func (s *TaskStep) Route(fn func(tp *TaskProvider) (Result, error)) *TaskStep {
	s.route = fn
	return s
}

// With returns a copy of the step with extra options (reuse with tweaks).
func (s *TaskStep) With(opts ...StepOption) *TaskStep {
	cp := *s
	cp.tasks = append([]Task(nil), s.tasks...)
	for _, o := range opts {
		o(&cp.cfg)
	}
	return &cp
}

// Renamed returns a copy under a new name (to use one step twice in a flow).
func (s *TaskStep) Renamed(name string) *TaskStep {
	cp := s.With()
	cp.name = name
	return cp
}

func (s *TaskStep) Name() string       { return s.name }
func (s *TaskStep) config() StepConfig { return s.cfg }

func (s *TaskStep) Execute(ctx context.Context, sc *StepContext) (Result, error) {
	tp := newTaskProvider(sc.Provider, sc, s.cfg.Overrides)
	for _, t := range s.tasks {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		if err := t.Run(ctx, tp); err != nil {
			return Result{}, fmt.Errorf("task %q: %w", t.Name(), err)
		}
	}
	if s.route != nil {
		return s.route(tp)
	}
	out, _ := tp.Scratch.Get(ScratchResult)
	return Result{Output: out}, nil
}

// ---------------------------------------------------------------------------
// FuncStep: for steps that don't need a task pipeline.
// ---------------------------------------------------------------------------

type FuncStep struct {
	name string
	fn   func(ctx context.Context, sc *StepContext) (Result, error)
	cfg  StepConfig
}

func StepFunc(name string, fn func(ctx context.Context, sc *StepContext) (Result, error), opts ...StepOption) *FuncStep {
	s := &FuncStep{name: name, fn: fn}
	for _, o := range opts {
		o(&s.cfg)
	}
	return s
}

func (s *FuncStep) Name() string       { return s.name }
func (s *FuncStep) config() StepConfig { return s.cfg }
func (s *FuncStep) Execute(ctx context.Context, sc *StepContext) (Result, error) {
	return s.fn(ctx, sc)
}

// ---------------------------------------------------------------------------
// AIStep: convenience constructor for the canonical pipeline
//   API call -> response processing -> tools execution -> result
// ---------------------------------------------------------------------------

type AIStepSpec struct {
	Build   func(tp *TaskProvider) (Request, error)                 // required
	Process func(tp *TaskProvider, resp Response) (Response, error) // optional
	Result  func(tp *TaskProvider) (any, error)                     // optional; default: Response.Text + tool results
	Tools   bool                                                    // run tool calls
}

func AIStep(name string, spec AIStepSpec, opts ...StepOption) *TaskStep {
	tasks := []Task{APICallTask{Build: spec.Build}}
	if spec.Process != nil {
		tasks = append(tasks, ProcessTask{Fn: spec.Process})
	}
	if spec.Tools {
		tasks = append(tasks, ToolsTask{})
	}
	res := spec.Result
	if res == nil {
		res = DefaultResult
	}
	tasks = append(tasks, ResultTask{Fn: res})
	return NewStep(name, tasks, opts...)
}

// AIOutput is the default output of an AIStep.
type AIOutput struct {
	Text        string
	ToolResults []ToolResult
}

func DefaultResult(tp *TaskProvider) (any, error) {
	resp, _ := Get[Response](tp.Scratch, ScratchResponse)
	tr, _ := Get[[]ToolResult](tp.Scratch, ScratchToolResults)
	return AIOutput{Text: resp.Text, ToolResults: tr}, nil
}

package v1

import (
	"context"
	"errors"
	"time"
)

// ---------------------------------------------------------------------
// 1. Shared execution contract
// ---------------------------------------------------------------------

type Executable[In, Out any] interface {
	Execute(ctx context.Context, in In) (Out, error)
}

// ---------------------------------------------------------------------
// 2. Service provider: the only door a step has into the flow
// ---------------------------------------------------------------------

type ServiceProvider interface {
	Data() Data
	Run() RunInfo
	Service(name string) (any, bool)
}

func Resolve[T any](sp ServiceProvider, name string) (T, error) {
	var zero T
	v, ok := sp.Service(name)
	if !ok {
		return zero, ErrServiceNotFound
	}

	t, ok := v.(T)
	if !ok {
		return zero, ErrServiceType
	}

	return t, nil
}

// ---------------------------------------------------------------------
// 3. Flow data
// ---------------------------------------------------------------------

type DataReader interface {
	Get(key string) (any, bool)
	Has(key string) bool
	Keys() []string
}

type DataWriter interface {
	Set(key string, value any)
	Delete(key string)
}

type Data interface {
	DataReader
	DataWriter
}

// Key
// var OrderID = flow.NewKey[string]("order.id")
// id, err := OrderID.Get(sp.Data())
// OrderID.Set(sp.Data(), "A-42")
type Key[T any] struct{ name string }

func NewKey[T any](name string) Key[T] { return Key[T]{name: name} }

func (k Key[T]) Name() string { return k.name }

func (k Key[T]) Get(r DataReader) (T, error) {
	var zero T
	v, ok := r.Get(k.name)
	if !ok {
		return zero, ErrDataNotFound
	}

	t, ok := v.(T)
	if !ok {
		return zero, ErrDataType
	}

	return t, nil
}

func (k Key[T]) Set(w DataWriter, v T) { w.Set(k.name, v) }

type Task[In, Out any] interface {
	Executable[In, Out]
}

type TaskFunc[In, Out any] func(ctx context.Context, in In) (Out, error)

func (f TaskFunc[In, Out]) Execute(ctx context.Context, in In) (Out, error) { return f(ctx, in) }

// ---------------------------------------------------------------------
// 5. Steps: the states of the machine
// ---------------------------------------------------------------------

type StepID string
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

type Result struct {
	Outcome Outcome
	Message string
}

type Step interface {
	ID() StepID
	Executable[ServiceProvider, Result]
}

type TaskStep[In, Out any] interface {
	Step
	Task() Task[In, Out]
	MapInput(sp ServiceProvider) (In, error)
	MapOutput(sp ServiceProvider, out Out) (Result, error)
}

// ---------------------------------------------------------------------
// 6. Transitions: the edges of the machine
// ---------------------------------------------------------------------

type TransitionContext struct {
	From     StepID
	Result   Result
	Services ServiceProvider
}

type Condition interface {
	Satisfied(tc TransitionContext) bool
}

type ConditionFunc func(tc TransitionContext) bool

func (f ConditionFunc) Satisfied(tc TransitionContext) bool { return f(tc) }

type Transition interface {
	From() StepID
	To() StepID
	Condition() Condition
	Priority() int
}

type Resolver interface {
	Resolve(candidates []Transition, tc TransitionContext) (next StepID, ok bool, err error)
}

// ---------------------------------------------------------------------
// 7. Definition (the graph) and Flow (the engine)
// ---------------------------------------------------------------------

type Definition interface {
	Name() string
	Start() StepID
	IsFinal(id StepID) bool
	Step(id StepID) (Step, bool)
	Steps() []Step
	Transitions() []Transition
	TransitionsFrom(id StepID) []Transition
}

type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusCancelled
)

type StepRecord struct {
	Step      StepID
	Result    Result
	Err       error
	StartedAt time.Time
	Duration  time.Duration
}

type RunInfo interface {
	RunID() string
	FlowName() string
	CurrentStep() StepID
	History() []StepRecord
}

type FlowResult struct {
	RunID    string
	Status   Status
	LastStep StepID
	History  []StepRecord
	Err      error
}

type Input interface {
	Apply(w DataWriter) error
}

type Flow interface {
	Executable[Input, FlowResult]
	Definition() Definition
}

// ---------------------------------------------------------------------
// 8. Construction
// ---------------------------------------------------------------------

type Builder interface {
	Name(name string) Builder
	StartAt(id StepID) Builder
	EndAt(ids ...StepID) Builder

	AddStep(s Step) Builder
	AddTransition(from, to StepID, cond Condition, priority int) Builder
	On(from StepID, outcome Outcome, to StepID) Builder // shorthand for outcome matching

	WithResolver(r Resolver) Builder
	WithScopeFactory(f ScopeFactory) Builder
	WithObserver(o Observer) Builder
	WithMiddleware(m ...Middleware) Builder

	BuildDefinition() (Definition, error) // checks the graph
	Build() (Flow, error)
}

type Validator interface {
	Validate(d Definition) []error
}

// ---------------------------------------------------------------------
// 9. Extension points
// ---------------------------------------------------------------------

type ScopeFactory interface {
	NewScope(ctx context.Context, d Definition, runID string) (Scope, error)
}

type Scope interface {
	Services() ServiceProvider
	Close() error
}

type Observer interface {
	FlowStarted(ctx context.Context, run RunInfo)
	StepStarted(ctx context.Context, run RunInfo, step StepID)
	StepCompleted(ctx context.Context, run RunInfo, rec StepRecord)
	Transitioned(ctx context.Context, run RunInfo, from, to StepID)
	FlowCompleted(ctx context.Context, res FlowResult)
}

type StepHandler func(ctx context.Context, step Step, sp ServiceProvider) (Result, error)

type Middleware func(next StepHandler) StepHandler

// ---------------------------------------------------------------------
// 10. Errors
// ---------------------------------------------------------------------

var (
	ErrDataNotFound    = errors.New("flow: data key not found")
	ErrDataType        = errors.New("flow: data value has unexpected type")
	ErrServiceNotFound = errors.New("flow: service not found")
	ErrServiceType     = errors.New("flow: service has unexpected type")
	ErrStepNotFound    = errors.New("flow: step not found")
	ErrAmbiguous       = errors.New("flow: more than one transition matched")
)

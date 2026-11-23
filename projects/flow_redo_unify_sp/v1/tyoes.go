// Package flow defines the contracts of a flexible state machine.
//
// Runtime picture:
//
//	Flow ──(new Scope → ServiceProvider)──► Step.Execute(ctx, sp)
//	                                           │
//	           read input via sp.Data() ◄──────┤
//	                                           │
//	                    in ──► Task.Execute(ctx, in) ──► out
//	                                           │
//	          write output via sp.Data() ◄─────┤
//	                                           │
//	Flow ◄──────────── Result{Outcome} ────────┘
//	  │
//	  └─► Resolver picks the next step from the outgoing transitions ─► loop
//
// The package is interface-first: apart from small value types and typed
// data keys, everything here is a contract to be implemented elsewhere.
package flow

import (
	"context"
	"errors"
	"time"
)

// ---------------------------------------------------------------------
// 1. Shared execution contract
// ---------------------------------------------------------------------

// Executable is anything that can run. Flows, steps and tasks are all
// executables; only their input and output types differ.
type Executable[In, Out any] interface {
	Execute(ctx context.Context, in In) (Out, error)
}

// ---------------------------------------------------------------------
// 2. Service provider: the only door a step has into the flow
// ---------------------------------------------------------------------

// ServiceProvider is handed to every step by the flow. It is scoped to a
// single flow run.
type ServiceProvider interface {
	// Data is the shared flow data of the current run.
	Data() Data

	// Run describes the current run (id, current step, history).
	Run() RunInfo

	// Service looks up any other registered dependency (logger, clients,
	// repositories and so on) by name. Use the generic Resolve helper for
	// typed access.
	Service(name string) (any, bool)
}

// Resolve is a typed helper over ServiceProvider.Service.
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

// DataReader is the read side of the flow data.
type DataReader interface {
	Get(key string) (any, bool)
	Has(key string) bool
	Keys() []string
}

// DataWriter is the write side of the flow data. Keeping it separate lets
// you hand a step read-only access.
type DataWriter interface {
	Set(key string, value any)
	Delete(key string)
}

// Data is full access to the flow data. Implementations must be safe for
// concurrent use if steps run tasks in parallel.
type Data interface {
	DataReader
	DataWriter
}

// Key is a typed handle for one entry in the flow data, so callers never
// have to write type assertions themselves.
//
//	var OrderID = flow.NewKey[string]("order.id")
//	id, err := OrderID.Get(sp.Data())
//	OrderID.Set(sp.Data(), "A-42")
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

// ---------------------------------------------------------------------
// 4. Tasks: pure units of work, no access to the service provider
// ---------------------------------------------------------------------

// Task receives data from its step and returns data to it. It knows
// nothing about the flow, which keeps it reusable and easy to test.
type Task[In, Out any] interface {
	Executable[In, Out]
}

// TaskFunc adapts a plain function to Task.
type TaskFunc[In, Out any] func(ctx context.Context, in In) (Out, error)

func (f TaskFunc[In, Out]) Execute(ctx context.Context, in In) (Out, error) { return f(ctx, in) }

// ---------------------------------------------------------------------
// 5. Steps: the states of the machine
// ---------------------------------------------------------------------

type StepID string

// Outcome is what a step reports. Transitions match on it.
type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeFailure Outcome = "failure"
)

// Result is returned by every step.
type Result struct {
	Outcome Outcome
	Message string
}

// Step is executed with the run's ServiceProvider. Its job:
// read from Data → call its task(s) → write to Data → report an Outcome.
//
// Returning an error means "the step broke" (the flow fails or a middleware
// retries). Returning a failure Outcome means "the step ran and the
// business result was negative", which a transition can route.
type Step interface {
	ID() StepID
	Executable[ServiceProvider, Result]
}

// TaskStep covers the common case of one step wrapping one task. A
// generic adapter can implement Step.Execute on top of these methods.
type TaskStep[In, Out any] interface {
	Step
	Task() Task[In, Out]

	// MapInput builds the task input from flow data.
	MapInput(sp ServiceProvider) (In, error)

	// MapOutput writes the task output back to flow data and picks the outcome.
	MapOutput(sp ServiceProvider, out Out) (Result, error)
}

// ---------------------------------------------------------------------
// 6. Transitions: the edges of the machine
// ---------------------------------------------------------------------

// TransitionContext is what a guard can look at.
type TransitionContext struct {
	From     StepID
	Result   Result
	Services ServiceProvider // guards may inspect flow data
}

// Condition guards a transition.
type Condition interface {
	Satisfied(tc TransitionContext) bool
}

// ConditionFunc adapts a function to Condition.
type ConditionFunc func(tc TransitionContext) bool

func (f ConditionFunc) Satisfied(tc TransitionContext) bool { return f(tc) }

type Transition interface {
	From() StepID
	To() StepID
	Condition() Condition // nil means "always"
	Priority() int        // higher wins when several match
}

// Resolver picks the next step. ok == false ends the flow normally;
// err != nil fails it (for example, a strict resolver on ambiguity).
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

// RunInfo is read-only information about a run, exposed through the
// ServiceProvider.
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

// Input seeds the flow data before the first step runs.
type Input interface {
	Apply(w DataWriter) error
}

// Flow is the engine. For each Execute call it:
//  1. creates a Scope (and so a ServiceProvider) for the run,
//  2. applies the Input to the run's Data,
//  3. loops: run step(sp) → resolve transition → next step,
//  4. stops on a final step, no matching transition, an error or ctx.Done().
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

// Validator checks a definition: missing start step, unreachable steps,
// transitions to unknown steps and so on.
type Validator interface {
	Validate(d Definition) []error
}

// ---------------------------------------------------------------------
// 9. Extension points
// ---------------------------------------------------------------------

// ScopeFactory creates the per-run scope where Data and RunInfo are wired
// in alongside any app services.
type ScopeFactory interface {
	NewScope(ctx context.Context, d Definition, runID string) (Scope, error)
}

type Scope interface {
	Services() ServiceProvider
	Close() error
}

// Observer receives lifecycle events (logging, metrics, tracing, persistence).
type Observer interface {
	FlowStarted(ctx context.Context, run RunInfo)
	StepStarted(ctx context.Context, run RunInfo, step StepID)
	StepCompleted(ctx context.Context, run RunInfo, rec StepRecord)
	Transitioned(ctx context.Context, run RunInfo, from, to StepID)
	FlowCompleted(ctx context.Context, res FlowResult)
}

// StepHandler is the signature a Middleware wraps.
type StepHandler func(ctx context.Context, step Step, sp ServiceProvider) (Result, error)

// Middleware wraps every step execution (retries, timeouts, recover and so on).
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

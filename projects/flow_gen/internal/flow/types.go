package flow

import (
	"context"
	"errors"
	"time"
)

// ---------------------------------------------------------------------
// Contracts
// ---------------------------------------------------------------------
var (
	ErrDataNotFound    = errors.New("flow: data key not found")
	ErrDataType        = errors.New("flow: data value has unexpected type")
	ErrServiceNotFound = errors.New("flow: service not found")
	ErrServiceType     = errors.New("flow: service has unexpected type")
	ErrStepNotFound    = errors.New("flow: step not found")
	ErrAmbiguous       = errors.New("flow: more than one transition matched")
)

type StepID string
type Outcome string

const (
	OutcomeFinal = "final"
)

type Status int

const (
	StatusPending Status = iota
	StatusRunning
	StatusCompleted
	StatusFailed
	StatusCancelled
)

// ---------------------------------------------------------------------
// Common
// ---------------------------------------------------------------------

type ServiceProvider interface {
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

type DataReader interface {
	Get(key string) (any, bool)
	Has(key string) bool
	Keys() []string
}

type DataWriter interface {
	Set(key string, value any)
	Delete(key string)
}

type DataStore interface {
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

type Result struct {
	Outcome Outcome
	Message string
}

type Transition struct {
	Next     StepID
	Attempts int
}

type StepGraph map[StepID]map[Outcome]Transition

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

// ---------------------------------------------------------------------
// Flow
// ---------------------------------------------------------------------

type FlowResult struct {
	RunID    string
	Status   Status
	LastStep StepID
	History  []StepRecord
	Err      error
}

type FlowMachine[S ServiceProvider] interface {
	Run(ctx context.Context) FlowResult
	Info() RunInfo
	Name() string
}

// ---------------------------------------------------------------------
// Step
// ---------------------------------------------------------------------

type FlowStep[S ServiceProvider, D DataStore] interface {
	Execute(ctx context.Context, s S, d D) (Result, error)
}

// ---------------------------------------------------------------------
// Task
// ---------------------------------------------------------------------

type FlowTask[In, Out any] interface {
	Process(ctx context.Context, in In) (Out, error)
}

type FlowTaskFunc[In, Out any] func(ctx context.Context, in In) (Out, error)

func (f FlowTaskFunc[In, Out]) Execute(ctx context.Context, in In) (Out, error) { return f(ctx, in) }

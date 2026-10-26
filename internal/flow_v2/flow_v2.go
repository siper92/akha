package flow_v2

import (
	"context"
	"errors"

	ds "github.com/siper92/akha/internal/data_sources"
)

const End = "END"

type Outcome string

const (
	OutcomeOK     Outcome = "ok"
	OutcomeFollow Outcome = "follow"
)

var (
	ErrNoTransition   = errors.New("no transition")
	ErrUnknownStep    = errors.New("unknown step")
	ErrDuplicateStep  = errors.New("duplicate step")
	ErrMaxTransitions = errors.New("max transitions reached")
	ErrNotInit        = errors.New("not initialised")
	ErrNoResult       = errors.New("no result")
	ErrBadPlan        = errors.New("bad plan")
)

type ExecutionPlan interface {
	Start() string
	Next(id string, outcome Outcome) (string, error)
}

type ServiceProvider interface {
	Dir(ctx context.Context, path string) (ds.DirDataSource, error)
}

type State interface {
	Get(id string) (any, bool)
	Set(id string, v any)
}

type Runtime interface {
	Services() ServiceProvider
	State() State
}

type Flow interface {
	Init(plan ExecutionPlan, steps []Step, sp ServiceProvider) error
	Run(ctx context.Context) error
}

type Step interface {
	ID() string
	Execute(ctx context.Context, rt Runtime) (Outcome, error)
}

type TaskStep[T, R any] interface {
	Step
	Init(tasks []Task[T, R], plan ExecutionPlan) error
	Data(ctx context.Context, rt Runtime) (T, error)
}

type Task[T, R any] interface {
	ID() string
	Init(data T)
	Validate() error
	Process(ctx context.Context) (R, error)
}

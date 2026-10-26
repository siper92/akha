package flow

import (
	"context"
	"errors"
	"fmt"

	"github.com/siper92/akha/internal/data_sources"
)

type ID string

var (
	ErrDuplicateID    = errors.New("duplicate id")
	ErrUnknownStep    = errors.New("unknown step")
	ErrUnknownTask    = errors.New("unknown task")
	ErrNotPrevious    = errors.New("not a previous entry")
	ErrNoResult       = errors.New("no result")
	ErrResultType     = errors.New("result type mismatch")
	ErrUnknownService = errors.New("unknown service")
	ErrServiceType    = errors.New("service type mismatch")
	ErrStepFailed     = errors.New("step failed")
	ErrTaskFailed     = errors.New("task failed")
	ErrDataGet        = errors.New("data get failed")
)

type Plan interface {
	Order() []ID
	Index(id ID) (int, bool)
	Validate() error
}

type Processor[E, R any] interface {
	Process(ctx context.Context, e E) (R, error)
}

type Services interface {
	Service(name string) (any, bool)
}

type DataProvider interface {
	Sources() data_sources.Provider
	Services() Services
}

type Results[R any] interface {
	Result(id ID) (R, bool)
}

type Caller[R any] interface {
	Call(ctx context.Context, id ID) (R, error)
}

type Scope[R any] interface {
	Results[R]
	Caller[R]
}

type StepScope interface {
	DataProvider
	Scope[any]
}

type Flow interface {
	StepScope
	Plan() Plan
	Exec(ctx context.Context) error
}

type Step interface {
	ID() ID
	Process(ctx context.Context, scope StepScope) (any, error)
}

type DataGetter[D any] interface {
	Get(ctx context.Context, scope StepScope) ([]D, error)
}

type Task[D, R any] interface {
	ID() ID
	Run(ctx context.Context, data D, scope Scope[R]) (R, error)
}

type StepRun struct {
	Step  Step
	Scope StepScope
}

type StepProcessor interface {
	Processor[StepRun, any]
}

type TaskRun[D, R any] struct {
	Task  Task[D, R]
	Data  D
	Scope Scope[R]
}

type TaskProcessor[D, R any] interface {
	Processor[TaskRun[D, R], R]
}

func ResultAs[T any](r Results[any], id ID) (T, error) {
	var zero T
	v, ok := r.Result(id)
	if !ok {
		return zero, fmt.Errorf("%w: %s", ErrNoResult, id)
	}
	t, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("%w: %s", ErrResultType, id)
	}
	return t, nil
}

func ServiceAs[T any](s Services, name string) (T, error) {
	var zero T
	v, ok := s.Service(name)
	if !ok {
		return zero, fmt.Errorf("%w: %s", ErrUnknownService, name)
	}
	t, ok := v.(T)
	if !ok {
		return zero, fmt.Errorf("%w: %s", ErrServiceType, name)
	}
	return t, nil
}

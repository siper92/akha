package v3

import (
	"context"
	"errors"

	"github.com/siper92/akha/internal/ai"
)

var (
	ErrFlowState   = errors.New("flow already ran")
	ErrUnknownTool = errors.New("unknown tool")
	ErrBadInput    = errors.New("invalid tool input")
)

type State string

const (
	StateIdle    State = "idle"
	StateRunning State = "running"
	StateDone    State = "done"
	StateFailed  State = "failed"
)

type TaskRecord struct {
	ID          string
	Title       string
	Description string
}

type TaskResult struct {
	TaskID string
	Output string
}

type TaskIn struct {
	Record TaskRecord
	Done   []TaskResult
}

type Data struct {
	Tasks   []TaskRecord
	Results []TaskResult
}

type Task func(ctx context.Context, sp TaskServices, in TaskIn) (TaskResult, error)

type FS interface {
	CreateFile(ctx context.Context, fileName string, content string) error
	ReadFile(ctx context.Context, fileName string) (string, error)
	DeleteFile(ctx context.Context, fileName string) error
}

type Debug interface {
	Log(format string, args ...any)
}

type Asker interface {
	Ask(ctx context.Context, prompt string, tools ai.ToolRunner) (string, error)
}

type TaskRepository interface {
	List(ctx context.Context) ([]TaskRecord, error)
}

type Store[D any] interface {
	Data() D
	Update(fn func(d D) D)
}

type TaskServices interface {
	DebugService() Debug
	AI() Asker
	Tools() ai.ToolRunner
}

type FlowServices interface {
	DebugService() Debug
	Tasks() TaskRepository
	TaskServices() TaskServices
}

type Step[D any] interface {
	Name() string
	Exec(ctx context.Context, sp FlowServices, store Store[D]) error
}

type Flow interface {
	Run(ctx context.Context) error
	State() State
}

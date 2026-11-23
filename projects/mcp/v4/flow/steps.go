package flow

import (
	"context"

	"aiflow/debug"
)

const (
	StepPick    = "pick"
	StepExecute = "execute"
	StepRecord  = "record"
)

type State struct {
	Current   TaskRecord
	Output    TaskOutput
	Processed []TaskRecord
}

type PickStep struct{}

type ExecuteStep struct {
	task Task[TaskServiceProvider, TaskInput, TaskOutput]
}

type RecordStep struct{}

var (
	_ Step[ServiceProvider, *State] = (*PickStep)(nil)
	_ Step[ServiceProvider, *State] = (*ExecuteStep)(nil)
	_ Step[ServiceProvider, *State] = (*RecordStep)(nil)
)

func NewTasker(sp ServiceProvider) (*Flow[ServiceProvider, *State], error) {
	return New[ServiceProvider, *State](sp, PickStep{}, ExecuteStep{task: AITask{}}, RecordStep{})
}

func (PickStep) Name() string { return StepPick }

func (PickStep) Exec(ctx context.Context, sp ServiceProvider, data *State) (Transition, error) {
	record, ok, err := sp.Repository().NextPending(ctx)
	if err != nil {
		return Transition{}, err
	}
	if !ok {
		return End(), nil
	}
	data.Current = record
	data.Output = TaskOutput{}
	debug.Service().Log("task %s: %s", record.ID, record.Title)
	return Next(StepExecute), nil
}

func (ExecuteStep) Name() string { return StepExecute }

func (s ExecuteStep) Exec(ctx context.Context, sp ServiceProvider, data *State) (Transition, error) {
	out, err := s.task.Exec(ctx, sp.Tasks(), TaskInput{
		Title:       data.Current.Title,
		Description: data.Current.Description,
	})
	if err != nil {
		data.Current.Status = StatusFailed
		data.Current.Result = err.Error()
		return Next(StepRecord), nil
	}
	data.Output = out
	data.Current.Status = StatusDone
	data.Current.Result = out.Text
	return Next(StepRecord), nil
}

func (RecordStep) Name() string { return StepRecord }

func (RecordStep) Exec(ctx context.Context, sp ServiceProvider, data *State) (Transition, error) {
	if err := sp.Repository().Save(ctx, data.Current); err != nil {
		return Transition{}, err
	}
	data.Processed = append(data.Processed, data.Current)
	debug.Service().Log("task %s: %s after %d tool calls: %s",
		data.Current.ID, data.Current.Status, data.Output.ToolCalls, data.Current.Result)
	if data.Current.Status == StatusFailed {
		return End(), nil
	}
	return Next(StepPick), nil
}

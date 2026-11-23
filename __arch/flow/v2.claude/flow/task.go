package flow

import (
	"context"
	"fmt"
)

// Task is the smallest unit of work. It receives everything it needs via
// the TaskProvider (services + scratch data) and writes its output back
// into tp.Scratch for the next task.
type Task interface {
	Name() string
	Run(ctx context.Context, tp *TaskProvider) error
}

// TaskFunc adapts a function into a Task.
type TaskFunc struct {
	TaskName string
	Fn       func(ctx context.Context, tp *TaskProvider) error
}

func (t TaskFunc) Name() string                                    { return t.TaskName }
func (t TaskFunc) Run(ctx context.Context, tp *TaskProvider) error { return t.Fn(ctx, tp) }

// ---------------------------------------------------------------------------
// Built-in tasks implementing:  API call -> response processing -> tools -> result
// Each is generic; behaviour is customised via small function fields.
// ---------------------------------------------------------------------------

// APICallTask builds a Request and calls the API from the task SP.
type APICallTask struct {
	Build func(tp *TaskProvider) (Request, error)
}

func (APICallTask) Name() string { return "api_call" }

func (t APICallTask) Run(ctx context.Context, tp *TaskProvider) error {
	api, err := tp.API()
	if err != nil {
		return err
	}
	req, err := t.Build(tp)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	tp.Scratch.Set(ScratchRequest, req)
	resp, err := api.Call(ctx, req)
	if err != nil {
		return fmt.Errorf("api call: %w", err)
	}
	tp.Scratch.Set(ScratchResponse, resp)
	return nil
}

// ProcessTask inspects/transforms the API response (validation, parsing,
// filtering tool calls...). Optional in a pipeline.
type ProcessTask struct {
	Fn func(tp *TaskProvider, resp Response) (Response, error)
}

func (ProcessTask) Name() string { return "process_response" }

func (t ProcessTask) Run(_ context.Context, tp *TaskProvider) error {
	resp, ok := Get[Response](tp.Scratch, ScratchResponse)
	if !ok {
		return fmt.Errorf("no response in scratch")
	}
	out, err := t.Fn(tp, resp)
	if err != nil {
		return err
	}
	tp.Scratch.Set(ScratchResponse, out)
	return nil
}

// ToolResult records one executed tool call.
type ToolResult struct {
	Call   ToolCall
	Output any
	Err    error
}

// ToolsTask executes every tool call in the response using tools from the
// task SP. StopOnError controls whether one failing tool fails the task.
type ToolsTask struct {
	StopOnError bool
}

func (ToolsTask) Name() string { return "execute_tools" }

func (t ToolsTask) Run(ctx context.Context, tp *TaskProvider) error {
	resp, _ := Get[Response](tp.Scratch, ScratchResponse)
	results := make([]ToolResult, 0, len(resp.ToolCalls))
	for _, call := range resp.ToolCalls {
		tool, err := tp.Tool(call.Name)
		if err == nil {
			var out any
			out, err = tool.Run(ctx, call.Args)
			results = append(results, ToolResult{Call: call, Output: out, Err: err})
		} else {
			results = append(results, ToolResult{Call: call, Err: err})
		}
		if err != nil && t.StopOnError {
			tp.Scratch.Set(ScratchToolResults, results)
			return fmt.Errorf("tool %q: %w", call.Name, err)
		}
	}
	tp.Scratch.Set(ScratchToolResults, results)
	return nil
}

// ResultTask turns scratch into the step's output value.
type ResultTask struct {
	Fn func(tp *TaskProvider) (any, error)
}

func (ResultTask) Name() string { return "build_result" }

func (t ResultTask) Run(_ context.Context, tp *TaskProvider) error {
	out, err := t.Fn(tp)
	if err != nil {
		return err
	}
	tp.Scratch.Set(ScratchResult, out)
	return nil
}

package flow

import (
	"context"
	"fmt"

	"aiflow/ai"
)

const taskSystemPrompt = "You complete one task at a time using the file tools. " +
	"Files are addressed by plain relative names. " +
	"When the task is complete, reply with one short sentence describing what you did."

type Task[SP any, In any, Out any] interface {
	Exec(ctx context.Context, sp SP, in In) (Out, error)
}

type TaskInput struct {
	Title       string
	Description string
}

type TaskOutput struct {
	Text      string
	ToolCalls int
}

type AITask struct{}

var _ Task[TaskServiceProvider, TaskInput, TaskOutput] = (*AITask)(nil)

func (AITask) Exec(ctx context.Context, sp TaskServiceProvider, in TaskInput) (TaskOutput, error) {
	res, err := sp.AI().Run(ctx, ai.Request{
		System: taskSystemPrompt,
		Prompt: fmt.Sprintf("Task: %s\n\n%s", in.Title, in.Description),
		Tools:  sp.Tools(),
	})
	if err != nil {
		return TaskOutput{}, err
	}
	return TaskOutput{Text: res.Text, ToolCalls: res.ToolCalls}, nil
}

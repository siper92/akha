package v3

import (
	"context"
	"fmt"
	"strings"
)

var _ Task = AITask

func AITask(ctx context.Context, sp TaskServices, in TaskIn) (TaskResult, error) {
	return aiTask{sp: sp, in: in}.exec(ctx)
}

type aiTask struct {
	sp TaskServices
	in TaskIn
}

func (t aiTask) exec(ctx context.Context) (TaskResult, error) {
	debug := t.sp.DebugService()
	debug.Log("task %s: %s", t.in.Record.ID, t.in.Record.Title)

	out, err := t.sp.AI().Ask(ctx, t.prompt(), t.sp.Tools())
	if err != nil {
		return TaskResult{}, err
	}
	debug.Log("task %s: %s", t.in.Record.ID, out)

	return TaskResult{TaskID: t.in.Record.ID, Output: out}, nil
}

func (t aiTask) prompt() string {
	var b strings.Builder
	fmt.Fprintf(&b, "task %s: %s\n%s\n", t.in.Record.ID, t.in.Record.Title, t.in.Record.Description)
	b.WriteString("do all the work with the file tools, call as many as the task needs, then reply with one line on what you did\n")
	if len(t.in.Done) > 0 {
		b.WriteString("completed tasks:\n")
	}
	for _, done := range t.in.Done {
		fmt.Fprintf(&b, "- %s: %s\n", done.TaskID, done.Output)
	}

	return b.String()
}

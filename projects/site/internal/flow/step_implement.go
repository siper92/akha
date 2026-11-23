package flow

import (
	"context"
	"errors"
	"fmt"

	flow "github.com/siper92/akha/internal/flow_v3"
	"github.com/siper92/akha/projects/site/internal/db"
	dbsdk "github.com/siper92/akha/projects/site/internal/sdk/db-sdk"
)

const implementGuide = "use the createFile tool to write every output file of the task, use readFile to read files written by earlier tasks, then reply with a short summary"

var _ siteStepRunner = implementStep{}

type implementStep struct {
	name     flow.StepName
	out      OutputConfig
	attempts int64
	via      AskVia
}

func (i implementStep) Name() flow.StepName {
	return i.name
}

func (i implementStep) run(ctx context.Context, s SiteService) (flow.Transition, error) {
	d := s.Data()
	if d.Plan == "" || len(d.StoryIDs) == 0 {
		return flow.UnknownTransition, ErrMissingData
	}

	t, ok, err := nextTask(ctx, s.Tasks(), d.StoryIDs)
	if err != nil {
		return flow.ErrorTransition, err
	}

	if !ok {
		return flow.DoneTransition, nil
	}

	return i.implement(ctx, s, d.Plan, t)
}

func (i implementStep) implement(ctx context.Context, s SiteService, plan string, t dbsdk.FlowTask) (flow.Transition, error) {
	attempt := t.Attempts + 1
	if err := setTaskStatus(ctx, s.Tasks(), t.ID, db.TaskStatusInProgress, attempt, ""); err != nil {
		return flow.UnknownTransition, err
	}

	key := fmt.Sprintf("%s_%s_%d", i.name, t.TaskKey, t.ID)
	prompt := func() (string, error) { return taskPrompt(plan, t), nil }

	_, err := cachedAsk(ctx, s.Cache(), key, i.out, prompt, askWith(i.via, viaAgent, s))
	if err == nil {
		if err = setTaskStatus(ctx, s.Tasks(), t.ID, db.TaskStatusCompleted, attempt, ""); err != nil {
			return flow.UnknownTransition, err
		}

		return TaskDoneTransition, nil
	}

	status := db.TaskStatusInProgress
	if attempt >= i.attempts {
		status, err = db.TaskStatusFailed, fmt.Errorf("%w: %w", ErrTaskFailed, err)
	}

	if uerr := setTaskStatus(ctx, s.Tasks(), t.ID, status, attempt, err.Error()); uerr != nil {
		return flow.UnknownTransition, errors.Join(err, uerr)
	}

	return flow.ErrorTransition, fmt.Errorf("task %q: %w", t.TaskKey, err)
}

func nextTask(ctx context.Context, repo db.TaskRepository, storyIDs []int64) (dbsdk.FlowTask, bool, error) {
	for _, storyID := range storyIDs {
		tasks, err := repo.ListTasks(ctx, storyID)
		if err != nil {
			return dbsdk.FlowTask{}, false, fmt.Errorf("list tasks: %w", err)
		}

		for _, t := range tasks {
			switch db.TaskStatus(t.Status) {
			case db.TaskStatusCompleted, db.TaskStatusFailed:
				continue
			}

			return t, true, nil
		}
	}

	return dbsdk.FlowTask{}, false, nil
}

func setTaskStatus(ctx context.Context, repo db.TaskRepository, id int64, status db.TaskStatus, attempt int64, msg string) error {
	if err := repo.UpdateTaskStatus(ctx, id, status, attempt, msg); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	return nil
}

func taskPrompt(plan string, t dbsdk.FlowTask) string {
	return plan + "\n\n# task " + t.TaskKey + ": " + t.Label + "\n\n" + t.Content + "\n\n" + implementGuide
}

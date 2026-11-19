package site

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/siper92/akha/internal/parse"
	"github.com/siper92/akha/projects/site/sdk/db"
)

const (
	tasksDocID      = "tasks"
	tasksDesignFlow = "site"
)

var ErrNoTasks = errors.New("no tasks in answer")

func parseTasks(answer string) ([]db.Task, error) {
	doc, err := parse.Parse("# @" + parse.ID_SECTION_KEY + ": " + tasksDocID + "\n" + answer)
	if err != nil {
		return nil, fmt.Errorf("parse tasks: %w", err)
	}

	var tasks []db.Task
	for _, sec := range doc.Content() {
		label := sec.Title().ID()
		if label == parse.ID_SECTION_KEY {
			continue
		}

		tasks = append(tasks, db.Task{
			Label:   label,
			Content: strings.TrimSpace(sec.Content().String()),
		})
	}

	if len(tasks) == 0 {
		return nil, ErrNoTasks
	}

	return tasks, nil
}

func saveTasks(ctx context.Context, s SiteService, answer string) (int64, error) {
	tasks, err := parseTasks(answer)
	if err != nil {
		return 0, err
	}

	storyID, err := s.Tasks().CreateStory(ctx, tasksDesignFlow, string(TaskStepName))
	if err != nil {
		return 0, fmt.Errorf("create story: %w", err)
	}

	if err = s.Tasks().SaveTasks(ctx, storyID, tasks); err != nil {
		return 0, fmt.Errorf("save tasks: %w", err)
	}

	return storyID, nil
}

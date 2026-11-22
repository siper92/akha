package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/siper92/akha/projects/site/internal/db"
)

const tasksDesignFlow = "site"

const tasksSchema = `{
    "stories": [
        {
            "story_id": "<story_id>",
            "story_description": "<story_description>",
            "tasks": [
                {
                "task_id": "<task_id>",
                "task_name": "<task_name>",
                "task_description": "<task_description>"
                }
            ]
        }
    ]
}`

var (
	ErrNoTasks      = errors.New("no tasks in answer")
	ErrMissingField = errors.New("missing field in answer")
)

type TaskPlan struct {
	Stories []TaskStory `json:"stories"`
}

type TaskStory struct {
	StoryID          string     `json:"story_id"`
	StoryDescription string     `json:"story_description"`
	Tasks            []TaskItem `json:"tasks"`
}

type TaskItem struct {
	TaskID          string `json:"task_id"`
	TaskName        string `json:"task_name"`
	TaskDescription string `json:"task_description"`
}

var tasksOutput = OutputConfig{
	Format: OutputJSON,
	Schema: tasksSchema,
	Validate: func(answer string) error {
		_, err := parseTasks(answer)
		return err
	},
}

func parseTasks(answer string) (TaskPlan, error) {
	var plan TaskPlan
	dec := json.NewDecoder(strings.NewReader(trimJSONFence(answer)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&plan); err != nil {
		return TaskPlan{}, fmt.Errorf("parse tasks: %w", err)
	}

	if len(plan.Stories) == 0 {
		return TaskPlan{}, ErrNoTasks
	}

	for _, st := range plan.Stories {
		if st.StoryID == "" {
			return TaskPlan{}, fmt.Errorf("%w: story_id", ErrMissingField)
		}

		if len(st.Tasks) == 0 {
			return TaskPlan{}, fmt.Errorf("story %q: %w", st.StoryID, ErrNoTasks)
		}

		for _, t := range st.Tasks {
			if t.TaskID == "" || t.TaskName == "" {
				return TaskPlan{}, fmt.Errorf("story %q: %w: task_id or task_name", st.StoryID, ErrMissingField)
			}
		}
	}

	return plan, nil
}

func saveTasks(ctx context.Context, s SiteService, plan TaskPlan) ([]int64, error) {
	stories := make([]db.Story, 0, len(plan.Stories))
	for _, st := range plan.Stories {
		tasks := make([]db.Task, 0, len(st.Tasks))
		for _, t := range st.Tasks {
			tasks = append(tasks, db.Task{
				Key:     t.TaskID,
				Label:   t.TaskName,
				Content: t.TaskDescription,
				Status:  db.TaskStatusPending,
			})
		}

		stories = append(stories, db.Story{
			Key:         st.StoryID,
			Description: st.StoryDescription,
			Tasks:       tasks,
		})
	}

	ids, err := s.Tasks().SaveStories(ctx, tasksDesignFlow, string(TaskStepName), stories)
	if err != nil {
		return nil, fmt.Errorf("save stories: %w", err)
	}

	return ids, nil
}

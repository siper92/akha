package db

import (
	"context"
	"database/sql"
	"fmt"

	dbsdk "github.com/siper92/akha/projects/site/sdk/db-sdk"
)

type Story struct {
	Key         string
	Description string
	Tasks       []Task
}

type Task struct {
	Key     string
	Label   string
	Content string
}

type TaskRepository interface {
	SaveStories(ctx context.Context, designFlow, filePath string, stories []Story) ([]int64, error)
	ListTasks(ctx context.Context, storyID int64) ([]dbsdk.FlowTask, error)
}

var _ TaskRepository = (*Repository)(nil)

type Repository struct {
	conn *sql.DB
	q    *dbsdk.Queries
}

func NewRepository(conn *sql.DB) *Repository {
	return &Repository{conn: conn, q: dbsdk.New(conn)}
}

func (r *Repository) SaveStories(ctx context.Context, designFlow, filePath string, stories []Story) ([]int64, error) {
	ids := make([]int64, 0, len(stories))
	err := r.tx(ctx, func(q *dbsdk.Queries) error {
		projectID, err := q.InsertFlowProject(ctx, designFlow)
		if err != nil {
			return fmt.Errorf("insert project: %w", err)
		}

		runID, err := q.InsertFlowRun(ctx, projectID)
		if err != nil {
			return fmt.Errorf("insert run: %w", err)
		}

		for _, st := range stories {
			storyID, err := q.InsertFlowStory(ctx, dbsdk.InsertFlowStoryParams{
				RunID:       runID,
				FilePath:    filePath,
				StoryKey:    st.Key,
				Description: st.Description,
			})
			if err != nil {
				return fmt.Errorf("insert story %q: %w", st.Key, err)
			}

			for i, t := range st.Tasks {
				if _, err = q.InsertFlowTask(ctx, dbsdk.InsertFlowTaskParams{
					StoryID:  storyID,
					Position: int64(i),
					TaskKey:  t.Key,
					Label:    t.Label,
					Content:  t.Content,
				}); err != nil {
					return fmt.Errorf("insert task %q: %w", t.Key, err)
				}
			}

			ids = append(ids, storyID)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *Repository) ListTasks(ctx context.Context, storyID int64) ([]dbsdk.FlowTask, error) {
	return r.q.ListFlowTasksByStory(ctx, storyID)
}

func (r *Repository) tx(ctx context.Context, fn func(q *dbsdk.Queries) error) error {
	tx, err := r.conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	if err = fn(r.q.WithTx(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

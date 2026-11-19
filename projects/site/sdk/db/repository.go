package db

import (
	"context"
	"database/sql"
	"fmt"

	dbsdk "github.com/siper92/akha/projects/site/sdk/db-sdk"
)

type Task struct {
	Label   string
	Content string
}

type TaskRepository interface {
	CreateStory(ctx context.Context, designFlow, filePath string) (int64, error)
	SaveTasks(ctx context.Context, storyID int64, tasks []Task) error
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

func (r *Repository) CreateStory(ctx context.Context, designFlow, filePath string) (int64, error) {
	var storyID int64
	err := r.tx(ctx, func(q *dbsdk.Queries) error {
		projectID, err := q.InsertFlowProject(ctx, designFlow)
		if err != nil {
			return fmt.Errorf("insert project: %w", err)
		}

		runID, err := q.InsertFlowRun(ctx, projectID)
		if err != nil {
			return fmt.Errorf("insert run: %w", err)
		}

		storyID, err = q.InsertFlowStory(ctx, dbsdk.InsertFlowStoryParams{
			RunID:    runID,
			FilePath: filePath,
		})
		if err != nil {
			return fmt.Errorf("insert story: %w", err)
		}

		return nil
	})

	return storyID, err
}

func (r *Repository) SaveTasks(ctx context.Context, storyID int64, tasks []Task) error {
	return r.tx(ctx, func(q *dbsdk.Queries) error {
		for i, t := range tasks {
			if _, err := q.InsertFlowTask(ctx, dbsdk.InsertFlowTaskParams{
				StoryID:  storyID,
				Position: int64(i),
				Label:    t.Label,
				Content:  t.Content,
			}); err != nil {
				return fmt.Errorf("insert task %q: %w", t.Label, err)
			}
		}

		return nil
	})
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

-- name: SelectFlowStories :exec
SELECT * FROM flow_stories;

-- name: InsertFlowProject :one
INSERT INTO flow_projects (design_flow) VALUES (?)
RETURNING id;

-- name: InsertFlowRun :one
INSERT INTO flow_runs (project_id) VALUES (?)
RETURNING id;

-- name: InsertFlowStory :one
INSERT INTO flow_stories (run_id, file_path, story_key, description) VALUES (?, ?, ?, ?)
RETURNING id;

-- name: InsertFlowTask :one
INSERT INTO flow_tasks (story_id, position, task_key, status, label, content) VALUES (?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateFlowTaskStatus :exec
UPDATE flow_tasks
SET status = ?, attempts = ?, error = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: ListFlowTasksByStory :many
SELECT * FROM flow_tasks
WHERE story_id = ?
ORDER BY position;

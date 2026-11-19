-- name: SelectFlowStories :exec
SELECT * FROM flow_stories;

-- name: InsertFlowProject :one
INSERT INTO flow_projects (design_flow) VALUES (?)
RETURNING id;

-- name: InsertFlowRun :one
INSERT INTO flow_runs (project_id) VALUES (?)
RETURNING id;

-- name: InsertFlowStory :one
INSERT INTO flow_stories (run_id, file_path) VALUES (?, ?)
RETURNING id;

-- name: InsertFlowTask :one
INSERT INTO flow_tasks (story_id, position, label, content) VALUES (?, ?, ?, ?)
RETURNING id;

-- name: ListFlowTasksByStory :many
SELECT * FROM flow_tasks
WHERE story_id = ?
ORDER BY position;

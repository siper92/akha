-- name: CreateFlowRun :one
INSERT INTO flow_runs (name, input, created_at, updated_at)
VALUES (sqlc.arg(name), sqlc.arg(input), CAST(sqlc.arg(now) AS INTEGER), CAST(sqlc.arg(now) AS INTEGER))
RETURNING *;

-- name: GetFlowRun :one
SELECT * FROM flow_runs WHERE id = sqlc.arg(id);

-- name: ListFlowRunsByStatus :many
SELECT * FROM flow_runs
WHERE status = sqlc.arg(status)
ORDER BY created_at DESC
LIMIT sqlc.arg(lim);

-- FlowRunProgress: task counts per state, used to decide whether a flow is done.
-- name: FlowRunProgress :one
SELECT
    CAST(COUNT(*) AS INTEGER)                                                   AS total,
    CAST(COALESCE(SUM(status = 'succeeded'), 0) AS INTEGER)                     AS succeeded,
    CAST(COALESCE(SUM(status = 'dead'), 0) AS INTEGER)                          AS dead,
    CAST(COALESCE(SUM(status = 'canceled'), 0) AS INTEGER)                      AS canceled,
    CAST(COALESCE(SUM(status IN ('pending', 'running', 'retry')), 0) AS INTEGER) AS active
FROM tasks
WHERE flow_run_id = sqlc.arg(flow_run_id);

-- name: FinishFlowRun :execrows
UPDATE flow_runs
SET status      = sqlc.arg(status),
    version     = version + 1,
    updated_at  = CAST(sqlc.arg(now) AS INTEGER),
    finished_at = CAST(sqlc.arg(now) AS INTEGER)
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(version)
  AND status = 'running';

-- =====================================================================
-- Enqueue
-- =====================================================================

-- EnqueueTask: idempotent when dedupe_key is set. If a task with the same
-- (queue, dedupe_key) already exists, nothing is inserted and the call
-- returns sql.ErrNoRows - fetch the existing one with GetTaskByDedupeKey.
-- name: EnqueueTask :one
INSERT INTO tasks (
    queue, kind, flow_run_id, step_name, step_index,
    resource, payload, priority, run_at,
    max_attempts, backoff_base_ms, backoff_max_ms, timeout_ms,
    dedupe_key, created_at, updated_at
) VALUES (
    sqlc.arg(queue), sqlc.arg(kind), sqlc.narg(flow_run_id), sqlc.narg(step_name), sqlc.narg(step_index),
    sqlc.arg(resource), sqlc.arg(payload), sqlc.arg(priority), sqlc.arg(run_at),
    sqlc.arg(max_attempts), sqlc.arg(backoff_base_ms), sqlc.arg(backoff_max_ms), sqlc.arg(timeout_ms),
    sqlc.narg(dedupe_key), CAST(sqlc.arg(now) AS INTEGER), CAST(sqlc.arg(now) AS INTEGER)
)
ON CONFLICT (queue, dedupe_key) WHERE dedupe_key IS NOT NULL DO NOTHING
RETURNING *;

-- name: AddTaskDependency :exec
INSERT INTO task_dependencies (task_id, depends_on_id)
VALUES (sqlc.arg(task_id), sqlc.arg(depends_on_id))
ON CONFLICT DO NOTHING;

-- =====================================================================
-- Worker loop: claim -> (heartbeat)* -> complete | fail
-- Run ClaimTask + StartAttempt in one IMMEDIATE transaction.
-- =====================================================================

-- ClaimTask: atomically take the best claimable task in a queue.
-- Claimable = pending/retry, run_at reached, all dependencies succeeded.
-- Returns sql.ErrNoRows when the queue is empty.
-- name: ClaimTask :one
UPDATE tasks
SET status       = 'running',
    attempt      = attempt + 1,
    locked_by    = CAST(sqlc.arg(worker_id) AS TEXT),
    locked_until = CAST(sqlc.arg(now) AS INTEGER) + timeout_ms,
    version      = version + 1,
    started_at   = COALESCE(started_at, CAST(sqlc.arg(now) AS INTEGER)),
    updated_at   = CAST(sqlc.arg(now) AS INTEGER)
WHERE id = (
    SELECT t.id
    FROM tasks t
    WHERE t.queue = sqlc.arg(queue)
      AND t.status IN ('pending', 'retry')
      AND t.run_at <= CAST(sqlc.arg(now) AS INTEGER)
      AND NOT EXISTS (
          SELECT 1
          FROM task_dependencies d
          JOIN tasks p ON p.id = d.depends_on_id
          WHERE d.task_id = t.id
            AND p.status <> 'succeeded'
      )
    ORDER BY t.priority DESC, t.run_at, t.id
    LIMIT 1
)
  AND status IN ('pending', 'retry')
RETURNING *;

-- name: StartAttempt :one
INSERT INTO task_attempts (task_id, attempt, worker_id, started_at)
VALUES (sqlc.arg(task_id), sqlc.arg(attempt), CAST(sqlc.arg(worker_id) AS TEXT), CAST(sqlc.arg(now) AS INTEGER))
RETURNING *;

-- HeartbeatTask: extend the lease for long-ish tasks. 0 rows = lease lost,
-- the worker must stop and must not report a result.
-- name: HeartbeatTask :execrows
UPDATE tasks
SET locked_until = CAST(sqlc.arg(now) AS INTEGER) + timeout_ms,
    updated_at   = CAST(sqlc.arg(now) AS INTEGER)
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(version)
  AND locked_by = CAST(sqlc.arg(worker_id) AS TEXT)
  AND status = 'running';

-- CompleteTask: 0 rows = lease lost / task was changed by someone else.
-- name: CompleteTask :execrows
UPDATE tasks
SET status       = 'succeeded',
    result       = sqlc.narg(result),
    last_error   = NULL,
    locked_by    = NULL,
    locked_until = NULL,
    version      = version + 1,
    updated_at   = CAST(sqlc.arg(now) AS INTEGER),
    finished_at  = CAST(sqlc.arg(now) AS INTEGER)
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(version)
  AND locked_by = CAST(sqlc.arg(worker_id) AS TEXT)
  AND status = 'running';

-- FailTask: record an error and schedule the retry with exponential backoff
--   delay = min(backoff_base_ms * 2^(attempt-1), backoff_max_ms) + jitter_ms
-- Goes to 'dead' when attempts are exhausted or permanent = 1 (non-retryable
-- error). Pass jitter_ms from Go (e.g. rand in [0, delay/5]) to avoid herds.
-- name: FailTask :one
UPDATE tasks
SET status = CASE
        WHEN CAST(sqlc.arg(permanent) AS BOOLEAN) OR attempt >= max_attempts THEN 'dead'
        ELSE 'retry'
    END,
    run_at = CASE
        WHEN CAST(sqlc.arg(permanent) AS BOOLEAN) OR attempt >= max_attempts THEN run_at
        ELSE CAST(sqlc.arg(now) AS INTEGER)
             + MIN(backoff_base_ms * (1 << MIN(attempt - 1, 30)), backoff_max_ms)
             + sqlc.arg(jitter_ms)
    END,
    finished_at = CASE
        WHEN CAST(sqlc.arg(permanent) AS BOOLEAN) OR attempt >= max_attempts THEN CAST(sqlc.arg(now) AS INTEGER)
        ELSE NULL
    END,
    last_error   = sqlc.arg(error),
    locked_by    = NULL,
    locked_until = NULL,
    version      = version + 1,
    updated_at   = CAST(sqlc.arg(now) AS INTEGER)
WHERE id = sqlc.arg(id)
  AND version = sqlc.arg(version)
  AND locked_by = CAST(sqlc.arg(worker_id) AS TEXT)
  AND status = 'running'
RETURNING id, status, attempt, max_attempts, run_at, version;

-- name: FinishAttempt :exec
UPDATE task_attempts
SET status      = sqlc.arg(status),
    error       = sqlc.narg(error),
    result      = sqlc.narg(result),
    finished_at = CAST(sqlc.arg(now) AS INTEGER)
WHERE task_id = sqlc.arg(task_id)
  AND attempt = sqlc.arg(attempt)
  AND status = 'running';

-- =====================================================================
-- Crash recovery (run periodically, e.g. every few seconds, by one reaper)
-- =====================================================================

-- ReapExpiredLeases: running tasks whose worker vanished. They go back to
-- retry (claimable now) or to dead when out of attempts. Follow with
-- TimeoutAttempts for each returned row, in the same transaction.
-- name: ReapExpiredLeases :many
UPDATE tasks
SET status       = CASE WHEN attempt >= max_attempts THEN 'dead' ELSE 'retry' END,
    run_at       = CAST(sqlc.arg(now) AS INTEGER),
    finished_at  = CASE WHEN attempt >= max_attempts THEN CAST(sqlc.arg(now) AS INTEGER) ELSE NULL END,
    last_error   = 'lease expired (worker ' || locked_by || ')',
    locked_by    = NULL,
    locked_until = NULL,
    version      = version + 1,
    updated_at   = CAST(sqlc.arg(now) AS INTEGER)
WHERE status = 'running'
  AND locked_until < CAST(sqlc.arg(now) AS INTEGER)
RETURNING id, attempt, status;

-- name: TimeoutAttempt :exec
UPDATE task_attempts
SET status      = 'timed_out',
    error       = 'lease expired',
    finished_at = CAST(sqlc.arg(now) AS INTEGER)
WHERE task_id = sqlc.arg(task_id)
  AND attempt = sqlc.arg(attempt)
  AND status = 'running';

-- =====================================================================
-- Control
-- =====================================================================

-- name: CancelTask :execrows
UPDATE tasks
SET status       = 'canceled',
    locked_by    = NULL,
    locked_until = NULL,
    version      = version + 1,
    updated_at   = CAST(sqlc.arg(now) AS INTEGER),
    finished_at  = CAST(sqlc.arg(now) AS INTEGER)
WHERE id = sqlc.arg(id)
  AND status IN ('pending', 'retry', 'running');

-- ListDownstreamTaskIDs: every task that (transitively) depends on task_id.
-- When a step dies, pass the result to CancelTasksByIDs, otherwise those
-- tasks wait forever.
-- name: ListDownstreamTaskIDs :many
WITH RECURSIVE downstream(id) AS (
    SELECT d.task_id FROM task_dependencies d WHERE d.depends_on_id = sqlc.arg(task_id)
    UNION
    SELECT d.task_id FROM task_dependencies d JOIN downstream ds ON d.depends_on_id = ds.id
)
SELECT CAST(id AS INTEGER) AS id FROM downstream;

-- name: CancelTasksByIDs :execrows
UPDATE tasks
SET status      = 'canceled',
    last_error  = CAST(sqlc.arg(reason) AS TEXT),
    version     = version + 1,
    updated_at  = CAST(sqlc.arg(now) AS INTEGER),
    finished_at = CAST(sqlc.arg(now) AS INTEGER)
WHERE status IN ('pending', 'retry')
  AND id IN (sqlc.slice(ids));

-- name: CancelFlowRunTasks :execrows
UPDATE tasks
SET status       = 'canceled',
    locked_by    = NULL,
    locked_until = NULL,
    version      = version + 1,
    updated_at   = CAST(sqlc.arg(now) AS INTEGER),
    finished_at  = CAST(sqlc.arg(now) AS INTEGER)
WHERE flow_run_id = sqlc.arg(flow_run_id)
  AND status IN ('pending', 'retry', 'running');

-- RequeueTask: manual retry of a dead/canceled task. Attempt history is kept;
-- the budget is extended by extra_attempts instead of resetting the counter.
-- name: RequeueTask :execrows
UPDATE tasks
SET status       = 'pending',
    max_attempts = attempt + sqlc.arg(extra_attempts),
    run_at       = CAST(sqlc.arg(now) AS INTEGER),
    last_error   = NULL,
    finished_at  = NULL,
    version      = version + 1,
    updated_at   = CAST(sqlc.arg(now) AS INTEGER)
WHERE id = sqlc.arg(id)
  AND status IN ('dead', 'canceled');

-- =====================================================================
-- Reads
-- =====================================================================

-- name: GetTask :one
SELECT * FROM tasks WHERE id = sqlc.arg(id);

-- name: GetTaskByDedupeKey :one
SELECT * FROM tasks WHERE queue = sqlc.arg(queue) AND dedupe_key = sqlc.arg(dedupe_key);

-- name: ListFlowRunTasks :many
SELECT * FROM tasks
WHERE flow_run_id = sqlc.arg(flow_run_id)
ORDER BY step_index, id;

-- name: ListTasksByResource :many
SELECT * FROM tasks
WHERE resource = sqlc.arg(resource)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- ListTasksByResourceType: resource_type is the prefix before ':' ("file", "image").
-- name: ListTasksByResourceType :many
SELECT * FROM tasks
WHERE resource LIKE sqlc.arg(resource_type) || ':%'
  AND status = sqlc.arg(status)
ORDER BY id DESC
LIMIT sqlc.arg(lim);

-- name: ListTasksByStatus :many
SELECT * FROM tasks
WHERE queue = sqlc.arg(queue) AND status = sqlc.arg(status)
ORDER BY updated_at DESC
LIMIT sqlc.arg(lim);

-- name: ListTaskDependencies :many
SELECT p.* FROM task_dependencies d
JOIN tasks p ON p.id = d.depends_on_id
WHERE d.task_id = sqlc.arg(task_id);

-- name: CountTasksByStatus :many
SELECT queue, status, COUNT(*) AS n
FROM tasks
GROUP BY queue, status
ORDER BY queue, status;

-- name: ListTaskAttempts :many
SELECT * FROM task_attempts WHERE task_id = sqlc.arg(task_id) ORDER BY attempt;

-- name: ListTaskEvents :many
SELECT * FROM task_events WHERE task_id = sqlc.arg(task_id) ORDER BY id;

-- =====================================================================
-- Housekeeping
-- =====================================================================

-- PurgeFinishedTasks: delete terminal standalone tasks older than a cutoff
-- (attempts and events go with them via ON DELETE CASCADE).
-- name: PurgeFinishedTasks :execrows
DELETE FROM tasks
WHERE status IN ('succeeded', 'canceled', 'dead')
  AND finished_at < sqlc.arg(before)
  AND flow_run_id IS NULL;

-- name: PurgeFinishedFlowRuns :execrows
DELETE FROM flow_runs
WHERE status <> 'running'
  AND finished_at < sqlc.arg(before);

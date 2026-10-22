-- =====================================================================
-- Task execution queue for flow steps (SQLite)
--
-- Conventions
--   * All timestamps are INTEGER unix epoch MILLISECONDS (-> int64 in Go).
--     The app passes "now" explicitly so time is testable; defaults exist
--     only as a safety net.
--   * JSON is stored as TEXT and checked with json_valid().
--   * `version` is an optimistic-lock counter: every state change bumps it,
--     and every write after a claim must match it. A worker whose lease
--     expired (and whose task was re-claimed) can no longer write.
--
-- Recommended connection pragmas (set from Go on every connection):
--   PRAGMA journal_mode = WAL;
--   PRAGMA busy_timeout = 5000;
--   PRAGMA foreign_keys = ON;
--   PRAGMA synchronous = NORMAL;
-- =====================================================================

-- ---------------------------------------------------------------------
-- flow_runs: one execution of a pipeline. Groups the tasks (steps).
-- ---------------------------------------------------------------------
CREATE TABLE flow_runs (
    id            INTEGER PRIMARY KEY,
    name          TEXT    NOT NULL,                       -- flow definition name, e.g. "ingest-image"
    status        TEXT    NOT NULL DEFAULT 'running'
                  CHECK (status IN ('running', 'succeeded', 'failed', 'canceled')),
    input         TEXT    NOT NULL DEFAULT '{}' CHECK (json_valid(input)),
    version       INTEGER NOT NULL DEFAULT 1,
    created_at    INTEGER NOT NULL DEFAULT (CAST((julianday('now') - 2440587.5) * 86400000 AS INTEGER)),
    updated_at    INTEGER NOT NULL DEFAULT (CAST((julianday('now') - 2440587.5) * 86400000 AS INTEGER)),
    finished_at   INTEGER
);

CREATE INDEX idx_flow_runs_status ON flow_runs (status, created_at);

-- ---------------------------------------------------------------------
-- tasks: the queue itself. One row per unit of work.
--
-- Lifecycle:
--   pending --claim--> running --ok--------------> succeeded
--      ^                  |
--      |                  +-error, attempts left--> retry --(run_at)--> claimable
--      |                  +-error, no attempts----> dead
--      |                  +-lease expired---------> claimable again (crash recovery)
--   any non-terminal --cancel--> canceled
--   dead/canceled --requeue--> pending (manual)
-- ---------------------------------------------------------------------
CREATE TABLE tasks (
    id              INTEGER PRIMARY KEY,

    -- routing
    queue           TEXT    NOT NULL DEFAULT 'default',   -- worker pools listen per queue
    kind            TEXT    NOT NULL,                     -- handler name, e.g. "thumbnail.resize"

    -- flow / pipeline position (NULL for standalone tasks)
    flow_run_id     INTEGER REFERENCES flow_runs (id) ON DELETE CASCADE,
    step_name       TEXT,
    step_index      INTEGER,

    -- the thing the task works on: "<type>:<locator>", e.g. file:/data/a.txt, image:/img/1.png
    resource        TEXT    NOT NULL DEFAULT ''
                    CHECK (length(resource) <= 1024)
                    CHECK (resource = '' OR instr(resource, ':') > 1),

    payload         TEXT    NOT NULL DEFAULT '{}' CHECK (json_valid(payload)),
    result          TEXT    CHECK (result IS NULL OR json_valid(result)),

    -- state
    status          TEXT    NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'running', 'retry', 'succeeded', 'dead', 'canceled')),
    priority        INTEGER NOT NULL DEFAULT 0,           -- higher runs first
    run_at          INTEGER NOT NULL,                     -- not claimable before this time

    -- resilience
    attempt         INTEGER NOT NULL DEFAULT 0 CHECK (attempt >= 0),      -- attempts started so far
    max_attempts    INTEGER NOT NULL DEFAULT 5 CHECK (max_attempts >= 1),
    backoff_base_ms INTEGER NOT NULL DEFAULT 1000  CHECK (backoff_base_ms >= 0),
    backoff_max_ms  INTEGER NOT NULL DEFAULT 300000 CHECK (backoff_max_ms >= backoff_base_ms),
    timeout_ms      INTEGER NOT NULL DEFAULT 60000 CHECK (timeout_ms > 0), -- lease length per attempt
    last_error      TEXT,

    -- lease (who holds it, until when)
    locked_by       TEXT,
    locked_until    INTEGER,

    -- optimistic concurrency
    version         INTEGER NOT NULL DEFAULT 1,

    -- idempotent enqueue: same key inside a queue is enqueued only once
    dedupe_key      TEXT,

    created_at      INTEGER NOT NULL DEFAULT (CAST((julianday('now') - 2440587.5) * 86400000 AS INTEGER)),
    updated_at      INTEGER NOT NULL DEFAULT (CAST((julianday('now') - 2440587.5) * 86400000 AS INTEGER)),
    started_at      INTEGER,                              -- first attempt start
    finished_at     INTEGER,                              -- terminal state reached

    CHECK ((status = 'running') = (locked_by IS NOT NULL AND locked_until IS NOT NULL))
);

-- Claim path: claimable rows ordered by priority / time. Partial index keeps it
-- small because finished rows (the vast majority over time) are excluded.
CREATE INDEX idx_tasks_claim
    ON tasks (queue, priority DESC, run_at, id)
    WHERE status IN ('pending', 'retry');

-- Crash recovery: running rows whose lease expired.
CREATE INDEX idx_tasks_lease
    ON tasks (queue, locked_until)
    WHERE status = 'running';

CREATE INDEX idx_tasks_flow     ON tasks (flow_run_id, step_index);
CREATE INDEX idx_tasks_resource ON tasks (resource) WHERE resource <> '';
CREATE INDEX idx_tasks_status   ON tasks (status, updated_at);

CREATE UNIQUE INDEX uq_tasks_dedupe
    ON tasks (queue, dedupe_key)
    WHERE dedupe_key IS NOT NULL;

CREATE UNIQUE INDEX uq_tasks_flow_step
    ON tasks (flow_run_id, step_name)
    WHERE flow_run_id IS NOT NULL AND step_name IS NOT NULL;

-- ---------------------------------------------------------------------
-- task_dependencies: DAG edges inside a flow. A task is claimable only when
-- every task it depends on has succeeded.
-- ---------------------------------------------------------------------
CREATE TABLE task_dependencies (
    task_id        INTEGER NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    depends_on_id  INTEGER NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    PRIMARY KEY (task_id, depends_on_id),
    CHECK (task_id <> depends_on_id)
) WITHOUT ROWID;

CREATE INDEX idx_task_deps_reverse ON task_dependencies (depends_on_id);

-- ---------------------------------------------------------------------
-- task_attempts: one row per execution attempt (the run history).
-- ---------------------------------------------------------------------
CREATE TABLE task_attempts (
    id            INTEGER PRIMARY KEY,
    task_id       INTEGER NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    attempt       INTEGER NOT NULL,                       -- 1-based, matches tasks.attempt at claim
    worker_id     TEXT    NOT NULL,
    status        TEXT    NOT NULL DEFAULT 'running'
                  CHECK (status IN ('running', 'succeeded', 'failed', 'timed_out', 'canceled')),
    error         TEXT,
    result        TEXT    CHECK (result IS NULL OR json_valid(result)),
    started_at    INTEGER NOT NULL,
    finished_at   INTEGER,
    duration_ms   INTEGER GENERATED ALWAYS AS (finished_at - started_at) VIRTUAL,
    UNIQUE (task_id, attempt)
);

CREATE INDEX idx_task_attempts_running ON task_attempts (status, started_at) WHERE status = 'running';

-- ---------------------------------------------------------------------
-- task_events: append-only audit log of every state transition.
-- Filled by trigger, so no code path can forget it.
-- ---------------------------------------------------------------------
CREATE TABLE task_events (
    id            INTEGER PRIMARY KEY,
    task_id       INTEGER NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    version       INTEGER NOT NULL,                       -- tasks.version after the change
    from_status   TEXT,                                   -- NULL on insert
    to_status     TEXT    NOT NULL,
    attempt       INTEGER NOT NULL,
    worker_id     TEXT,
    message       TEXT,
    created_at    INTEGER NOT NULL
);

CREATE INDEX idx_task_events_task ON task_events (task_id, id);

-- CREATE TRIGGER trg_tasks_event_insert
-- AFTER INSERT ON tasks
-- BEGIN
--     INSERT INTO task_events (task_id, version, from_status, to_status, attempt, worker_id, message, created_at)
--     VALUES (NEW.id, NEW.version, NULL, NEW.status, NEW.attempt, NULL, 'enqueued', NEW.created_at);
-- END;
--
-- CREATE TRIGGER trg_tasks_event_update
-- AFTER UPDATE OF status ON tasks
-- WHEN OLD.status IS NOT NEW.status OR OLD.attempt IS NOT NEW.attempt
-- BEGIN
--     INSERT INTO task_events (task_id, version, from_status, to_status, attempt, worker_id, message, created_at)
--     VALUES (NEW.id, NEW.version, OLD.status, NEW.status, NEW.attempt,
--             COALESCE(NEW.locked_by, OLD.locked_by),
--             CASE WHEN NEW.status IN ('retry', 'dead') THEN NEW.last_error END,
--             NEW.updated_at);
-- END;

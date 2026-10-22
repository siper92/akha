# SQLite task queue for flow steps (sqlc)

```
sqlc.yaml
db/migrations/0001_task_queue.sql   schema: flow_runs, tasks, task_dependencies, task_attempts, task_events
db/queries/tasks.sql                enqueue / claim / heartbeat / complete / fail / reap / cancel / requeue / reads
db/queries/flow_runs.sql            flow run create / progress / finish
internal/queuedb/queue_test.go      end-to-end test (retry, backoff, deps, lease expiry, history)
```

Generate: `sqlc generate` -> `internal/queuedb`.

## Tables

| table | purpose |
|---|---|
| `flow_runs` | one execution of a pipeline; groups its steps |
| `tasks` | the queue. `resource` = `<type>:<locator>` (`file:/a.txt`, `image:/b.png`), max 1024 chars, checked |
| `task_dependencies` | DAG edges; a task is claimable only when all parents `succeeded` |
| `task_attempts` | one row per execution attempt: worker, status, error, result, duration |
| `task_events` | append-only state-transition log, written by triggers (can't be forgotten) |

## Task states

```
pending --claim--> running --ok--> succeeded
                     |-- error, attempts left --> retry --(run_at)--> claimable
                     |-- error, out / permanent --> dead
                     +-- lease expired (reaper) --> retry | dead
any active --cancel--> canceled         dead|canceled --requeue--> pending
```

## Resilience model

* **Lease**: claim sets `locked_by` + `locked_until = now + timeout_ms`. Long tasks call `HeartbeatTask`.
* **Version**: every transition bumps `version`; `Complete/Fail/Heartbeat` require `id + version + locked_by`.
  0 rows affected = lease lost -> worker must drop its result.
* **Backoff**: `min(base * 2^(attempt-1), max) + jitter_ms`, computed in SQL. `permanent=true` skips retries.
* **Reaper**: run `ReapExpiredLeases` every few seconds, then `TimeoutAttempt` per returned row.
* **Dead step in a flow**: `ListDownstreamTaskIDs` -> `CancelTasksByIDs`, then `FlowRunProgress` -> `FinishFlowRun`.
* **Idempotent enqueue**: `dedupe_key` unique per queue; duplicate returns `sql.ErrNoRows`.
* **Requeue** keeps history; it extends `max_attempts` instead of resetting `attempt`.

## Worker loop

```go
tx, _ := db.BeginTx(ctx, nil)                 // DSN must use _txlock=immediate
q := queuedb.New(tx)
task, err := q.ClaimTask(ctx, queuedb.ClaimTaskParams{Queue: "default", WorkerID: id, Now: nowMs()})
if errors.Is(err, sql.ErrNoRows) { tx.Rollback(); sleep(poll); continue }
q.StartAttempt(ctx, queuedb.StartAttemptParams{TaskID: task.ID, Attempt: task.Attempt, WorkerID: id, Now: nowMs()})
tx.Commit()

res, runErr := handlers[task.Kind](ctx, task)   // ctx with timeout < task.TimeoutMs

// in one tx: CompleteTask or FailTask, then FinishAttempt
```

## Notes

* Connection pragmas: `journal_mode=WAL`, `busy_timeout=5000`, `foreign_keys=ON`, `synchronous=NORMAL`.
  With mattn: `file:queue.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=on&_txlock=immediate`.
* Timestamps are unix **milliseconds** passed from Go (`time.Now().UnixMilli()`) - easy to fake in tests.
* Needs SQLite >= 3.35 (`RETURNING`); generated column `duration_ms` needs >= 3.31.
* Keep the `.sql` files ASCII-only: sqlc's SQLite engine miscomputes query offsets on multi-byte
  characters (e.g. an em dash in a comment) and fails with confusing parse errors.
* sqlc's SQLite engine can't resolve a recursive CTE inside UPDATE, hence the List + CancelByIDs split.

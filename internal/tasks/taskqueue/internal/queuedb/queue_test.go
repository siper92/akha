package queuedb_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"example.com/taskqueue/internal/queuedb"
)

func open(t *testing.T) (*sql.DB, *queuedb.Queries) {
	t.Helper()
	// _txlock=immediate: BEGIN IMMEDIATE so claim transactions never deadlock on upgrade.
	db, err := sql.Open("sqlite3", "file::memory:?cache=shared&_foreign_keys=on&_busy_timeout=5000&_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	schema, err := os.ReadFile("../../db/migrations/0001_task_queue.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, queuedb.New(db)
}

func ptr[T any](v T) *T { return &v }

// claim = ClaimTask + StartAttempt in one transaction.
func claim(t *testing.T, db *sql.DB, queue, worker string, now int64) (queuedb.Task, bool) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	q := queuedb.New(tx)
	task, err := q.ClaimTask(ctx, queuedb.ClaimTaskParams{Queue: queue, WorkerID: worker, Now: now})
	if errors.Is(err, sql.ErrNoRows) {
		return task, false
	}
	if err != nil {
		t.Fatal(err)
	}
	if _, err := q.StartAttempt(ctx, queuedb.StartAttemptParams{TaskID: task.ID, Attempt: task.Attempt, WorkerID: worker, Now: now}); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return task, true
}

func TestFlowRetryLeaseAndHistory(t *testing.T) {
	ctx := context.Background()
	db, q := open(t)
	const now = int64(1_000)

	flow, err := q.CreateFlowRun(ctx, queuedb.CreateFlowRunParams{Name: "ingest", Input: "{}", Now: now})
	if err != nil {
		t.Fatal(err)
	}
	enqueue := func(step string, idx int64, res string) queuedb.Task {
		task, err := q.EnqueueTask(ctx, queuedb.EnqueueTaskParams{
			Queue: "default", Kind: "demo." + step, FlowRunID: &flow.ID, StepName: &step, StepIndex: &idx,
			Resource: res, Payload: "{}", RunAt: now, MaxAttempts: 2,
			BackoffBaseMs: 100, BackoffMaxMs: 1000, TimeoutMs: 500, Now: now,
		})
		if err != nil {
			t.Fatal(err)
		}
		return task
	}
	a := enqueue("read", 0, "file:/data/in.txt")
	b := enqueue("thumb", 1, "image:/data/in.png")
	c := enqueue("publish", 2, "")
	must(t, q.AddTaskDependency(ctx, queuedb.AddTaskDependencyParams{TaskID: b.ID, DependsOnID: a.ID}))
	must(t, q.AddTaskDependency(ctx, queuedb.AddTaskDependencyParams{TaskID: c.ID, DependsOnID: b.ID}))

	// dedupe: second enqueue with same key is a no-op
	p := queuedb.EnqueueTaskParams{Queue: "default", Kind: "x", Payload: "{}", RunAt: now + 1e9, MaxAttempts: 1, TimeoutMs: 1, Now: now, DedupeKey: ptr("k1")}
	if _, err := q.EnqueueTask(ctx, p); err != nil {
		t.Fatal(err)
	}
	if _, err := q.EnqueueTask(ctx, p); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("dedupe: want ErrNoRows, got %v", err)
	}

	// attempt 1: only A is claimable (B, C blocked by deps), then it fails -> retry with backoff
	t1, ok := claim(t, db, "default", "w1", now)
	if !ok || t1.ID != a.ID {
		t.Fatalf("want A, got %+v", t1)
	}
	if _, ok := claim(t, db, "default", "w2", now); ok {
		t.Fatal("B must be blocked by dependency")
	}
	f, err := q.FailTask(ctx, queuedb.FailTaskParams{ID: t1.ID, Version: t1.Version, WorkerID: "w1", Now: now, Error: ptr("boom")})
	if err != nil || f.Status != "retry" || f.RunAt != now+100 {
		t.Fatalf("fail: %+v %v", f, err)
	}
	must(t, q.FinishAttempt(ctx, queuedb.FinishAttemptParams{TaskID: a.ID, Attempt: 1, Status: "failed", Error: ptr("boom"), Now: now}))

	if _, ok := claim(t, db, "default", "w1", now+50); ok {
		t.Fatal("claimed before backoff elapsed")
	}

	// attempt 2: worker dies; reaper expires the lease -> dead (max_attempts=2)
	t2, ok := claim(t, db, "default", "w1", now+100)
	if !ok || t2.Attempt != 2 {
		t.Fatalf("attempt 2: %+v", t2)
	}
	reaped, err := q.ReapExpiredLeases(ctx, now+1000)
	if err != nil || len(reaped) != 1 || reaped[0].Status != "dead" {
		t.Fatalf("reap: %+v %v", reaped, err)
	}
	must(t, q.TimeoutAttempt(ctx, queuedb.TimeoutAttemptParams{TaskID: a.ID, Attempt: 2, Now: now + 1000}))

	// the stale worker can no longer write
	if n, _ := q.CompleteTask(ctx, queuedb.CompleteTaskParams{ID: a.ID, Version: t2.Version, WorkerID: "w1", Now: now + 1000}); n != 0 {
		t.Fatal("stale worker completed a reaped task")
	}

	// cancel everything downstream of the dead step
	ids, err := q.ListDownstreamTaskIDs(ctx, a.ID)
	if err != nil || len(ids) != 2 {
		t.Fatalf("downstream: %v %v", ids, err)
	}
	if n, _ := q.CancelTasksByIDs(ctx, queuedb.CancelTasksByIDsParams{Ids: ids, Reason: "upstream failed", Now: now + 1000}); n != 2 {
		t.Fatalf("canceled %d", n)
	}
	prog, _ := q.FlowRunProgress(ctx, &flow.ID)
	if prog.Dead != 1 || prog.Canceled != 2 || prog.Active != 0 {
		t.Fatalf("progress %+v", prog)
	}

	// manual requeue with 1 extra attempt, then success
	if n, _ := q.RequeueTask(ctx, queuedb.RequeueTaskParams{ID: a.ID, ExtraAttempts: 1, Now: now + 2000}); n != 1 {
		t.Fatal("requeue")
	}
	t3, ok := claim(t, db, "default", "w3", now+2000)
	if !ok || t3.Attempt != 3 {
		t.Fatalf("attempt 3: %+v", t3)
	}
	if n, _ := q.HeartbeatTask(ctx, queuedb.HeartbeatTaskParams{ID: t3.ID, Version: t3.Version, WorkerID: "w3", Now: now + 2050}); n != 1 {
		t.Fatal("heartbeat")
	}
	if n, _ := q.CompleteTask(ctx, queuedb.CompleteTaskParams{ID: t3.ID, Version: t3.Version, WorkerID: "w3", Result: ptr(`{"ok":true}`), Now: now + 2100}); n != 1 {
		t.Fatal("complete")
	}
	must(t, q.FinishAttempt(ctx, queuedb.FinishAttemptParams{TaskID: a.ID, Attempt: 3, Status: "succeeded", Now: now + 2100}))

	attempts, _ := q.ListTaskAttempts(ctx, a.ID)
	events, _ := q.ListTaskEvents(ctx, a.ID)
	if len(attempts) != 3 || attempts[1].Status != "timed_out" || *attempts[2].DurationMs != 100 {
		t.Fatalf("attempts %+v", attempts)
	}
	want := []string{"pending", "running", "retry", "running", "dead", "pending", "running", "succeeded"}
	if len(events) != len(want) {
		t.Fatalf("events %+v", events)
	}
	for i, e := range events {
		if e.ToStatus != want[i] {
			t.Fatalf("event %d: %s want %s", i, e.ToStatus, want[i])
		}
	}

	// resource constraints
	for _, bad := range []string{"no-prefix", string(make([]byte, 1025))} {
		p := p
		p.DedupeKey, p.Resource = nil, bad
		if _, err := q.EnqueueTask(ctx, p); err == nil {
			t.Fatalf("resource %q accepted", bad[:min(len(bad), 10)])
		}
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

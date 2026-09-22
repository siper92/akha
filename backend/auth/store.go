package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	db_sdk "github.com/siper92/akha/sdk/db-sdk"
	_ "modernc.org/sqlite"
)

const sqliteDriver = "sqlite"

type IssuedTokenStore interface {
	Store(ctx context.Context, workerID int64, token string, expires time.Time) error
	Active(ctx context.Context, token string) (bool, error)
	Revoke(ctx context.Context, token string) error
}

type Store interface {
	AccessTokenStore
	LoginLog
	IssuedTokenStore
	Seed(ctx context.Context, accessTokens []string) error
}

type store struct {
	q db_sdk.Querier
}

//go:embed schema.sql
var Schema string

var _ Store = (*store)(nil)

func NewStore(q db_sdk.Querier) Store {
	return &store{q: q}
}

func OpenDB(ctx context.Context, path string) (*sql.DB, error) {
	if path != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return nil, err
		}
	}
	db, err := sql.Open(sqliteDriver, path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, Schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return db, nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *store) Seed(ctx context.Context, accessTokens []string) error {
	for i, tok := range accessTokens {
		err := s.q.UpsertWorker(ctx, db_sdk.UpsertWorkerParams{
			Name:      fmt.Sprintf("worker-%d", i+1),
			TokenHash: HashToken(tok),
			Tier:      string(TierWorker),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *store) Lookup(ctx context.Context, hash string) (Worker, bool, error) {
	w, err := s.q.GetWorkerByTokenHash(ctx, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return Worker{}, false, nil
	}
	if err != nil {
		return Worker{}, false, err
	}
	return Worker{ID: w.ID, Name: w.Name, Tier: Tier(w.Tier)}, true, nil
}

func (s *store) Record(ctx context.Context, a Attempt) error {
	at := a.At
	if at.IsZero() {
		at = time.Now()
	}
	var wid sql.NullInt64
	if a.WorkerID != nil {
		wid = sql.NullInt64{Int64: *a.WorkerID, Valid: true}
	}
	return s.q.RecordLogin(ctx, db_sdk.RecordLoginParams{
		WorkerID:  wid,
		Addr:      a.Addr,
		Ok:        a.OK,
		Reason:    a.Reason,
		CreatedAt: at,
	})
}

func (s *store) Store(ctx context.Context, workerID int64, token string, expires time.Time) error {
	return s.q.StoreIssuedToken(ctx, db_sdk.StoreIssuedTokenParams{
		WorkerID:  sql.NullInt64{Int64: workerID, Valid: true},
		Token:     token,
		ExpiresAt: expires,
	})
}

func (s *store) Active(ctx context.Context, token string) (bool, error) {
	t, err := s.q.GetIssuedToken(ctx, token)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return t.IsActive && time.Now().Before(t.ExpiresAt), nil
}

func (s *store) Revoke(ctx context.Context, token string) error {
	return s.q.RevokeIssuedToken(ctx, token)
}

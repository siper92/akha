package auth

import (
	"context"
	"time"
)

type Tier string

const (
	TierBackend Tier = "backend"
	TierWorker  Tier = "worker"
)

type Claims struct {
	Subject   string
	Tier      Tier
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type Worker struct {
	ID   int64
	Name string
	Tier Tier
}

type Attempt struct {
	WorkerID *int64
	Addr     string
	OK       bool
	Reason   string
	At       time.Time
}

type TokenIssuer interface {
	Issue(ctx context.Context, subject string, tier Tier) (token string, expires time.Time, err error)
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (Claims, error)
}

type AccessTokenStore interface {
	Lookup(ctx context.Context, hash string) (Worker, bool, error)
}

type LoginLog interface {
	Record(ctx context.Context, a Attempt) error
}

type Authenticator interface {
	Register(ctx context.Context, accessToken string) (token string, expires time.Time, err error)
	TokenVerifier
}

package backend

import (
	"context"
	"log/slog"

	"github.com/siper92/akha/backend/auth"
	"github.com/siper92/akha/lang/runner"
	"github.com/siper92/akha/worker"
	"github.com/siper92/akha/worker/client"
)

const embeddedSubject = "backend"

type localBackend struct {
	keys auth.Keyring
}

var (
	_ client.Backend     = (*localBackend)(nil)
	_ client.TokenSource = (*localBackend)(nil)
)

func NewLocalBackend(keys auth.Keyring) *localBackend {
	return &localBackend{keys: keys}
}

func (l *localBackend) Login(ctx context.Context) (string, error) {
	token, _, err := l.keys.Issue(ctx, embeddedSubject, auth.TierBackend)
	return token, err
}

func (l *localBackend) Validate(ctx context.Context, token string) (bool, error) {
	_, err := l.keys.Verify(ctx, token)
	return err == nil, nil
}

func (l *localBackend) Token(ctx context.Context) (string, error) { return l.Login(ctx) }

func (l *localBackend) Invalidate(ctx context.Context) error { return nil }

func EmbeddedFactory(keys auth.Keyring, run runner.Runner, log *slog.Logger) worker.Factory {
	return func(ctx context.Context, i int) (worker.Worker, error) {
		lb := NewLocalBackend(keys)
		return worker.New(lb, lb, run, log.With("worker", i)), nil
	}
}

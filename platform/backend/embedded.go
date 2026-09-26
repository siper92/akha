package backend

import (
	"context"
	"log/slog"

	"github.com/siper92/akha/lang/runner"
	auth2 "github.com/siper92/akha/platform/backend/auth"
	worker2 "github.com/siper92/akha/platform/worker"
	"github.com/siper92/akha/platform/worker/client"
)

const embeddedSubject = "backend"

type localBackend struct {
	keys auth2.Keyring
}

var (
	_ client.Backend     = (*localBackend)(nil)
	_ client.TokenSource = (*localBackend)(nil)
)

func NewLocalBackend(keys auth2.Keyring) *localBackend {
	return &localBackend{keys: keys}
}

func (l *localBackend) Login(ctx context.Context) (string, error) {
	token, _, err := l.keys.Issue(ctx, embeddedSubject, auth2.TierBackend)
	return token, err
}

func (l *localBackend) Validate(ctx context.Context, token string) (bool, error) {
	_, err := l.keys.Verify(ctx, token)
	return err == nil, nil
}

func (l *localBackend) Token(ctx context.Context) (string, error) { return l.Login(ctx) }

func (l *localBackend) Invalidate(ctx context.Context) error { return nil }

func EmbeddedFactory(keys auth2.Keyring, run runner.Runner, log *slog.Logger) worker2.Factory {
	return func(ctx context.Context, i int) (worker2.Worker, error) {
		lb := NewLocalBackend(keys)
		return worker2.New(lb, lb, run, log.With("worker", i)), nil
	}
}

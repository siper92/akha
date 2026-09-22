package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrRevoked      = errors.New("token revoked")
)

type addrKey struct{}

func WithAddr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, addrKey{}, addr)
}

func GetAdders(ctx context.Context) string {
	addr, _ := ctx.Value(addrKey{}).(string)
	return addr
}

type authenticator struct {
	keys  Keyring
	store Store
	log   *slog.Logger
}

var _ Authenticator = (*authenticator)(nil)

func NewAuthenticator(keys Keyring, store Store, log *slog.Logger) Authenticator {
	if log == nil {
		log = slog.Default()
	}

	return &authenticator{keys: keys, store: store, log: log}
}

func (a *authenticator) Register(ctx context.Context, accessToken string) (string, time.Time, error) {
	addr := GetAdders(ctx)
	w, ok, err := a.store.Lookup(ctx, HashToken(accessToken))
	if err != nil {
		return "", time.Time{}, err
	}

	if !ok {
		a.record(ctx, Attempt{Addr: addr, OK: false, Reason: "unknown access token"})
		return "", time.Time{}, ErrUnauthorized
	}

	token, exp, err := a.keys.Issue(ctx, w.Name, w.Tier)
	if err != nil {
		a.record(ctx, Attempt{WorkerID: &w.ID, Addr: addr, OK: false, Reason: "issue failed"})
		return "", time.Time{}, err
	}

	if err := a.store.Store(ctx, w.ID, token, exp); err != nil {
		return "", time.Time{}, err
	}
	a.record(ctx, Attempt{WorkerID: &w.ID, Addr: addr, OK: true})

	return token, exp, nil
}

func (a *authenticator) Verify(ctx context.Context, token string) (Claims, error) {
	c, err := a.keys.Verify(ctx, token)
	if err != nil {
		return Claims{}, err
	}

	active, err := a.store.Active(ctx, token)
	if err != nil {
		return Claims{}, err
	}

	if !active {
		return Claims{}, ErrRevoked
	}

	return c, nil
}

func (a *authenticator) record(ctx context.Context, at Attempt) {
	at.At = time.Now()
	a.log.Info("login", "addr", at.Addr, "ok", at.OK, "reason", at.Reason)
	if err := a.store.Record(ctx, at); err != nil {
		a.log.Error("login log", "err", err)
	}
}

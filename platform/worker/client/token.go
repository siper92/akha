package client

import (
	"context"

	"github.com/siper92/akha/internal/cache"
)

const DefaultTokenKey = "worker.jwt"

type tokenSource struct {
	be  Backend
	c   cache.Cache
	key string
}

var _ TokenSource = (*tokenSource)(nil)

func NewTokenSource(be Backend, c cache.Cache, key string) TokenSource {
	if key == "" {
		key = DefaultTokenKey
	}

	return &tokenSource{be: be, c: c, key: key}
}

func (t *tokenSource) Token(ctx context.Context) (string, error) {
	if b, ok, err := t.c.Get(ctx, t.key); err != nil {
		return "", err
	} else if ok {
		return string(b), nil
	}
	tok, err := t.be.Login(ctx)
	if err != nil {
		return "", err
	}
	if err := t.c.Put(ctx, t.key, []byte(tok), cache.NoExpiry); err != nil {
		return "", err
	}
	return tok, nil
}

func (t *tokenSource) Invalidate(ctx context.Context) error {
	return t.c.Del(ctx, t.key)
}

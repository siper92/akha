package client

import "context"

type Backend interface {
	Login(ctx context.Context) (string, error)
	Validate(ctx context.Context, token string) (bool, error)
}

type TokenSource interface {
	Token(ctx context.Context) (string, error)
	Invalidate(ctx context.Context) error
}

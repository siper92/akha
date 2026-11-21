package v1

import (
	"context"
	"errors"
)

const (
	ServerName    = "akha-files"
	ServerVersion = "v0.1.0"
	DefaultModel  = "claude-opus-5-5"
)

var (
	ErrRefused       = errors.New("model refused the request")
	ErrMaxIterations = errors.New("max tool iterations reached")
	ErrBadFileName   = errors.New("invalid file name")
)

type Agent interface {
	Run(ctx context.Context, prompt string) (string, error)
}

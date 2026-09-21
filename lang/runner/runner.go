package runner

import (
	"context"

	"github.com/siper92/akha/lang/check"
)

type Options struct {
	RunID    string
	Root     string
	CacheDir string
}

type Result struct {
	ExitCode    int
	Diagnostics []check.Diagnostic
	LogPath     string
	DebugPath   string
}

type Runner interface {
	Check(ctx context.Context, src string) ([]check.Diagnostic, error)
	Run(ctx context.Context, src string, opts Options) (Result, error)
}

package worker

import (
	"context"

	"github.com/siper92/akha/lang/runner"
)

type Worker interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Run(ctx context.Context, src string, opts runner.Options) (runner.Result, error)
}

type Spawner interface {
	Spawn(ctx context.Context, n int) error
}

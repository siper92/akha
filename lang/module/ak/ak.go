package ak

import (
	"context"

	"github.com/siper92/akha/lang/eval"
)

const ModuleName = "Ak"

type Output interface {
	Setup(ctx context.Context, logPath, debugPath string) error
	Log(ctx context.Context, msg string) error
	Debug(ctx context.Context, msg string, args ...eval.Value) error
	Close() error
}

type Allowed interface {
	Allow(names ...string)
	IsAllowed(name string) bool
	List() []string
}

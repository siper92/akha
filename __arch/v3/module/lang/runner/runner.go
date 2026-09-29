package runner

import (
	"context"
	"errors"

	"github.com/siper92/akha/lang/check"
	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/parser"
)

type Runner interface {
	Run(ctx context.Context, src string, input eval.Value) (eval.Value, error)
}

var _ Runner = (*runner)(nil)

type runner struct {
	opts eval.Options
}

func New(opts eval.Options) Runner {
	return &runner{opts: opts}
}

func (r *runner) Run(ctx context.Context, src string, input eval.Value) (eval.Value, error) {
	script, err := parser.New("", src).Parse()
	if err != nil {
		if le, ok := errors.AsType[*lexer.Error](err); ok {
			le.File = r.opts.File
		}
		return nil, err
	}

	if err := check.New(r.opts.File).Check(ctx, script); err != nil {
		return nil, err
	}

	return eval.New(r.opts).Run(ctx, script, input)
}

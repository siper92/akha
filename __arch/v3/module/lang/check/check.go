package check

import (
	"context"

	"github.com/siper92/akha/lang/ast"
)

type Checker interface {
	Check(ctx context.Context, script *ast.Script) error
}

type Known func(x ast.Expr) ast.Expr

type ValueChecker interface {
	CheckValue(ctx context.Context, x ast.Expr, known Known) error
}

package check

import (
	"context"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/token"
)

type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
)

type Diagnostic struct {
	Pos      token.Pos
	Severity Severity
	Msg      string
}

type Checker interface {
	Check(ctx context.Context, s *ast.Script, reg eval.Registry) []Diagnostic
}

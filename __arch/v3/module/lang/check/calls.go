package check

import (
	"context"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/module"
)

var _ ValueChecker = (*CallChecker)(nil)

type CallChecker struct {
	Modules module.IRegistry
}

func (c *CallChecker) CheckValue(_ context.Context, x ast.Expr, known Known) error {
	call, ok := x.(*ast.CallExpr)
	if !ok {
		return nil
	}
	name, fn, ok := module.Callee(call)
	if !ok {
		return nil
	}
	m, ok := c.Modules.Module(name)
	if !ok {
		return nil
	}

	f, ok := m.Func(fn)
	if !ok {
		return newError(call.Pos, lexer.CodeUnknownFunc, "", "unknown function %s.%s", name, fn)
	}
	if len(call.Args) != len(f.Params()) {
		return newError(call.Pos, lexer.CodeArity, "", "wrong number of arguments for %s.%s, want %d, got %d", name, fn, len(f.Params()), len(call.Args))
	}
	for i, p := range f.Params() {
		if !p.Type.Accepts(known(call.Args[i])) {
			return argKind(call.Pos, name, fn, p)
		}
	}
	for _, kw := range call.Kwargs {
		p, ok := f.Kwarg(kw.Name)
		if !ok {
			return newError(kw.Pos, lexer.CodeUnknownKwarg, "", "unknown named argument %s for %s.%s", kw.Name, name, fn)
		}
		if !p.Type.Accepts(known(kw.Value)) {
			return argKind(kw.Pos, name, fn, p)
		}
	}

	return nil
}

func argKind(pos ast.Pos, name, fn string, p module.Param) error {
	return newError(pos, lexer.CodeArgKind, "", "argument %s of %s.%s must be a %s", p.Name, name, fn, p.Type.Name())
}

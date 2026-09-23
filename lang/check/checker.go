package check

import (
	"context"
	"fmt"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/token"
)

const (
	akModule = "Ak"
	akAllow  = "Allow"
)

type checker struct{}

var _ Checker = (*checker)(nil)

func New() Checker {
	return &checker{}
}

type scope struct {
	parent *scope
	names  map[string]bool
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, names: map[string]bool{}}
}

func (s *scope) declared(name string) bool {
	for sc := s; sc != nil; sc = sc.parent {
		if sc.names[name] {
			return true
		}
	}
	return false
}

type walk struct {
	reg     eval.Registry
	allowed map[string]bool
	diags   []Diagnostic
}

func (c *checker) Check(ctx context.Context, s *ast.Script, reg eval.Registry) []Diagnostic {
	if s == nil || len(s.Stmts) == 0 || reg == nil {
		return nil
	}
	w := &walk{reg: reg, allowed: map[string]bool{akModule: true}}
	sc := newScope(nil)

	for i, st := range s.Stmts {
		if ctx.Err() != nil {
			break
		}
		if st == nil {
			continue
		}
		w.stmt(st, sc, i == 0)
	}

	return w.diags
}

func (w *walk) errorf(pos token.Pos, format string, args ...any) {
	w.diags = append(w.diags, Diagnostic{Pos: pos, Severity: SeverityError, Msg: fmt.Sprintf(format, args...)})
}

func (w *walk) stmt(st ast.Stmt, sc *scope, first bool) {
	if first && !isAllowStmt(st) {
		w.errorf(st.Pos(), "first statement must be Ak.Allow")
	}
	switch s := st.(type) {
	case *ast.CallStmt:
		w.call(s.Call, sc, first)
	case *ast.Let:
		w.expr(s.Value, sc)
		if sc.names[s.Name] {
			w.errorf(s.P, "variable %s already declared", s.Name)
			return
		}
		sc.names[s.Name] = true
	case *ast.Assign:
		w.expr(s.Value, sc)
		if !sc.declared(s.Name) {
			w.errorf(s.P, "undefined variable %s", s.Name)
		}
	case *ast.If:
	case *ast.For:
	case *ast.While:
	case *ast.Break:
	case *ast.Continue:
	}
}

func (w *walk) expr(x ast.Expr, sc *scope) {
	switch v := x.(type) {
	case *ast.Ident:
		if !sc.declared(v.Name) {
			w.errorf(v.P, "undefined variable %s", v.Name)
		}
	case *ast.Call:
		w.call(v, sc, false)
	case *ast.Binary:
		w.expr(v.X, sc)
		w.expr(v.Y, sc)
	case *ast.Unary:
		w.expr(v.X, sc)
	case *ast.Spread:
		w.errorf(v.P, "spread only allowed in Ak.Allow")
	}
}

func (w *walk) call(call *ast.Call, sc *scope, first bool) {
	t := call.Target
	isAllow := t.Module == akModule && t.Name == akAllow
	if !first && isAllow {
		w.errorf(call.P, "Ak.Allow must be the first statement")
	}

	mod, ok := w.reg.Lookup(t.Module)
	if !ok {
		w.errorf(call.P, "unknown module %s", t.Module)
		return
	}

	if !w.allowed[t.Module] {
		w.errorf(call.P, "module %s not allowed, add Ak.Allow(%s...)", t.Module, t.Module)
	}

	fn, ok := mod.Func(t.Name)
	if !ok {
		w.errorf(call.P, "unknown function %s.%s", t.Module, t.Name)
		return
	}
	spec := fn.Spec()

	w.positional(call, sc, isAllow && first, isAllow)
	w.arity(call, spec)
	w.kwargs(call, sc, spec)
}

func (w *walk) positional(call *ast.Call, sc *scope, extend, isAllow bool) {
	for _, arg := range call.Args {
		if arg == nil {
			continue
		}

		sp, isSpread := arg.(*ast.Spread)
		switch {
		case isSpread && isAllow:
			if _, ok := w.reg.Lookup(sp.Module); !ok {
				w.errorf(sp.P, "unknown module %s", sp.Module)
				continue
			}

			if extend {
				w.allowed[sp.Module] = true
			}
		case isSpread:
			w.errorf(sp.P, "spread only allowed in Ak.Allow")
		case isAllow:
			w.errorf(arg.Pos(), "Ak.Allow accepts only module spreads")
		default:
			w.expr(arg, sc)
		}
	}
}

func (w *walk) arity(call *ast.Call, spec eval.Spec) {
	n := len(call.Args)
	name := call.Target.Module + "." + call.Target.Name
	if n < spec.MinArgs {
		w.errorf(call.P, "%s expects at least %d arguments, got %d", name, spec.MinArgs, n)
		return
	}

	if !spec.Variadic && n > spec.MaxArgs {
		w.errorf(call.P, "%s expects at most %d arguments, got %d", name, spec.MaxArgs, n)
	}
}

func (w *walk) kwargs(call *ast.Call, sc *scope, spec eval.Spec) {
	name := call.Target.Module + "." + call.Target.Name
	seen := map[string]bool{}

	for _, kw := range call.Kwargs {
		w.expr(kw.Value, sc)
		if seen[kw.Name] {
			w.errorf(kw.P, "duplicate argument %s", kw.Name)
			continue
		}
		seen[kw.Name] = true
		if !hasKwarg(spec, kw.Name) {
			w.errorf(kw.P, "%s has no argument %s", name, kw.Name)
		}
	}
}

func isAllowStmt(st ast.Stmt) bool {
	cs, ok := st.(*ast.CallStmt)
	if !ok || cs.Call == nil {
		return false
	}
	return cs.Call.Target.Module == akModule && cs.Call.Target.Name == akAllow
}

func hasKwarg(spec eval.Spec, name string) bool {
	for _, k := range spec.Kwargs {
		if k == name {
			return true
		}
	}
	return false
}

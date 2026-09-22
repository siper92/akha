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

type walk struct {
	reg     eval.Registry
	allowed map[string]bool
	diags   []Diagnostic
}

func (c *checker) Check(ctx context.Context, s *ast.Script, reg eval.Registry) []Diagnostic {
	if s == nil || len(s.Calls) == 0 || reg == nil {
		return nil
	}
	w := &walk{reg: reg, allowed: map[string]bool{akModule: true}}

	for i, call := range s.Calls {
		if ctx.Err() != nil {
			break
		}
		if call == nil {
			continue
		}
		w.call(i, call)
	}

	return w.diags
}

func (w *walk) errorf(pos token.Pos, format string, args ...any) {
	w.diags = append(w.diags, Diagnostic{Pos: pos, Severity: SeverityError, Msg: fmt.Sprintf(format, args...)})
}

func (w *walk) call(i int, call *ast.Call) {
	t := call.Target
	isAllow := t.Module == akModule && t.Name == akAllow
	if i == 0 && !isAllow {
		w.errorf(call.P, "first call must be Ak.Allow")
	}

	if i > 0 && isAllow {
		w.errorf(call.P, "Ak.Allow must be the first call")
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

	w.positional(call, isAllow && i == 0, isAllow)
	w.arity(call, spec)
	w.kwargs(call, spec)
}

func (w *walk) positional(call *ast.Call, extend, isAllow bool) {
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

func (w *walk) kwargs(call *ast.Call, spec eval.Spec) {
	name := call.Target.Module + "." + call.Target.Name
	seen := map[string]bool{}

	for _, kw := range call.Kwargs {
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

func hasKwarg(spec eval.Spec, name string) bool {
	for _, k := range spec.Kwargs {
		if k == name {
			return true
		}
	}
	return false
}

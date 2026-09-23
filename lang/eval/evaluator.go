package eval

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/token"
)

var (
	ErrUnknownModule   = errors.New("unknown module")
	ErrUnknownFunction = errors.New("unknown function")
	ErrBadLiteral      = errors.New("bad literal")
	ErrBadExpr         = errors.New("bad expression")
	ErrBadStmt         = errors.New("bad statement")
	ErrNotImplemented  = errors.New("not implemented")
)

type signal int

const (
	sigNone signal = iota
	sigBreak
	sigContinue
)

type evaluator struct {
	reg Registry
}

var _ Evaluator = (*evaluator)(nil)

func New(reg Registry) Evaluator {
	return &evaluator{reg: reg}
}

func (e *evaluator) Eval(ctx context.Context, s *ast.Script) error {
	if s == nil {
		return nil
	}

	env := NewEnv(nil)
	for _, st := range s.Stmts {
		if err := ctx.Err(); err != nil {
			return err
		}
		if st == nil {
			continue
		}
		if _, err := e.stmt(ctx, env, st); err != nil {
			return err
		}
	}

	return nil
}

func (e *evaluator) block(ctx context.Context, env Env, b *ast.Block) (signal, error) {
	if b == nil {
		return sigNone, nil
	}

	inner := env.Child()
	for _, st := range b.Stmts {
		if err := ctx.Err(); err != nil {
			return sigNone, err
		}

		if st == nil {
			continue
		}

		sig, err := e.stmt(ctx, inner, st)
		if err != nil || sig != sigNone {
			return sig, err
		}
	}

	return sigNone, nil
}

func (e *evaluator) stmt(ctx context.Context, env Env, st ast.Stmt) (signal, error) {
	switch s := st.(type) {
	case *ast.CallStmt:
		_, err := e.call(ctx, env, s.Call)
		return sigNone, err
	case *ast.Let:
		v, err := e.expr(ctx, env, s.Value)
		if err != nil {
			return sigNone, err
		}

		if err := env.Define(s.Name, v); err != nil {
			return sigNone, fmt.Errorf("%s: %w", posString(s.P), err)
		}

		return sigNone, nil
	case *ast.Assign:
		v, err := e.expr(ctx, env, s.Value)
		if err != nil {
			return sigNone, err
		}
		if err := env.Assign(s.Name, v); err != nil {
			return sigNone, fmt.Errorf("%s: %w", posString(s.P), err)
		}

		return sigNone, nil
	case *ast.If:
		return sigNone, fmt.Errorf("%s: %w", posString(s.P), ErrNotImplemented)
	case *ast.For:
		return sigNone, fmt.Errorf("%s: %w", posString(s.P), ErrNotImplemented)
	case *ast.While:
		return sigNone, fmt.Errorf("%s: %w", posString(s.P), ErrNotImplemented)
	case *ast.Break:
		return sigNone, fmt.Errorf("%s: %w", posString(s.P), ErrNotImplemented)
	case *ast.Continue:
		return sigNone, fmt.Errorf("%s: %w", posString(s.P), ErrNotImplemented)
	}

	return sigNone, fmt.Errorf("%s: %w", posString(st.Pos()), ErrBadStmt)
}

func (e *evaluator) expr(ctx context.Context, env Env, x ast.Expr) (Value, error) {
	switch v := x.(type) {
	case nil:
		return None(), nil
	case *ast.Literal:
		return literal(v)
	case *ast.Spread:
		return Str(v.Module), nil
	case *ast.Ident:
		val, ok := env.Lookup(v.Name)
		if !ok {
			return nil, fmt.Errorf("%s: %w: %s", posString(v.P), ErrUndefined, v.Name)
		}
		return val, nil
	case *ast.Call:
		return e.call(ctx, env, v)
	case *ast.Unary:
		operand, err := e.expr(ctx, env, v.X)
		if err != nil {
			return nil, err
		}
		out, err := Unary(v.Op, operand)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", posString(v.P), err)
		}
		return out, nil
	case *ast.Binary:
		return e.binary(ctx, env, v)
	}

	return nil, fmt.Errorf("%s: %w", posString(x.Pos()), ErrBadExpr)
}

func (e *evaluator) binary(ctx context.Context, env Env, b *ast.Binary) (Value, error) {
	left, err := e.expr(ctx, env, b.X)
	if err != nil {
		return nil, err
	}
	if b.Op == token.AND || b.Op == token.OR {
		l, ok := AsBool(left)
		if !ok {
			return nil, fmt.Errorf("%s: %w: %s on %s", posString(b.P), ErrOpType, opLit(b.Op), TypeOf(left))
		}
		if (b.Op == token.AND && !l) || (b.Op == token.OR && l) {
			return Bool(l), nil
		}
	}
	right, err := e.expr(ctx, env, b.Y)
	if err != nil {
		return nil, err
	}
	out, err := Binary(b.Op, left, right)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", posString(b.P), err)
	}
	return out, nil
}

func (e *evaluator) call(ctx context.Context, env Env, c *ast.Call) (Value, error) {
	name := c.Target.Module + "." + c.Target.Name
	mod, ok := e.reg.Lookup(c.Target.Module)
	if !ok {
		return nil, fmt.Errorf("%s: %w: %s", posString(c.P), ErrUnknownModule, c.Target.Module)
	}
	fn, ok := mod.Func(c.Target.Name)
	if !ok {
		return nil, fmt.Errorf("%s: %w: %s", posString(c.P), ErrUnknownFunction, name)
	}
	args := make([]Value, 0, len(c.Args))
	for _, a := range c.Args {
		v, err := e.expr(ctx, env, a)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}
	kwargs := make(map[string]Value, len(c.Kwargs))
	for _, kw := range c.Kwargs {
		v, err := e.expr(ctx, env, kw.Value)
		if err != nil {
			return nil, err
		}
		kwargs[kw.Name] = v
	}
	out, err := fn.Call(ctx, args, kwargs)
	if err != nil {
		var exit *ExitError
		if errors.As(err, &exit) {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %s: %w", posString(c.P), name, err)
	}
	if out == nil {
		return None(), nil
	}
	return out, nil
}

func literal(v *ast.Literal) (Value, error) {
	switch v.Kind {
	case token.STRING:
		return Str(v.Value), nil
	case token.INT:
		i, err := strconv.ParseInt(v.Value, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%s: %w: %s", posString(v.P), ErrBadLiteral, v.Value)
		}
		return Int(i), nil
	case token.TRUE:
		return Bool(true), nil
	case token.FALSE:
		return Bool(false), nil
	}
	return nil, fmt.Errorf("%s: %w: %s", posString(v.P), ErrBadLiteral, v.Value)
}

func posString(p token.Pos) string {
	return strconv.Itoa(p.Line) + ":" + strconv.Itoa(p.Col)
}

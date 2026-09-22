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
	for _, call := range s.Calls {
		if err := ctx.Err(); err != nil {
			return err
		}
		if call == nil {
			continue
		}
		if err := e.call(ctx, call); err != nil {
			return err
		}
	}
	return nil
}

func (e *evaluator) call(ctx context.Context, c *ast.Call) error {
	name := c.Target.Module + "." + c.Target.Name
	mod, ok := e.reg.Lookup(c.Target.Module)
	if !ok {
		return fmt.Errorf("%s: %w: %s", posString(c.P), ErrUnknownModule, c.Target.Module)
	}
	fn, ok := mod.Func(c.Target.Name)
	if !ok {
		return fmt.Errorf("%s: %w: %s", posString(c.P), ErrUnknownFunction, name)
	}
	args := make([]Value, 0, len(c.Args))
	for _, a := range c.Args {
		v, err := literal(a)
		if err != nil {
			return err
		}
		args = append(args, v)
	}
	kwargs := make(map[string]Value, len(c.Kwargs))
	for _, kw := range c.Kwargs {
		v, err := literal(kw.Value)
		if err != nil {
			return err
		}
		kwargs[kw.Name] = v
	}
	if _, err := fn.Call(ctx, args, kwargs); err != nil {
		var exit *ExitError
		if errors.As(err, &exit) {
			return err
		}
		return fmt.Errorf("%s: %s: %w", posString(c.P), name, err)
	}
	return nil
}

func literal(x ast.Expr) (Value, error) {
	switch v := x.(type) {
	case *ast.Literal:
		switch v.Kind {
		case token.STRING:
			return Str(v.Value), nil
		case token.INT:
			i, err := strconv.ParseInt(v.Value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("%s: %w: %s", posString(v.P), ErrBadLiteral, v.Value)
			}
			return Int(i), nil
		case token.IDENT:
			switch v.Value {
			case "true":
				return Bool(true), nil
			case "false":
				return Bool(false), nil
			}
		}
		return nil, fmt.Errorf("%s: %w: %s", posString(v.P), ErrBadLiteral, v.Value)
	case *ast.Spread:
		return Str(v.Module), nil
	case nil:
		return None(), nil
	}
	return nil, fmt.Errorf("%s: %w", posString(x.Pos()), ErrBadLiteral)
}

func posString(p token.Pos) string {
	return strconv.Itoa(p.Line) + ":" + strconv.Itoa(p.Col)
}

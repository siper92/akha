package ak

import (
	"context"
	"fmt"
	"strings"

	"github.com/siper92/akha/lang/eval"
)

const (
	FuncAllow = "Allow"
	FuncSetup = "Setup"
	FuncLog   = "Log"
	FuncDebug = "Debug"
	FuncExit  = "Exit"
)

func New(out Output, allowed Allowed) eval.Module {
	return eval.NewModule(ModuleName,
		eval.NewBuiltin(eval.Spec{Name: FuncAllow, Variadic: true}, allowFn(allowed)),
		eval.NewBuiltin(eval.Spec{Name: FuncSetup, Kwargs: []string{"log", "debug"}}, setupFn(out)),
		eval.NewBuiltin(eval.Spec{Name: FuncLog, MinArgs: 1, MaxArgs: 1}, logFn(out)),
		eval.NewBuiltin(eval.Spec{Name: FuncDebug, MinArgs: 1, Variadic: true}, debugFn(out)),
		eval.NewBuiltin(eval.Spec{Name: FuncExit, MaxArgs: 1, Kwargs: []string{"code"}}, exitFn(out)),
	)
}

func allowFn(allowed Allowed) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		names := make([]string, 0, len(args))
		for i := range args {
			name, err := eval.ArgString(args, i)
			if err != nil {
				return nil, err
			}
			names = append(names, name)
		}
		allowed.Allow(names...)
		return eval.None(), nil
	}
}

func setupFn(out Output) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		logPath, err := eval.KwargString(kwargs, "log", "")
		if err != nil {
			return nil, err
		}
		debugPath, err := eval.KwargString(kwargs, "debug", "")
		if err != nil {
			return nil, err
		}
		return eval.None(), out.Setup(ctx, logPath, debugPath)
	}
}

func logFn(out Output) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		msg, err := eval.ArgString(args, 0)
		if err != nil {
			return nil, err
		}
		return eval.None(), out.Log(ctx, msg)
	}
}

func debugFn(out Output) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		msg, err := eval.ArgString(args, 0)
		if err != nil {
			return nil, err
		}
		return eval.None(), out.Debug(ctx, msg, args[1:]...)
	}
}

func exitFn(out Output) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		msg := ""
		if len(args) > 0 {
			m, err := eval.ArgString(args, 0)
			if err != nil {
				return nil, err
			}
			msg = m
		}
		code, err := eval.KwargInt(kwargs, "code", 0)
		if err != nil {
			return nil, err
		}
		line := fmt.Sprintf("exit code=%d", code)
		if msg != "" {
			line += " " + msg
		}
		if err := out.Log(ctx, line); err != nil {
			return nil, err
		}
		return eval.None(), &eval.ExitError{Code: int(code), Msg: msg}
	}
}

type allowed struct {
	names map[string]bool
	order []string
}

var _ Allowed = (*allowed)(nil)

func NewAllowed() Allowed {
	a := &allowed{names: map[string]bool{}}
	a.Allow(ModuleName)
	return a
}

func (a *allowed) Allow(names ...string) {
	for _, n := range names {
		if a.names[n] {
			continue
		}
		a.names[n] = true
		a.order = append(a.order, n)
	}
}

func (a *allowed) IsAllowed(name string) bool { return a.names[name] }

func (a *allowed) List() []string {
	out := make([]string, len(a.order))
	copy(out, a.order)
	return out
}

func formatArgs(args []eval.Value) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, a := range args {
		if a == nil {
			parts[i] = "none"
			continue
		}
		parts[i] = a.String()
	}
	return " " + strings.Join(parts, " ")
}

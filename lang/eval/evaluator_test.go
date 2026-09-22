package eval

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/parser"
)

type recorder struct {
	calls []string
}

func (r *recorder) module() Module {
	rec := func(name string) Builtin {
		return NewBuiltin(Spec{Name: name, Variadic: true, Kwargs: []string{"k", "code"}}, func(ctx context.Context, args []Value, kwargs map[string]Value) (Value, error) {
			parts := make([]string, 0, len(args))
			for _, a := range args {
				parts = append(parts, fmt.Sprintf("%s:%s", TypeOf(a), a))
			}
			line := name + "(" + strings.Join(parts, ",") + ")"
			if v, ok := kwargs["k"]; ok {
				line += " k=" + v.String()
			}
			r.calls = append(r.calls, line)
			if name == "Exit" {
				code, _ := KwargInt(kwargs, "code", 0)
				return None(), &ExitError{Code: int(code), Msg: "exit"}
			}
			if name == "Fail" {
				return nil, errors.New("boom")
			}
			return None(), nil
		})
	}
	return NewModule("T", rec("F"), rec("Exit"), rec("Fail"))
}

func run(src string) (*recorder, error) {
	s, err := parser.New(lexer.New(src)).Parse()
	if err != nil {
		return nil, err
	}
	r := &recorder{}
	reg := NewRegistry()
	if err := reg.Register(r.module()); err != nil {
		return nil, err
	}
	return r, New(reg).Eval(context.Background(), s)
}

func TestEval(t *testing.T) {
	// --- literal conversion and call recording
	cases := []tu.Case[string, string]{
		{
			Name:     "string_int_bool_args",
			Input:    "T.F(\"a\", 1, true, false)\n",
			Expected: "F(string:a,int:1,bool:true,bool:false)",
		},
		{
			Name:     "kwargs_are_passed",
			Input:    "T.F(k=\"v\")\n",
			Expected: "F() k=v",
		},
		{
			Name:     "calls_run_in_order",
			Input:    "T.F(1)\nT.F(2)\n",
			Expected: "F(int:1)|F(int:2)",
		},
		{
			Name:     "escaped_string",
			Input:    "T.F(\"a\\nb\")\n",
			Expected: "F(string:a\nb)",
		},
	}
	fn := func(src string) (string, error) {
		r, err := run(src)
		if err != nil {
			return "", err
		}
		return strings.Join(r.calls, "|"), nil
	}
	tu.Run(tu.New(t), cases, fn, nil)
}

func TestEvalErrors(t *testing.T) {
	// --- error kinds
	cases := []tu.Case[string, int]{
		{
			Name:  "unknown_module",
			Input: "X.F()\n",
			Err:   ErrUnknownModule,
		},
		{
			Name:  "unknown_function",
			Input: "T.Nope()\n",
			Err:   ErrUnknownFunction,
		},
		{
			Name:     "exit_stops_evaluation",
			Input:    "T.F(1)\nT.Exit(code=4)\nT.F(2)\n",
			Expected: 2,
		},
		{
			Name:     "call_error_stops_evaluation",
			Input:    "T.F(1)\nT.Fail()\nT.F(2)\n",
			Expected: 2,
		},
	}
	fn := func(src string) (int, error) {
		r, err := run(src)
		var exit *ExitError
		if errors.As(err, &exit) {
			if exit.Code != 4 {
				return 0, fmt.Errorf("want exit code 4, got %d", exit.Code)
			}
			return len(r.calls), nil
		}
		if err != nil && r != nil && strings.Contains(err.Error(), "boom") {
			return len(r.calls), nil
		}
		return 0, err
	}
	tu.Run(tu.New(t), cases, fn, nil)
}

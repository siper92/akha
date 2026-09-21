package eval

import (
	"context"

	"github.com/siper92/akha/lang/ast"
)

type Type int

const (
	TypeNone Type = iota
	TypeString
	TypeInt
	TypeBool
	TypeList
)

type Value interface {
	Type() Type
	String() string
}

type Spec struct {
	Name     string
	MinArgs  int
	MaxArgs  int
	Variadic bool
	Kwargs   []string
}

type Builtin interface {
	Spec() Spec
	Call(ctx context.Context, args []Value, kwargs map[string]Value) (Value, error)
}

type Module interface {
	Name() string
	Func(name string) (Builtin, bool)
	Funcs() []string
}

type Registry interface {
	Register(m Module) error
	Lookup(name string) (Module, bool)
	Names() []string
}

type Evaluator interface {
	Eval(ctx context.Context, s *ast.Script) error
}

type ExitError struct {
	Code int
	Msg  string
}

var _ error = (*ExitError)(nil)

func (e *ExitError) Error() string { return e.Msg }

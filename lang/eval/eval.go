package eval

import (
	"context"
	"fmt"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/module"
)

type Kind int

const (
	KindNull Kind = iota
	KindBool
	KindNumber
	KindString
	KindArray
	KindObject
)

type Value interface {
	fmt.Stringer
	Kind() Kind
	Truth() bool
	Equal(Value) bool
	Clone() Value
}

type Comparer interface {
	Value
	Compare(Value) (int, error)
}

type Adder interface {
	Value
	Add(Value) (Value, error)
}

type OperationMinus interface {
	Value
	Minus(Value) (Value, error)
}

type OperationMul interface {
	Value
	Mul(Value) (Value, error)
}

type OperationDiv interface {
	Value
	Div(Value) (Value, error)
}

type OperationMod interface {
	Value
	Mod(Value) (Value, error)
}

type Negator interface {
	Value
	Neg() (Value, error)
}

type Container interface {
	Value
	Contains(Value) (bool, error)
}

type Indexer interface {
	Value
	Index(Value) (Value, error)
}

type IndexSetter interface {
	Indexer
	SetIndex(Value, Value) error
}

type Memberer interface {
	Value
	Member(string) (Value, error)
}

type MemberSetter interface {
	Memberer
	SetMember(string, Value) error
}

type Iterable interface {
	Value
	Len() int
	Items() []Pair
}

type Pair struct {
	Key   Value
	Value Value
}

type Sizer interface {
	Size() int
}

type Formatter interface {
	Format() string
}

type Operators interface {
	Binary(op lexer.Kind, l, r Value) (Value, error)
	Unary(op lexer.Kind, x Value) (Value, error)
}

type Binding struct {
	Value   Value
	Mutable bool
}

type Env interface {
	Parent() Env
	Child() Env
	Declare(name string, v Value, mutable bool) error
	Lookup(name string) (*Binding, bool)
	Assign(name string, v Value) error
}

type Flow int

const (
	FlowNext Flow = iota
	FlowBreak
	FlowContinue
	FlowReturn
)

type Options struct {
	File          string
	MaxStmts      int
	MaxIterations int
	MaxValueSize  int
	Modules       module.IRegistry
}

type Evaluator interface {
	Run(ctx context.Context, script *ast.Script, input Value) (Value, error)
}

type Interp interface {
	Exec(ctx context.Context, env Env, s ast.Stmt) (Flow, error)
	ExecBlock(ctx context.Context, env Env, b *ast.Block) (Flow, error)
	Eval(ctx context.Context, env Env, x ast.Expr) (Value, error)
}

type Target interface {
	Set(ctx context.Context, env Env, t ast.Expr, v Value) error
}

type Codec interface {
	FromGo(any) (Value, error)
	ToGo(Value) any
}

func (k Kind) String() string {
	switch k {
	case KindNull:
		return "null"
	case KindBool:
		return "boolean"
	case KindNumber:
		return "number"
	case KindString:
		return "string"
	case KindArray:
		return "array"
	case KindObject:
		return "object"
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

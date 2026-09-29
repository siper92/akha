package eval

import (
	"errors"

	"github.com/siper92/akha/lang/lexer"
)

var _ Operators = (*operators)(nil)

type operators struct{}

var verbs = map[lexer.Kind]string{
	lexer.Plus:    "add",
	lexer.Minus:   "subtract",
	lexer.Star:    "multiply",
	lexer.Slash:   "divide",
	lexer.Percent: "modulo",
	lexer.Lt:      "compare",
	lexer.LtEq:    "compare",
	lexer.Gt:      "compare",
	lexer.GtEq:    "compare",
	lexer.In:      "check membership of",
	lexer.NotIn:   "check membership of",
}

func NewOperators() Operators {
	return &operators{}
}

func (o *operators) Binary(op lexer.Kind, l, r Value) (Value, error) {
	switch op {
	case lexer.Eq:
		return Bool(l.Equal(r)), nil
	case lexer.NotEq:
		return Bool(!l.Equal(r)), nil
	case lexer.Lt, lexer.LtEq, lexer.Gt, lexer.GtEq:
		return o.compare(op, l, r)
	case lexer.In, lexer.NotIn:
		c, ok := r.(Container)
		if !ok {
			return nil, mismatch(op, l, r)
		}
		has, err := c.Contains(l)
		if err != nil {
			return nil, wrap(op, l, r, err)
		}
		return Bool(has == (op == lexer.In)), nil
	default:
		v, err := o.arith(op, l, r)
		if err != nil {
			return nil, wrap(op, l, r, err)
		}

		return v, nil
	}
}

func (o *operators) arith(op lexer.Kind, l, r Value) (Value, error) {
	switch op {
	case lexer.Plus:
		if x, ok := l.(Adder); ok {
			return x.Add(r)
		}
	case lexer.Minus:
		if x, ok := l.(OperationMinus); ok {
			return x.Minus(r)
		}
	case lexer.Star:
		if x, ok := l.(OperationMul); ok {
			return x.Mul(r)
		}
	case lexer.Slash:
		if x, ok := l.(OperationDiv); ok {
			return x.Div(r)
		}
	case lexer.Percent:
		if x, ok := l.(OperationMod); ok {
			return x.Mod(r)
		}
	default:
		return nil, newError(CodeInternal, "unknown operator %s", op)
	}
	return nil, errKindMismatch
}

func (o *operators) compare(op lexer.Kind, l, r Value) (Value, error) {
	c, ok := l.(Comparer)
	if !ok || l.Kind() != r.Kind() {
		return nil, newError(CodeOrdering, "cannot compare %s and %s", l.Kind(), r.Kind())
	}
	n, err := c.Compare(r)
	if err != nil {
		return nil, wrap(op, l, r, err)
	}
	switch op {
	case lexer.Lt:
		return Bool(n < 0), nil
	case lexer.LtEq:
		return Bool(n <= 0), nil
	case lexer.Gt:
		return Bool(n > 0), nil
	}
	return Bool(n >= 0), nil
}

func (o *operators) Unary(op lexer.Kind, x Value) (Value, error) {
	switch op {
	case lexer.Not:
		return Bool(!x.Truth()), nil
	case lexer.Minus:
		if n, ok := x.(Negator); ok {
			return n.Neg()
		}
		return nil, newError(CodeKindMismatch, "cannot negate %s", x.Kind())
	}
	return nil, newError(CodeInternal, "unknown unary operator %s", op)
}

func wrap(op lexer.Kind, l, r Value, err error) error {
	if errors.Is(err, errKindMismatch) {
		return mismatch(op, l, r)
	}
	return err
}

func mismatch(op lexer.Kind, l, r Value) error {
	return newError(CodeKindMismatch, "cannot %s %s and %s", verbs[op], l.Kind(), r.Kind())
}

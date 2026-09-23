package eval

import (
	"errors"
	"fmt"

	"github.com/siper92/akha/lang/token"
)

var (
	ErrOpType  = errors.New("operator type")
	ErrDivZero = errors.New("division by zero")
)

var opLits = map[token.Kind]string{
	token.PLUS:    "+",
	token.MINUS:   "-",
	token.STAR:    "*",
	token.SLASH:   "/",
	token.PERCENT: "%",
	token.EQ:      "==",
	token.NEQ:     "!=",
	token.LT:      "<",
	token.LTE:     "<=",
	token.GT:      ">",
	token.GTE:     ">=",
	token.AND:     "and",
	token.OR:      "or",
	token.NOT:     "not",
}

func Binary(op token.Kind, x, y Value) (Value, error) {
	switch op {
	case token.EQ:
		return Bool(Equal(x, y)), nil
	case token.NEQ:
		return Bool(!Equal(x, y)), nil
	case token.AND, token.OR:
		a, aok := AsBool(x)
		b, bok := AsBool(y)
		if !aok || !bok {
			return nil, binaryErr(op, x, y)
		}
		if op == token.AND {
			return Bool(a && b), nil
		}
		return Bool(a || b), nil
	case token.PLUS:
		if a, ok := AsString(x); ok {
			b, ok := AsString(y)
			if !ok {
				return nil, binaryErr(op, x, y)
			}
			return Str(a + b), nil
		}
		return arith(op, x, y)
	case token.MINUS, token.STAR, token.SLASH, token.PERCENT:
		return arith(op, x, y)
	case token.LT, token.LTE, token.GT, token.GTE:
		return compare(op, x, y)
	}
	return nil, fmt.Errorf("%w: unknown operator %s", ErrOpType, op)
}

func Unary(op token.Kind, x Value) (Value, error) {
	switch op {
	case token.MINUS:
		if i, ok := AsInt(x); ok {
			return Int(-i), nil
		}
	case token.NOT:
		if b, ok := AsBool(x); ok {
			return Bool(!b), nil
		}
	default:
		return nil, fmt.Errorf("%w: unknown operator %s", ErrOpType, op)
	}
	return nil, fmt.Errorf("%w: %s on %s", ErrOpType, opLit(op), TypeOf(x))
}

func Equal(x, y Value) bool {
	if TypeOf(x) != TypeOf(y) {
		return false
	}
	switch a := x.(type) {
	case ListValue:
		b := y.(ListValue)
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if !Equal(a[i], b[i]) {
				return false
			}
		}
		return true
	case nil, NoneValue:
		return true
	}
	return x == y
}

func arith(op token.Kind, x, y Value) (Value, error) {
	a, aok := AsInt(x)
	b, bok := AsInt(y)
	if !aok || !bok {
		return nil, binaryErr(op, x, y)
	}
	switch op {
	case token.PLUS:
		return Int(a + b), nil
	case token.MINUS:
		return Int(a - b), nil
	case token.STAR:
		return Int(a * b), nil
	case token.SLASH:
		if b == 0 {
			return nil, ErrDivZero
		}
		return Int(a / b), nil
	case token.PERCENT:
		if b == 0 {
			return nil, ErrDivZero
		}
		return Int(a % b), nil
	}
	return nil, binaryErr(op, x, y)
}

func compare(op token.Kind, x, y Value) (Value, error) {
	var c int
	switch a := x.(type) {
	case IntValue:
		b, ok := AsInt(y)
		if !ok {
			return nil, binaryErr(op, x, y)
		}
		c = cmpInt(int64(a), b)
	case StringValue:
		b, ok := AsString(y)
		if !ok {
			return nil, binaryErr(op, x, y)
		}
		c = cmpString(string(a), b)
	default:
		return nil, binaryErr(op, x, y)
	}
	switch op {
	case token.LT:
		return Bool(c < 0), nil
	case token.LTE:
		return Bool(c <= 0), nil
	case token.GT:
		return Bool(c > 0), nil
	}
	return Bool(c >= 0), nil
}

func cmpInt(a, b int64) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func cmpString(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

func binaryErr(op token.Kind, x, y Value) error {
	return fmt.Errorf("%w: %s on %s and %s", ErrOpType, opLit(op), TypeOf(x), TypeOf(y))
}

func opLit(op token.Kind) string {
	if lit, ok := opLits[op]; ok {
		return lit
	}
	return op.String()
}

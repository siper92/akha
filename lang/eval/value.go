package eval

import (
	"fmt"
	"strconv"
	"strings"
)

type StringValue string

type IntValue int64

type BoolValue bool

type ListValue []Value

type NoneValue struct{}

var (
	_ Value        = StringValue("")
	_ Value        = IntValue(0)
	_ Value        = BoolValue(false)
	_ Value        = ListValue(nil)
	_ Value        = NoneValue{}
	_ fmt.Stringer = Type(0)
)

var typeNames = [...]string{
	TypeNone:   "none",
	TypeString: "string",
	TypeInt:    "int",
	TypeBool:   "bool",
	TypeList:   "list",
}

func (t Type) String() string {
	if t >= 0 && int(t) < len(typeNames) {
		return typeNames[t]
	}
	return "type(" + strconv.Itoa(int(t)) + ")"
}

func Str(s string) Value { return StringValue(s) }

func Int(i int64) Value { return IntValue(i) }

func Bool(b bool) Value { return BoolValue(b) }

func List(vs ...Value) Value {
	out := make(ListValue, len(vs))
	copy(out, vs)
	return out
}

func None() Value { return NoneValue{} }

func (StringValue) Type() Type { return TypeString }

func (s StringValue) String() string { return string(s) }

func (IntValue) Type() Type { return TypeInt }

func (i IntValue) String() string { return strconv.FormatInt(int64(i), 10) }

func (BoolValue) Type() Type { return TypeBool }

func (b BoolValue) String() string { return strconv.FormatBool(bool(b)) }

func (ListValue) Type() Type { return TypeList }

func (l ListValue) String() string {
	parts := make([]string, len(l))
	for i, v := range l {
		parts[i] = valueString(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (NoneValue) Type() Type { return TypeNone }

func (NoneValue) String() string { return "none" }

func AsString(v Value) (string, bool) {
	s, ok := v.(StringValue)
	return string(s), ok
}

func AsInt(v Value) (int64, bool) {
	i, ok := v.(IntValue)
	return int64(i), ok
}

func AsBool(v Value) (bool, bool) {
	b, ok := v.(BoolValue)
	return bool(b), ok
}

func AsList(v Value) ([]Value, bool) {
	l, ok := v.(ListValue)
	return []Value(l), ok
}

func TypeOf(v Value) Type {
	if v == nil {
		return TypeNone
	}
	return v.Type()
}

func valueString(v Value) string {
	if v == nil {
		return "none"
	}
	return v.String()
}

package eval

import (
	"encoding/json"
	"math"
	"slices"
	"strconv"
	"strings"
)

var (
	_ Value = Null{}
	_ Value = Bool(false)

	_ Comparer       = Number(0)
	_ Adder          = Number(0)
	_ OperationMinus = Number(0)
	_ OperationMul   = Number(0)
	_ OperationDiv   = Number(0)
	_ OperationMod   = Number(0)
	_ Negator        = Number(0)

	_ Comparer  = String("")
	_ Adder     = String("")
	_ Container = String("")
	_ Formatter = String("")

	_ Adder       = (*Array)(nil)
	_ Container   = (*Array)(nil)
	_ IndexSetter = (*Array)(nil)
	_ Iterable    = (*Array)(nil)
	_ Sizer       = (*Array)(nil)

	_ Container    = (*Object)(nil)
	_ IndexSetter  = (*Object)(nil)
	_ MemberSetter = (*Object)(nil)
	_ Iterable     = (*Object)(nil)
	_ Sizer        = (*Object)(nil)
)

type Null struct{}

type Bool bool

type Number float64

type String string

type Array struct {
	Elems []Value
}

type Object struct {
	Keys   []string
	Fields map[string]Value
}

func NewArray(elems ...Value) *Array {
	return &Array{Elems: elems}
}

func NewObject() *Object {
	return &Object{Fields: map[string]Value{}}
}

func (Null) Kind() Kind          { return KindNull }
func (Null) Truth() bool         { return false }
func (Null) Equal(v Value) bool  { return v.Kind() == KindNull }
func (n Null) Clone() Value      { return n }
func (Null) String() string      { return "null" }
func (b Bool) Kind() Kind        { return KindBool }
func (b Bool) Truth() bool       { return bool(b) }
func (b Bool) Clone() Value      { return b }
func (b Bool) String() string    { return strconv.FormatBool(bool(b)) }
func (n Number) Kind() Kind      { return KindNumber }
func (n Number) Truth() bool     { return n != 0 }
func (n Number) Clone() Value    { return n }
func (s String) Kind() Kind      { return KindString }
func (s String) Truth() bool     { return s != "" }
func (s String) Clone() Value    { return s }
func (s String) Format() string  { return string(s) }
func (a *Array) Kind() Kind      { return KindArray }
func (a *Array) Truth() bool     { return len(a.Elems) > 0 }
func (a *Array) Len() int        { return len(a.Elems) }
func (o *Object) Kind() Kind     { return KindObject }
func (o *Object) Truth() bool    { return len(o.Keys) > 0 }
func (o *Object) Len() int       { return len(o.Keys) }
func (s String) String() string  { return quote(string(s)) }
func (a *Array) String() string  { return a.json() }
func (o *Object) String() string { return o.json() }

func (b Bool) Equal(v Value) bool {
	o, ok := v.(Bool)
	return ok && o == b
}

func (n Number) Equal(v Value) bool {
	o, ok := v.(Number)
	return ok && o == n
}

func (n Number) String() string {
	if n == 0 {
		return "0"
	}
	return strconv.FormatFloat(float64(n), 'f', -1, 64)
}

func (n Number) Compare(v Value) (int, error) {
	o, ok := v.(Number)
	if !ok {
		return 0, errKindMismatch
	}
	switch {
	case n < o:
		return -1, nil
	case n > o:
		return 1, nil
	}
	return 0, nil
}

func (n Number) Add(v Value) (Value, error) {
	return n.arith(v, func(a, b float64) float64 { return a + b })
}

func (n Number) Minus(v Value) (Value, error) {
	return n.arith(v, func(a, b float64) float64 { return a - b })
}

func (n Number) Mul(v Value) (Value, error) {
	return n.arith(v, func(a, b float64) float64 { return a * b })
}

func (n Number) Div(v Value) (Value, error) {
	if o, ok := v.(Number); ok && o == 0 {
		return nil, newError(CodeDivZero, "division by zero")
	}
	return n.arith(v, func(a, b float64) float64 { return a / b })
}

func (n Number) Mod(v Value) (Value, error) {
	if o, ok := v.(Number); ok && o == 0 {
		return nil, newError(CodeDivZero, "modulo by zero")
	}
	return n.arith(v, math.Mod)
}

func (n Number) Neg() (Value, error) {
	return -n, nil
}

func (n Number) arith(v Value, fn func(a, b float64) float64) (Value, error) {
	o, ok := v.(Number)
	if !ok {
		return nil, errKindMismatch
	}
	return checkNumber(fn(float64(n), float64(o)))
}

func checkNumber(f float64) (Value, error) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, newError(CodeNumberRange, "number out of range")
	}
	if f == 0 {
		f = 0
	}
	return Number(f), nil
}

func (s String) Equal(v Value) bool {
	o, ok := v.(String)
	return ok && o == s
}

func (s String) Compare(v Value) (int, error) {
	o, ok := v.(String)
	if !ok {
		return 0, errKindMismatch
	}
	return strings.Compare(string(s), string(o)), nil
}

func (s String) Add(v Value) (Value, error) {
	o, ok := v.(String)
	if !ok {
		return nil, errKindMismatch
	}
	return s + o, nil
}

func (s String) Contains(v Value) (bool, error) {
	o, ok := v.(String)
	if !ok {
		return false, errKindMismatch
	}
	return strings.Contains(string(s), string(o)), nil
}

func (a *Array) Equal(v Value) bool {
	o, ok := v.(*Array)
	if !ok || len(o.Elems) != len(a.Elems) {
		return false
	}
	for i, e := range a.Elems {
		if !e.Equal(o.Elems[i]) {
			return false
		}
	}
	return true
}

func (a *Array) Clone() Value {
	elems := make([]Value, len(a.Elems))
	for i, e := range a.Elems {
		elems[i] = e.Clone()
	}
	return &Array{Elems: elems}
}

func (a *Array) Add(v Value) (Value, error) {
	o, ok := v.(*Array)
	if !ok {
		return nil, errKindMismatch
	}
	return &Array{Elems: slices.Concat(a.Clone().(*Array).Elems, o.Clone().(*Array).Elems)}, nil
}

func (a *Array) Contains(v Value) (bool, error) {
	return slices.ContainsFunc(a.Elems, v.Equal), nil
}

func (a *Array) Index(i Value) (Value, error) {
	at, err := a.at(i)
	if err != nil {
		return nil, err
	}
	return a.Elems[at], nil
}

func (a *Array) SetIndex(i, v Value) error {
	at, err := a.at(i)
	if err != nil {
		return err
	}
	a.Elems[at] = v
	return nil
}

func (a *Array) at(i Value) (int, error) {
	n, ok := i.(Number)
	if !ok {
		return 0, newError(CodeIndexKind, "array index must be a number, got %s", i.Kind())
	}
	if float64(n) != math.Trunc(float64(n)) {
		return 0, newError(CodeNonIntegral, "array index %s is not integral", n)
	}
	if n < 0 || int(n) >= len(a.Elems) {
		return 0, newError(CodeIndexRange, "index %s out of range for length %d", n, len(a.Elems))
	}
	return int(n), nil
}

func (a *Array) Items() []Pair {
	items := make([]Pair, len(a.Elems))
	for i, e := range a.Elems {
		items[i] = Pair{Key: Number(i), Value: e}
	}
	return items
}

func (a *Array) Size() int {
	size := len(a.Elems)
	for _, e := range a.Elems {
		if s, ok := e.(Sizer); ok {
			size += s.Size()
		}
	}
	return size
}

func (a *Array) json() string {
	items := make([]string, len(a.Elems))
	for i, e := range a.Elems {
		items[i] = e.String()
	}
	return "[" + strings.Join(items, ",") + "]"
}

func (o *Object) Equal(v Value) bool {
	other, ok := v.(*Object)
	if !ok || len(other.Keys) != len(o.Keys) {
		return false
	}
	for k, f := range o.Fields {
		g, ok := other.Fields[k]
		if !ok || !f.Equal(g) {
			return false
		}
	}
	return true
}

func (o *Object) Clone() Value {
	c := &Object{Keys: slices.Clone(o.Keys), Fields: make(map[string]Value, len(o.Fields))}
	for k, f := range o.Fields {
		c.Fields[k] = f.Clone()
	}
	return c
}

func (o *Object) Contains(v Value) (bool, error) {
	k, ok := v.(String)
	if !ok {
		return false, errKindMismatch
	}
	_, ok = o.Fields[string(k)]
	return ok, nil
}

func (o *Object) Index(k Value) (Value, error) {
	s, ok := k.(String)
	if !ok {
		return nil, newError(CodeIndexKind, "object key must be a string, got %s", k.Kind())
	}
	return o.Member(string(s))
}

func (o *Object) SetIndex(k, v Value) error {
	s, ok := k.(String)
	if !ok {
		return newError(CodeIndexKind, "object key must be a string, got %s", k.Kind())
	}
	return o.SetMember(string(s), v)
}

func (o *Object) Member(name string) (Value, error) {
	if f, ok := o.Fields[name]; ok {
		return f, nil
	}
	return Null{}, nil
}

func (o *Object) SetMember(name string, v Value) error {
	if _, ok := o.Fields[name]; !ok {
		o.Keys = append(o.Keys, name)
	}
	o.Fields[name] = v
	return nil
}

func (o *Object) Items() []Pair {
	items := make([]Pair, len(o.Keys))
	for i, k := range o.Keys {
		items[i] = Pair{Key: String(k), Value: o.Fields[k]}
	}
	return items
}

func (o *Object) Size() int {
	size := len(o.Keys)
	for _, f := range o.Fields {
		if s, ok := f.(Sizer); ok {
			size += s.Size()
		}
	}
	return size
}

func (o *Object) json() string {
	items := make([]string, len(o.Keys))
	for i, k := range o.Keys {
		items[i] = quote(k) + ":" + o.Fields[k].String()
	}
	return "{" + strings.Join(items, ",") + "}"
}

func format(v Value) string {
	if f, ok := v.(Formatter); ok {
		return f.Format()
	}
	return v.String()
}

func quote(s string) string {
	var b strings.Builder
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimSuffix(b.String(), "\n")
}

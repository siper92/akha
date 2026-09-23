package eval

import (
	"errors"
	"fmt"
)

var (
	ErrUndefined  = errors.New("undefined variable")
	ErrRedeclared = errors.New("variable already declared")
)

type env struct {
	parent Env
	vars   map[string]Value
}

var _ Env = (*env)(nil)

func NewEnv(parent Env) Env {
	return &env{parent: parent, vars: map[string]Value{}}
}

func (e *env) Lookup(name string) (Value, bool) {
	if v, ok := e.vars[name]; ok {
		return v, true
	}
	if e.parent != nil {
		return e.parent.Lookup(name)
	}
	return nil, false
}

func (e *env) Define(name string, v Value) error {
	if _, ok := e.vars[name]; ok {
		return fmt.Errorf("%w: %s", ErrRedeclared, name)
	}
	e.vars[name] = v
	return nil
}

func (e *env) Assign(name string, v Value) error {
	if _, ok := e.vars[name]; ok {
		e.vars[name] = v
		return nil
	}
	if e.parent != nil {
		return e.parent.Assign(name, v)
	}
	return fmt.Errorf("%w: %s", ErrUndefined, name)
}

func (e *env) Child() Env {
	return NewEnv(e)
}

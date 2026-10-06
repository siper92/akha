package eval

var _ Env = (*env)(nil)

type env struct {
	parent *env
	names  map[string]*Binding
}

func NewEnv() Env {
	return &env{names: map[string]*Binding{}}
}

func (e *env) Parent() Env {
	if e.parent == nil {
		return nil
	}
	return e.parent
}

func (e *env) Child() Env {
	return &env{parent: e, names: map[string]*Binding{}}
}

func (e *env) Declare(name string, v Value, mutable bool) error {
	if name == "_" {
		return nil
	}

	if _, ok := e.names[name]; ok {
		return newError(CodeRedeclared, "%s is already declared in this block", name)
	}

	e.names[name] = &Binding{Value: v, Mutable: mutable}
	return nil
}

func (e *env) Lookup(name string) (*Binding, bool) {
	for s := e; s != nil; s = s.parent {
		if b, ok := s.names[name]; ok {
			return b, true
		}
	}

	return nil, false
}

func (e *env) Assign(name string, v Value) error {
	b, ok := e.Lookup(name)
	if !ok {
		return newError(CodeUndeclared, "%s is not declared", name)
	}

	if !b.Mutable {
		return newError(CodeImmutable, "cannot assign to %s", name)
	}
	b.Value = v

	return nil
}

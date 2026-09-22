package eval

import "context"

type CallFunc func(ctx context.Context, args []Value, kwargs map[string]Value) (Value, error)

type builtin struct {
	spec Spec
	call CallFunc
}

var _ Builtin = (*builtin)(nil)

func NewBuiltin(spec Spec, call CallFunc) Builtin {
	return &builtin{spec: spec, call: call}
}

func (b *builtin) Spec() Spec { return b.spec }

func (b *builtin) Call(ctx context.Context, args []Value, kwargs map[string]Value) (Value, error) {
	return b.call(ctx, args, kwargs)
}

type module struct {
	name  string
	funcs map[string]Builtin
	order []string
}

var _ Module = (*module)(nil)

func NewModule(name string, funcs ...Builtin) Module {
	m := &module{name: name, funcs: make(map[string]Builtin, len(funcs))}
	for _, f := range funcs {
		n := f.Spec().Name
		m.funcs[n] = f
		m.order = append(m.order, n)
	}
	return m
}

func (m *module) Name() string { return m.name }

func (m *module) Func(name string) (Builtin, bool) {
	f, ok := m.funcs[name]
	return f, ok
}

func (m *module) Funcs() []string {
	out := make([]string, len(m.order))
	copy(out, m.order)
	return out
}

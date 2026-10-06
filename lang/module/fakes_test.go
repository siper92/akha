package module

import (
	"fmt"
	"slices"

	"github.com/siper92/akha/lang/ast"
)

var (
	_ IValue      = (*fakeValue)(nil)
	_ IType       = (*fakeType)(nil)
	_ IModuleFunc = (*fakeFunc)(nil)
	_ ILoader     = (*fakeLoader)(nil)
	_ IModule     = (*fakeModule)(nil)
	_ IRegistry   = (*fakeRegistry)(nil)
	_ IScope      = (*fakeScope)(nil)
)

type fakeValue struct {
	v float64
}

func (f *fakeValue) String() string { return fmt.Sprint(f.v) }
func (f *fakeValue) Truth() bool    { return f.v != 0 }

type fakeType struct{}

func (fakeType) Name() string { return "number" }

func (fakeType) Accepts(known ast.Expr) bool {
	_, ok := known.(*ast.NumberLit)
	return ok
}

func (fakeType) Valid(v IValue) bool {
	_, ok := v.(*fakeValue)
	return ok
}

type fakeFunc struct {
	name   string
	params []Param
	kwargs []Param
	impl   func(args ...IValue) (IValue, error)
}

func (f *fakeFunc) Name() string    { return f.name }
func (f *fakeFunc) Params() []Param { return f.params }
func (f *fakeFunc) Kwargs() []Param { return f.kwargs }

func (f *fakeFunc) Kwarg(name string) (Param, bool) {
	i := slices.IndexFunc(f.kwargs, func(p Param) bool { return p.Name == name })
	if i < 0 {
		return Param{}, false
	}
	return f.kwargs[i], true
}

func (f *fakeFunc) Bind(args []IValue, kwargs map[string]IValue) ([]IValue, error) {
	if len(args) != len(f.params) {
		return nil, fmt.Errorf("%s: %w", f.name, ErrArity)
	}
	for i, p := range f.params {
		if !p.Type.Valid(args[i]) {
			return nil, fmt.Errorf("%s: %w", p.Name, ErrArgKind)
		}
	}
	for name := range kwargs {
		if _, ok := f.Kwarg(name); !ok {
			return nil, fmt.Errorf("%s: %w", name, ErrUnknownKwarg)
		}
	}

	bound := slices.Clone(args)
	for _, p := range f.kwargs {
		v, ok := kwargs[p.Name]
		switch {
		case ok:
		case p.Default != nil:
			v = p.Default
		default:
			return nil, fmt.Errorf("%s: %w", p.Name, ErrNoDefault)
		}
		bound = append(bound, v)
	}
	return bound, nil
}

func (f *fakeFunc) Call(args ...IValue) (IValue, error) {
	if f.impl == nil {
		return nil, fmt.Errorf("%s: %w", f.name, ErrNoImpl)
	}
	return f.impl(args...)
}

type fakeLoader struct {
	fakeFunc
}

func (f *fakeLoader) Load(scope IScope, name string) error {
	return scope.Import(name)
}

type fakeModule struct {
	name  string
	funcs []IModuleFunc
}

func (m *fakeModule) Name() string         { return m.name }
func (m *fakeModule) Funcs() []IModuleFunc { return m.funcs }

func (m *fakeModule) Func(name string) (IModuleFunc, bool) {
	i := slices.IndexFunc(m.funcs, func(f IModuleFunc) bool { return f.Name() == name })
	if i < 0 {
		return nil, false
	}
	return m.funcs[i], true
}

type fakeRegistry struct {
	modules map[string]IModule
}

func newFakeRegistry() *fakeRegistry {
	return &fakeRegistry{modules: make(map[string]IModule)}
}

func (r *fakeRegistry) Register(m IModule) error {
	if _, ok := r.modules[m.Name()]; ok {
		return fmt.Errorf("%s: %w", m.Name(), ErrModuleExists)
	}
	r.modules[m.Name()] = m
	return nil
}

func (r *fakeRegistry) Module(name string) (IModule, bool) {
	m, ok := r.modules[name]
	return m, ok
}

func (r *fakeRegistry) Lookup(module, fn string) (IModuleFunc, error) {
	m, ok := r.Module(module)
	if !ok {
		return nil, fmt.Errorf("%s: %w", module, ErrUnknownModule)
	}
	f, ok := m.Func(fn)
	if !ok {
		return nil, fmt.Errorf("%s.%s: %w", module, fn, ErrUnknownFunc)
	}
	return f, nil
}

type fakeScope struct {
	reg      IRegistry
	imported map[string]bool
}

func newFakeScope(reg IRegistry) *fakeScope {
	return &fakeScope{reg: reg, imported: map[string]bool{Core: true}}
}

func (s *fakeScope) Import(name string) error {
	if _, ok := s.reg.Module(name); !ok {
		return fmt.Errorf("%s: %w", name, ErrUnknownModule)
	}
	if s.imported[name] {
		return fmt.Errorf("%s: %w", name, ErrImported)
	}
	s.imported[name] = true
	return nil
}

func (s *fakeScope) Imported(name string) bool {
	return s.imported[name]
}

func (s *fakeScope) Callable(module, fn string) (IModuleFunc, error) {
	if !s.Imported(module) {
		return nil, fmt.Errorf("%s: %w", module, ErrNotImported)
	}
	return s.reg.Lookup(module, fn)
}

func num(v float64) IValue {
	return &fakeValue{v: v}
}

func fixture() (*fakeRegistry, *fakeFunc, *fakeLoader) {
	add := &fakeFunc{
		name:   "Add",
		params: []Param{{Name: "a", Type: fakeType{}}, {Name: "b", Type: fakeType{}}},
		kwargs: []Param{{Name: "scale", Type: fakeType{}, Default: num(1)}},
		impl: func(args ...IValue) (IValue, error) {
			a, b, s := args[0].(*fakeValue), args[1].(*fakeValue), args[2].(*fakeValue)
			return num((a.v + b.v) * s.v), nil
		},
	}
	allow := &fakeLoader{fakeFunc: fakeFunc{name: "Allow"}}

	reg := newFakeRegistry()
	_ = reg.Register(&fakeModule{name: Core, funcs: []IModuleFunc{allow}})
	_ = reg.Register(&fakeModule{name: "math", funcs: []IModuleFunc{add}})
	return reg, add, allow
}

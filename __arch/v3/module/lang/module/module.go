package module

import (
	"fmt"
	"slices"

	"github.com/siper92/akha/lang/ast"
)

var (
	_ IModuleFunc = (*Func)(nil)
	_ ILoader     = (*Loader)(nil)
	_ IModule     = (*Module)(nil)
	_ IRegistry   = (*Registry)(nil)
	_ IScope      = (*Scope)(nil)
)

type Impl func(args ...IValue) (IValue, error)

type Func struct {
	name   string
	params []Param
	kwargs []Param
	impl   Impl
}

type Loader struct {
	*Func
}

type Module struct {
	name  string
	funcs []IModuleFunc
}

type Registry struct {
	modules map[string]IModule
}

type Scope struct {
	reg      IRegistry
	imported map[string]bool
}

func NewFunc(name string, params, kwargs []Param, impl Impl) (*Func, error) {
	for _, p := range kwargs {
		if p.Default == nil {
			return nil, fmt.Errorf("%s(%s): %w", name, p.Name, ErrNoDefault)
		}
	}
	return &Func{name: name, params: params, kwargs: kwargs, impl: impl}, nil
}

func (f *Func) Name() string    { return f.name }
func (f *Func) Params() []Param { return f.params }
func (f *Func) Kwargs() []Param { return f.kwargs }

func (f *Func) Kwarg(name string) (Param, bool) {
	i := slices.IndexFunc(f.kwargs, func(p Param) bool { return p.Name == name })
	if i < 0 {
		return Param{}, false
	}

	return f.kwargs[i], true
}

func (f *Func) Bind(args []IValue, kwargs map[string]IValue) ([]IValue, error) {
	if len(args) != len(f.params) {
		return nil, fmt.Errorf("%s: %w, want %d, got %d", f.name, ErrArity, len(f.params), len(args))
	}

	for i, p := range f.params {
		if !p.Type.Valid(args[i]) {
			return nil, fmt.Errorf("%s(%s): %w, want %s", f.name, p.Name, ErrArgKind, p.Type.Name())
		}
	}
	for name := range kwargs {
		if _, ok := f.Kwarg(name); !ok {
			return nil, fmt.Errorf("%s(%s): %w", f.name, name, ErrUnknownKwarg)
		}
	}

	bound := slices.Clone(args)
	for _, p := range f.kwargs {
		v, ok := kwargs[p.Name]
		if !ok {
			v = p.Default
		}
		if !p.Type.Valid(v) {
			return nil, fmt.Errorf("%s(%s): %w, want %s", f.name, p.Name, ErrArgKind, p.Type.Name())
		}
		bound = append(bound, v)
	}

	return bound, nil
}

func (f *Func) Call(args ...IValue) (IValue, error) {
	if f.impl == nil {
		return nil, fmt.Errorf("%s: %w", f.name, ErrNoImpl)
	}

	return f.impl(args...)
}

func NewLoader(name string, param Param) (*Loader, error) {
	f, err := NewFunc(name, []Param{param}, nil, nil)
	if err != nil {
		return nil, err
	}

	return &Loader{Func: f}, nil
}

func (l *Loader) Load(scope IScope, name string) error {
	return scope.Import(name)
}

func NewModule(name string, funcs ...IModuleFunc) (*Module, error) {
	m := &Module{name: name}
	for _, f := range funcs {
		if _, ok := m.Func(f.Name()); ok {
			return nil, fmt.Errorf("%s.%s: %w", name, f.Name(), ErrFuncExists)
		}
		m.funcs = append(m.funcs, f)
	}

	return m, nil
}

func (m *Module) Name() string         { return m.name }
func (m *Module) Funcs() []IModuleFunc { return m.funcs }

func (m *Module) Func(name string) (IModuleFunc, bool) {
	i := slices.IndexFunc(m.funcs, func(f IModuleFunc) bool { return f.Name() == name })
	if i < 0 {
		return nil, false
	}

	return m.funcs[i], true
}

func NewRegistry(modules ...IModule) (*Registry, error) {
	r := &Registry{modules: make(map[string]IModule)}
	for _, m := range modules {
		if err := r.Register(m); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Registry) Register(m IModule) error {
	if _, ok := r.modules[m.Name()]; ok {
		return fmt.Errorf("%s: %w", m.Name(), ErrModuleExists)
	}
	r.modules[m.Name()] = m
	return nil
}

func (r *Registry) Module(name string) (IModule, bool) {
	m, ok := r.modules[name]
	return m, ok
}

func (r *Registry) Lookup(module, fn string) (IModuleFunc, error) {
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

func NewScope(reg IRegistry) *Scope {
	return &Scope{reg: reg, imported: map[string]bool{Core: true}}
}

func (s *Scope) Import(name string) error {
	if _, ok := s.reg.Module(name); !ok {
		return fmt.Errorf("%s: %w", name, ErrUnknownModule)
	}
	if s.imported[name] {
		return fmt.Errorf("%s: %w", name, ErrImported)
	}
	s.imported[name] = true
	return nil
}

func (s *Scope) Imported(name string) bool {
	return s.imported[name]
}

func (s *Scope) Callable(module, fn string) (IModuleFunc, error) {
	if !s.Imported(module) {
		return nil, fmt.Errorf("%s: %w", module, ErrNotImported)
	}
	return s.reg.Lookup(module, fn)
}

func Callee(call *ast.CallExpr) (string, string, bool) {
	m, ok := call.Fn.(*ast.MemberExpr)
	if !ok {
		return "", "", false
	}
	id, ok := m.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}
	return id.Name, m.Name, true
}

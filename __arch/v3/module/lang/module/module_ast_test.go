package module

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/tests_utils"
)

type lookup struct {
	module, fn string
}

type bindInput struct {
	args   []IValue
	kwargs map[string]IValue
}

func TestRegistryLookup(t *testing.T) {
	cases := []tests_utils.Case[lookup, string]{
		// --- registered functions resolve by module and name
		{
			Name:     "known_func",
			Input:    lookup{module: "math", fn: "Add"},
			Expected: "Add",
		},
		// --- missing modules and functions
		{
			Name:  "unknown_module",
			Input: lookup{module: "http", fn: "Get"},
			Err:   ErrUnknownModule,
		},
		{
			Name:  "unknown_func",
			Input: lookup{module: "math", fn: "Sub"},
			Err:   ErrUnknownFunc,
		},
	}

	reg, _, _ := fixture()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			f, err := reg.Lookup(c.Input.module, c.Input.fn)
			if c.Err != nil {
				if !errors.Is(err, c.Err) {
					t.Fatalf("got %v, want %v", err, c.Err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.Name() != c.Expected {
				t.Errorf("got %q, want %q", f.Name(), c.Expected)
			}
		})
	}
}

func TestRegistryRegister(t *testing.T) {
	reg, _, _ := fixture()
	err := reg.Register(&fakeModule{name: "math"})
	if !errors.Is(err, ErrModuleExists) {
		t.Errorf("got %v, want %v", err, ErrModuleExists)
	}
}

func TestScopeCallable(t *testing.T) {
	cases := []tests_utils.Case[lookup, string]{
		// --- the core module is always imported
		{
			Name:     "core_without_import",
			Input:    lookup{module: Core, fn: "Allow"},
			Expected: "Allow",
		},
		// --- other modules need an import first
		{
			Name:  "module_not_imported",
			Input: lookup{module: "math", fn: "Add"},
			Err:   ErrNotImported,
		},
	}

	reg, _, _ := fixture()
	scope := newFakeScope(reg)
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			f, err := scope.Callable(c.Input.module, c.Input.fn)
			if c.Err != nil {
				if !errors.Is(err, c.Err) {
					t.Fatalf("got %v, want %v", err, c.Err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.Name() != c.Expected {
				t.Errorf("got %q, want %q", f.Name(), c.Expected)
			}
		})
	}
}

func TestLoaderLoad(t *testing.T) {
	cases := []tests_utils.Case[[]string, struct{}]{
		// --- a loader imports modules into the scope
		{
			Name:  "load_known_module",
			Input: []string{"math"},
		},
		// --- import errors
		{
			Name:  "load_twice",
			Input: []string{"math", "math"},
			Err:   ErrImported,
		},
		{
			Name:  "load_unknown_module",
			Input: []string{"http"},
			Err:   ErrUnknownModule,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			reg, _, allow := fixture()
			scope := newFakeScope(reg)

			var err error
			for _, name := range c.Input {
				if err = allow.Load(scope, name); err != nil {
					break
				}
			}

			if c.Err != nil {
				if !errors.Is(err, c.Err) {
					t.Fatalf("got %v, want %v", err, c.Err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if _, err := scope.Callable("math", "Add"); err != nil {
				t.Errorf("unexpected error after load: %v", err)
			}
		})
	}
}

func TestFuncBindAndCall(t *testing.T) {
	cases := []tests_utils.Case[bindInput, string]{
		// --- named params fall back to their default
		{
			Name:     "default_kwarg",
			Input:    bindInput{args: []IValue{num(1), num(2)}},
			Expected: "3",
		},
		{
			Name:     "explicit_kwarg",
			Input:    bindInput{args: []IValue{num(1), num(2)}, kwargs: map[string]IValue{"scale": num(10)}},
			Expected: "30",
		},
		// --- bind errors
		{
			Name:  "wrong_arity",
			Input: bindInput{args: []IValue{num(1)}},
			Err:   ErrArity,
		},
		{
			Name:  "unknown_kwarg",
			Input: bindInput{args: []IValue{num(1), num(2)}, kwargs: map[string]IValue{"x": num(1)}},
			Err:   ErrUnknownKwarg,
		},
	}

	_, add, _ := fixture()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			args, err := add.Bind(c.Input.args, c.Input.kwargs)
			if c.Err != nil {
				if !errors.Is(err, c.Err) {
					t.Fatalf("got %v, want %v", err, c.Err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected bind error: %v", err)
			}

			v, err := add.Call(args...)
			if err != nil {
				t.Fatalf("unexpected call error: %v", err)
			}

			if v.String() != c.Expected {
				t.Errorf("got %q, want %q", v.String(), c.Expected)
			}
		})
	}
}

func TestFuncNoImpl(t *testing.T) {
	_, _, allow := fixture()
	if _, err := allow.Call(); !errors.Is(err, ErrNoImpl) {
		t.Errorf("got %v, want %v", err, ErrNoImpl)
	}
}

func TestTypeAccepts(t *testing.T) {
	cases := []tests_utils.Case[ast.Expr, bool]{
		// --- a type accepts matching known values
		{
			Name:     "number_literal",
			Input:    &ast.NumberLit{Value: 1, Raw: "1"},
			Expected: true,
		},
		{
			Name:     "string_literal",
			Input:    &ast.StringLit{Value: "a"},
			Expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := (fakeType{}).Accepts(c.Input); got != c.Expected {
				t.Errorf("got %v, want %v", got, c.Expected)
			}
		})
	}
}

package module

import (
	"errors"
	"fmt"

	"github.com/siper92/akha/lang/ast"
)

const Core = "ak"

var (
	ErrModuleExists  = errors.New("module already registered")
	ErrFuncExists    = errors.New("function already defined")
	ErrUnknownModule = errors.New("unknown module")
	ErrUnknownFunc   = errors.New("unknown function")
	ErrNotImported   = errors.New("module not imported")
	ErrImported      = errors.New("module already imported")
	ErrNoDefault     = errors.New("named param without default")
	ErrArity         = errors.New("wrong number of arguments")
	ErrUnknownKwarg  = errors.New("unknown named argument")
	ErrArgKind       = errors.New("wrong argument kind")
	ErrNoImpl        = errors.New("function has no implementation")
)

type IValue interface {
	fmt.Stringer
	Truth() bool
}

type ICallable interface {
	Call(args ...IValue) (IValue, error)
}

type IType interface {
	Name() string
	Accepts(known ast.Expr) bool
	Valid(v IValue) bool
}

type Param struct {
	Name    string
	Type    IType
	Default IValue
}

type IModuleFunc interface {
	ICallable
	Name() string
	Params() []Param
	Kwargs() []Param
	Kwarg(name string) (Param, bool)
	Bind(args []IValue, kwargs map[string]IValue) ([]IValue, error)
}

type ILoader interface {
	IModuleFunc
	Load(scope IScope, name string) error
}

type IModule interface {
	Name() string
	Func(name string) (IModuleFunc, bool)
	Funcs() []IModuleFunc
}

type IRegistry interface {
	Register(m IModule) error
	Module(name string) (IModule, bool)
	Lookup(module, fn string) (IModuleFunc, error)
}

type IScope interface {
	Import(name string) error
	Imported(name string) bool
	Callable(module, fn string) (IModuleFunc, error)
}

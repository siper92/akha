package ak

import (
	"log/slog"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
	"github.com/siper92/akha/lang/module/types"
)

const (
	Name   = "ak"
	Import = "import"
	Setup  = "setup"
	Log    = "log"
	Debug  = "debug"
)

type Module interface {
	Import(args ...module.IValue) (module.IValue, error)
	Setup(args ...module.IValue) (module.IValue, error)
	Log(args ...module.IValue) (module.IValue, error)
	Debug(args ...module.IValue) (module.IValue, error)
}

var (
	_ module.IModule = (*akModule)(nil)
	_ Module         = (*akModule)(nil)
)

type akModule struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) (module.IModule, error) {
	return &akModule{logger: logger}, nil
}

func (a akModule) Import(args ...module.IValue) (module.IValue, error) {
	//TODO implement me
	panic("implement me")
}

func (a akModule) Setup(args ...module.IValue) (module.IValue, error) {
	//TODO implement me
	panic("implement me")
}

func (a akModule) Log(args ...module.IValue) (module.IValue, error) {
	//TODO implement me
	panic("implement me")
}

func (a akModule) Debug(args ...module.IValue) (module.IValue, error) {
	//TODO implement me
	panic("implement me")
}

func (a akModule) Name() string {
	return Name
}

func (a akModule) Func(name string) (module.IModuleFunc, bool) {
	switch name {
	case Import:
		funcDed, err := module.NewFunc(Import, []module.Param{
			{Name: "name", Type: types.String{}},
		}, nil, a.Import)
		if err != nil {
			return nil, false
		}

		return funcDed, true
	case Setup:
		funcDed, err := module.NewFunc(Setup, nil, []module.Param{
			{Name: "log", Type: types.String{}, Default: eval.String("")},
			{Name: "debug", Type: types.String{}, Default: eval.String("")},
		}, a.Setup)
		if err != nil {
			return nil, false
		}
		return funcDed, true
	case Log:
		funcDed, err := module.NewFunc(Log, []module.Param{{Name: "msg", Type: types.String{}}}, nil, a.Log)
		if err != nil {
			return nil, false
		}
		return funcDed, true
	case Debug:
		funcDed, err := module.NewFunc(Debug, []module.Param{{Name: "msg", Type: types.String{}}}, nil, a.Debug)
		if err != nil {
			return nil, false
		}
		return funcDed, true
	default:
		return nil, false
	}
}

func (a akModule) Funcs() []module.IModuleFunc {
	var funcs []module.IModuleFunc
	if f, ok := a.Func(Import); ok {
		funcs = append(funcs, f)
	}
	if f, ok := a.Func(Setup); ok {
		funcs = append(funcs, f)
	}
	if f, ok := a.Func(Log); ok {
		funcs = append(funcs, f)
	}
	if f, ok := a.Func(Debug); ok {
		funcs = append(funcs, f)
	}
	return funcs
}

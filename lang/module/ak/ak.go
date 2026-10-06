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

func (a *akModule) Setup(_ ...module.IValue) (module.IValue, error) {
	return eval.Null{}, nil
}

func (a *akModule) Log(args ...module.IValue) (module.IValue, error) {
	a.logger.Info(msg(args[0]))
	return eval.Null{}, nil
}

func (a *akModule) Debug(args ...module.IValue) (module.IValue, error) {
	a.logger.Debug(msg(args[0]))
	return eval.Null{}, nil
}

func msg(v module.IValue) string {
	if s, ok := v.(eval.String); ok {
		return string(s)
	}
	return v.String()
}

func (a *akModule) Name() string {
	return Name
}

func (a *akModule) Func(name string) (module.IModuleFunc, bool) {
	var (
		f   module.IModuleFunc
		err error
	)

	switch name {
	case Import:
		f, err = module.NewLoader(Import, module.Param{Name: "name", Type: types.String{}})
	case Setup:
		f, err = module.NewFunc(Setup, nil, []module.Param{
			{Name: "log", Type: types.String{}, Default: eval.String("")},
			{Name: "debug", Type: types.String{}, Default: eval.String("")},
		}, a.Setup)
	case Log:
		f, err = module.NewFunc(Log, []module.Param{{Name: "msg", Type: types.Any{}}}, nil, a.Log)
	case Debug:
		f, err = module.NewFunc(Debug, []module.Param{{Name: "msg", Type: types.Any{}}}, nil, a.Debug)
	default:
		return nil, false
	}

	if err != nil {
		return nil, false
	}

	return f, true
}

func (a *akModule) Funcs() []module.IModuleFunc {
	var funcs []module.IModuleFunc
	for _, name := range []string{Import, Setup, Log, Debug} {
		if f, ok := a.Func(name); ok {
			funcs = append(funcs, f)
		}
	}

	return funcs
}

package fs

import (
	"errors"
	iofs "io/fs"
	"os"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
	"github.com/siper92/akha/lang/module/types"
)

const (
	Name   = "fs"
	Read   = "read"
	Write  = "write"
	Exists = "exists"
)

type Files interface {
	read(args ...module.IValue) (module.IValue, error)
	write(args ...module.IValue) (module.IValue, error)
	exists(args ...module.IValue) (module.IValue, error)
}

var (
	_ Files          = (*files)(nil)
	_ module.IModule = (*files)(nil)
)

type files struct {
	root string
}

func New(root string) (module.IModule, error) {
	return &files{root: root}, nil
}

func (f *files) read(args ...module.IValue) (module.IValue, error) {
	root, err := os.OpenRoot(f.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	b, err := root.ReadFile(str(args[0]))
	if err != nil {
		return nil, err
	}
	return eval.String(b), nil
}

func (f *files) write(args ...module.IValue) (module.IValue, error) {
	root, err := os.OpenRoot(f.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	fileName := str(args[0])
	val := str(args[1])
	if fileName == "" {
		return nil, errors.New("file name cannot be empty")
	}

	if _, err = root.Stat(fileName); err == nil {
		fR, err := root.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		defer fR.Close()

		if _, err = fR.Write([]byte(val)); err != nil {
			return nil, err
		}
	} else {
		if err = root.WriteFile(fileName, []byte(val), 0o644); err != nil {
			return nil, err
		}
	}

	return eval.Null{}, nil
}

func (f *files) exists(args ...module.IValue) (module.IValue, error) {
	root, err := os.OpenRoot(f.root)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	_, err = root.Stat(str(args[0]))
	switch {
	case err == nil:
		return eval.Bool(true), nil
	case errors.Is(err, iofs.ErrNotExist):
		return eval.Bool(false), nil
	default:
		return nil, err
	}
}

func (f *files) Name() string {
	return Name
}

func (f *files) Func(name string) (module.IModuleFunc, bool) {
	switch name {
	case Read:
		funcDed, err := module.NewFunc(Read, []module.Param{
			{Name: "path", Type: types.String{}},
		}, nil, f.read)
		if err != nil {
			return nil, false
		}
		return funcDed, true
	case Write:
		funcDed, err := module.NewFunc(Write, []module.Param{
			{Name: "path", Type: types.String{}},
			{Name: "content", Type: types.String{}},
		}, nil, f.write)
		if err != nil {
			return nil, false
		}
		return funcDed, true
	case Exists:
		funcDed, err := module.NewFunc(Exists, []module.Param{
			{Name: "path", Type: types.String{}},
		}, nil, f.exists)
		if err != nil {
			return nil, false
		}
		return funcDed, true
	default:
		return nil, false
	}
}

func (f *files) Funcs() []module.IModuleFunc {
	var funcs []module.IModuleFunc
	if f, ok := f.Func(Read); ok {
		funcs = append(funcs, f)
	}
	if f, ok := f.Func(Write); ok {
		funcs = append(funcs, f)
	}
	if f, ok := f.Func(Exists); ok {
		funcs = append(funcs, f)
	}
	return funcs
}

func str(v module.IValue) string {
	s, _ := v.(eval.String)
	return string(s)
}

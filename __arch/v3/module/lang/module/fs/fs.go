package fs

import (
	"errors"
	"fmt"
	iofs "io/fs"
	"os"
	"sync"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
	"github.com/siper92/akha/lang/module/types"
)

const (
	Name      = "fs"
	Read      = "read"
	Write     = "write"
	Overwrite = "overwrite"
	Exists    = "exists"
	ReadJSON  = "readJSON"
	WriteJSON = "writeJSON"
	ReadYAML  = "readYAML"
	WriteYAML = "writeYAML"

	filePerm = 0o644
	dirPerm  = 0o755
	maxDepth = 512
)

var (
	ErrRoot   = errors.New("fs root unavailable")
	ErrPath   = errors.New("invalid path")
	ErrDecode = errors.New("decode failed")
	ErrEncode = errors.New("encode failed")
	ErrDepth  = errors.New("nesting too deep")
)

type Storage interface {
	ReadFile(name string) ([]byte, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Stat(name string) (iofs.FileInfo, error)
	MkdirAll(name string, perm os.FileMode) error
}

type FS interface {
	Read(args ...module.IValue) (module.IValue, error)
	Write(args ...module.IValue) (module.IValue, error)
	Overwrite(args ...module.IValue) (module.IValue, error)
	Exists(args ...module.IValue) (module.IValue, error)
	ReadJSON(args ...module.IValue) (module.IValue, error)
	WriteJSON(args ...module.IValue) (module.IValue, error)
	ReadYAML(args ...module.IValue) (module.IValue, error)
	WriteYAML(args ...module.IValue) (module.IValue, error)
}

var (
	_ Storage        = (*os.Root)(nil)
	_ FS             = (*files)(nil)
	_ module.IModule = (*files)(nil)
)

type files struct {
	*module.Module
	store func() (Storage, error)
	codec eval.Codec
}

type op func(s Storage, p string) (module.IValue, error)

func New(root string) (module.IModule, error) {
	if root == "" {
		return nil, fmt.Errorf("%w: empty root", ErrRoot)
	}

	return newFiles(sync.OnceValues(func() (Storage, error) {
		r, err := os.OpenRoot(root)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRoot, err)
		}
		return r, nil
	}))
}

func NewWithStorage(store Storage) (module.IModule, error) {
	if store == nil {
		return nil, fmt.Errorf("%w: nil storage", ErrRoot)
	}

	return newFiles(func() (Storage, error) { return store, nil })
}

func newFiles(store func() (Storage, error)) (module.IModule, error) {
	f := &files{store: store, codec: eval.NewCodec()}

	pathParam := module.Param{Name: "path", Type: types.String{}}
	content := module.Param{Name: "content", Type: types.String{}}
	data := module.Param{Name: "data", Type: types.Any{}}

	defs := []struct {
		name   string
		params []module.Param
		impl   module.Impl
	}{
		{Read, []module.Param{pathParam}, f.Read},
		{Write, []module.Param{pathParam, content}, f.Write},
		{Overwrite, []module.Param{pathParam, content}, f.Overwrite},
		{Exists, []module.Param{pathParam}, f.Exists},
		{ReadJSON, []module.Param{pathParam}, f.ReadJSON},
		{WriteJSON, []module.Param{pathParam, data}, f.WriteJSON},
		{ReadYAML, []module.Param{pathParam}, f.ReadYAML},
		{WriteYAML, []module.Param{pathParam, data}, f.WriteYAML},
	}

	funcs := make([]module.IModuleFunc, 0, len(defs))
	for _, d := range defs {
		fn, err := module.NewFunc(d.name, d.params, nil, d.impl)
		if err != nil {
			return nil, err
		}
		funcs = append(funcs, fn)
	}

	m, err := module.NewModule(Name, funcs...)
	if err != nil {
		return nil, err
	}
	f.Module = m

	return f, nil
}

func (f *files) do(name string, args []module.IValue, want int, fn op) (module.IValue, error) {
	if len(args) != want {
		return nil, fmt.Errorf("%s: %w, want %d, got %d", name, module.ErrArity, want, len(args))
	}

	p, err := clean(args[0])
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", name, args[0], err)
	}

	s, err := f.store()
	if err != nil {
		return nil, err
	}

	out, err := fn(s, p)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", name, args[0], err)
	}

	return out, nil
}

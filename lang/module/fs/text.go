package fs

import (
	"errors"
	iofs "io/fs"
	"os"
	"path/filepath"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
)

func (f *files) Read(args ...module.IValue) (module.IValue, error) {
	return f.do(Read, args, 1, func(s Storage, p string) (module.IValue, error) {
		b, err := s.ReadFile(p)
		if err != nil {
			return nil, err
		}
		return eval.String(b), nil
	})
}

func (f *files) Write(args ...module.IValue) (module.IValue, error) {
	return f.do(Write, args, 2, func(s Storage, p string) (module.IValue, error) {
		return eval.Null{}, put(s, p, []byte(text(args[1])), os.O_APPEND)
	})
}

func (f *files) Overwrite(args ...module.IValue) (module.IValue, error) {
	return f.do(Overwrite, args, 2, func(s Storage, p string) (module.IValue, error) {
		return eval.Null{}, put(s, p, []byte(text(args[1])), os.O_TRUNC)
	})
}

func (f *files) Exists(args ...module.IValue) (module.IValue, error) {
	return f.do(Exists, args, 1, func(s Storage, p string) (module.IValue, error) {
		_, err := s.Stat(p)
		switch {
		case err == nil:
			return eval.Bool(true), nil
		case errors.Is(err, iofs.ErrNotExist):
			return eval.Bool(false), nil
		default:
			return nil, err
		}
	})
}

func put(s Storage, p string, data []byte, mode int) error {
	if dir := filepath.Dir(p); dir != "." {
		if err := s.MkdirAll(dir, dirPerm); err != nil {
			return err
		}
	}

	file, err := s.OpenFile(p, os.O_WRONLY|os.O_CREATE|mode, filePerm)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	return errors.Join(err, file.Close())
}

func text(v module.IValue) string {
	s, _ := v.(eval.String)
	return string(s)
}

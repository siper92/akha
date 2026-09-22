package fs

import (
	"context"

	"github.com/siper92/akha/lang/eval"
)

const (
	FuncReadFile   = "ReadFile"
	FuncWriteFile  = "WriteFile"
	FuncUpdateFile = "UpdateFile"
	FuncListFiles  = "ListFiles"
)

func NewModule(f FS) eval.Module {
	return eval.NewModule(ModuleName,
		eval.NewBuiltin(eval.Spec{Name: FuncReadFile, MinArgs: 1, MaxArgs: 1}, readFn(f)),
		eval.NewBuiltin(eval.Spec{Name: FuncWriteFile, MinArgs: 2, MaxArgs: 2}, writeFn(f)),
		eval.NewBuiltin(eval.Spec{Name: FuncUpdateFile, MinArgs: 2, MaxArgs: 2}, updateFn(f)),
		eval.NewBuiltin(eval.Spec{Name: FuncListFiles, MinArgs: 1, MaxArgs: 1}, listFn(f)),
	)
}

func readFn(f FS) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		path, err := eval.ArgString(args, 0)
		if err != nil {
			return nil, err
		}
		content, err := f.ReadFile(ctx, path)
		if err != nil {
			return nil, err
		}
		return eval.Str(content), nil
	}
}

func writeFn(f FS) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		path, content, err := pathContent(args)
		if err != nil {
			return nil, err
		}
		return eval.None(), f.WriteFile(ctx, path, content)
	}
}

func updateFn(f FS) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		path, content, err := pathContent(args)
		if err != nil {
			return nil, err
		}
		return eval.None(), f.UpdateFile(ctx, path, content)
	}
}

func listFn(f FS) eval.CallFunc {
	return func(ctx context.Context, args []eval.Value, kwargs map[string]eval.Value) (eval.Value, error) {
		dir, err := eval.ArgString(args, 0)
		if err != nil {
			return nil, err
		}
		names, err := f.ListFiles(ctx, dir)
		if err != nil {
			return nil, err
		}
		vals := make([]eval.Value, len(names))
		for i, n := range names {
			vals[i] = eval.Str(n)
		}
		return eval.List(vals...), nil
	}
}

func pathContent(args []eval.Value) (string, string, error) {
	path, err := eval.ArgString(args, 0)
	if err != nil {
		return "", "", err
	}
	content, err := eval.ArgString(args, 1)
	if err != nil {
		return "", "", err
	}
	return path, content, nil
}

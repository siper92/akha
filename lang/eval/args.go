package eval

import (
	"errors"
	"fmt"
)

var (
	ErrArgType    = errors.New("argument type")
	ErrMissingArg = errors.New("missing argument")
)

func ArgString(args []Value, i int) (string, error) {
	if i < 0 || i >= len(args) {
		return "", fmt.Errorf("%w: argument %d", ErrMissingArg, i+1)
	}
	s, ok := AsString(args[i])
	if !ok {
		return "", fmt.Errorf("%w: argument %d must be string, got %s", ErrArgType, i+1, TypeOf(args[i]))
	}
	return s, nil
}

func ArgInt(args []Value, i int) (int64, error) {
	if i < 0 || i >= len(args) {
		return 0, fmt.Errorf("%w: argument %d", ErrMissingArg, i+1)
	}

	n, ok := AsInt(args[i])
	if !ok {
		return 0, fmt.Errorf("%w: argument %d must be int, got %s", ErrArgType, i+1, TypeOf(args[i]))
	}

	return n, nil
}

func KwargString(kw map[string]Value, name, def string) (string, error) {
	v, ok := kw[name]
	if !ok {
		return def, nil
	}
	s, ok := AsString(v)
	if !ok {
		return "", fmt.Errorf("%w: argument %s must be string, got %s", ErrArgType, name, TypeOf(v))
	}
	return s, nil
}

func KwargInt(kw map[string]Value, name string, def int64) (int64, error) {
	v, ok := kw[name]
	if !ok {
		return def, nil
	}
	i, ok := AsInt(v)
	if !ok {
		return 0, fmt.Errorf("%w: argument %s must be int, got %s", ErrArgType, name, TypeOf(v))
	}
	return i, nil
}

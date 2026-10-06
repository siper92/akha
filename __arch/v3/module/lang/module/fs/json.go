package fs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
)

func (f *files) ReadJSON(args ...module.IValue) (module.IValue, error) {
	return f.do(ReadJSON, args, 1, func(s Storage, p string) (module.IValue, error) {
		b, err := s.ReadFile(p)
		if err != nil {
			return nil, err
		}
		return f.decodeJSON(b)
	})
}

func (f *files) WriteJSON(args ...module.IValue) (module.IValue, error) {
	return f.do(WriteJSON, args, 2, func(s Storage, p string) (module.IValue, error) {
		b, err := encodeJSON(args[1])
		if err != nil {
			return nil, err
		}
		return eval.Null{}, put(s, p, b, os.O_TRUNC)
	})
}

func encodeJSON(v module.IValue) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(v.String()), "", "  "); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEncode, err)
	}
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

func (f *files) decodeJSON(b []byte) (eval.Value, error) {
	dec := json.NewDecoder(bytes.NewReader(b))
	v, err := f.jsonValue(dec, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecode, err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("%w: trailing data", ErrDecode)
	}

	return v, nil
}

func (f *files) jsonValue(dec *json.Decoder, depth int) (eval.Value, error) {
	if depth > maxDepth {
		return nil, ErrDepth
	}

	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}

	delim, ok := tok.(json.Delim)
	if !ok {
		return f.codec.FromGo(tok)
	}

	switch delim {
	case '[':
		arr := eval.NewArray()
		for dec.More() {
			e, err := f.jsonValue(dec, depth+1)
			if err != nil {
				return nil, err
			}
			arr.Elems = append(arr.Elems, e)
		}
		_, err = dec.Token()
		return arr, err
	case '{':
		obj := eval.NewObject()
		for dec.More() {
			k, err := dec.Token()
			if err != nil {
				return nil, err
			}
			v, err := f.jsonValue(dec, depth+1)
			if err != nil {
				return nil, err
			}
			_ = obj.SetMember(k.(string), v)
		}
		_, err = dec.Token()
		return obj, err
	default:
		return nil, fmt.Errorf("unexpected %s", delim)
	}
}

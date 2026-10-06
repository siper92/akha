package eval

import (
	"slices"
)

var _ Codec = (*codec)(nil)

type codec struct{}

func NewCodec() Codec {
	return &codec{}
}

func (c *codec) FromGo(v any) (Value, error) {
	switch v := v.(type) {
	case nil:
		return Null{}, nil
	case bool:
		return Bool(v), nil
	case float64:
		return checkNumber(v)
	case int:
		return Number(v), nil
	case string:
		return String(v), nil
	case []any:
		arr := &Array{Elems: make([]Value, 0, len(v))}
		for _, e := range v {
			ev, err := c.FromGo(e)
			if err != nil {
				return nil, err
			}
			arr.Elems = append(arr.Elems, ev)
		}
		return arr, nil
	case map[string]any:
		obj := NewObject()
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			fv, err := c.FromGo(v[k])
			if err != nil {
				return nil, err
			}
			_ = obj.SetMember(k, fv)
		}
		return obj, nil
	}
	return nil, newError(CodeInternal, "unsupported go value %T", v)
}

func (c *codec) ToGo(v Value) any {
	switch v := v.(type) {
	case Bool:
		return bool(v)
	case Number:
		return float64(v)
	case String:
		return string(v)
	case *Array:
		out := make([]any, len(v.Elems))
		for i, e := range v.Elems {
			out[i] = c.ToGo(e)
		}
		return out
	case *Object:
		out := make(map[string]any, len(v.Keys))
		for _, k := range v.Keys {
			out[k] = c.ToGo(v.Fields[k])
		}
		return out
	}
	return nil
}

package fs

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
)

func (f *files) ReadYAML(args ...module.IValue) (module.IValue, error) {
	return f.do(ReadYAML, args, 1, func(s Storage, p string) (module.IValue, error) {
		b, err := s.ReadFile(p)
		if err != nil {
			return nil, err
		}
		return f.decodeYAML(b)
	})
}

func (f *files) WriteYAML(args ...module.IValue) (module.IValue, error) {
	return f.do(WriteYAML, args, 2, func(s Storage, p string) (module.IValue, error) {
		b, err := f.encodeYAML(args[1])
		if err != nil {
			return nil, err
		}
		return eval.Null{}, put(s, p, b, os.O_TRUNC)
	})
}

func (f *files) encodeYAML(v module.IValue) ([]byte, error) {
	ev, ok := v.(eval.Value)
	if !ok {
		return nil, fmt.Errorf("%w: unsupported value %s", ErrEncode, v)
	}

	node, err := f.yamlNode(ev, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEncode, err)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := errors.Join(enc.Encode(node), enc.Close()); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrEncode, err)
	}

	return buf.Bytes(), nil
}

func (f *files) yamlNode(v eval.Value, depth int) (*yaml.Node, error) {
	if depth > maxDepth {
		return nil, ErrDepth
	}

	switch v := v.(type) {
	case *eval.Array:
		n := &yaml.Node{Kind: yaml.SequenceNode}
		for _, e := range v.Elems {
			c, err := f.yamlNode(e, depth+1)
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, c)
		}
		return n, nil
	case *eval.Object:
		n := &yaml.Node{Kind: yaml.MappingNode}
		for _, k := range v.Keys {
			key := &yaml.Node{}
			if err := key.Encode(k); err != nil {
				return nil, err
			}
			c, err := f.yamlNode(v.Fields[k], depth+1)
			if err != nil {
				return nil, err
			}
			n.Content = append(n.Content, key, c)
		}
		return n, nil
	default:
		n := &yaml.Node{}
		if err := n.Encode(f.codec.ToGo(v)); err != nil {
			return nil, err
		}
		return n, nil
	}
}

func (f *files) decodeYAML(b []byte) (eval.Value, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecode, err)
	}

	v, err := f.yamlValue(&doc, 0)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecode, err)
	}

	return v, nil
}

func (f *files) yamlValue(n *yaml.Node, depth int) (eval.Value, error) {
	if depth > maxDepth {
		return nil, ErrDepth
	}

	switch n.Kind {
	case 0:
		return eval.Null{}, nil
	case yaml.DocumentNode:
		if len(n.Content) == 0 {
			return eval.Null{}, nil
		}
		return f.yamlValue(n.Content[0], depth+1)
	case yaml.AliasNode:
		return f.yamlValue(n.Alias, depth+1)
	case yaml.SequenceNode:
		arr := eval.NewArray()
		for _, c := range n.Content {
			e, err := f.yamlValue(c, depth+1)
			if err != nil {
				return nil, err
			}
			arr.Elems = append(arr.Elems, e)
		}
		return arr, nil
	case yaml.MappingNode:
		obj := eval.NewObject()
		for i := 0; i+1 < len(n.Content); i += 2 {
			k := n.Content[i]
			if k.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("line %d: mapping keys must be scalars", k.Line)
			}
			v, err := f.yamlValue(n.Content[i+1], depth+1)
			if err != nil {
				return nil, err
			}
			_ = obj.SetMember(k.Value, v)
		}
		return obj, nil
	case yaml.ScalarNode:
		var v any
		if err := n.Decode(&v); err != nil {
			return nil, err
		}
		return f.yamlScalar(v)
	default:
		return nil, fmt.Errorf("line %d: unsupported node kind %d", n.Line, n.Kind)
	}
}

func (f *files) yamlScalar(v any) (eval.Value, error) {
	switch v := v.(type) {
	case int64:
		return f.codec.FromGo(float64(v))
	case uint64:
		return f.codec.FromGo(float64(v))
	case time.Time:
		return eval.String(v.Format(time.RFC3339Nano)), nil
	default:
		return f.codec.FromGo(v)
	}
}

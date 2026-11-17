package pages

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

func splitFrontMatter(body []byte) (Meta, []byte, error) {
	b := bytes.TrimSpace(body)
	first := bytes.IndexByte(b, '\n')
	if first < 0 || string(bytes.TrimSpace(b[:first])) != "---" {
		return Meta{}, body, nil
	}
	rest := b[first+1:]
	for i := 0; i <= len(rest); {
		j := bytes.IndexByte(rest[i:], '\n')
		line, next := rest[i:], len(rest)
		if j >= 0 {
			line, next = rest[i:i+j], i+j+1
		}
		if string(bytes.TrimSpace(line)) == "---" {
			var m Meta
			if err := yaml.Unmarshal(rest[:i], &m); err != nil {
				return Meta{}, nil, fmt.Errorf("%w: %v", ErrFrontMatter, err)
			}
			return m, rest[next:], nil
		}
		if j < 0 {
			break
		}
		i = next
	}
	return Meta{}, nil, fmt.Errorf("%w: missing closing ---", ErrFrontMatter)
}

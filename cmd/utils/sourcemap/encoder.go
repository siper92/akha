package sourcemap

import (
	"bytes"
	"context"
	"fmt"
)

var _ Encoder = (*goymlEncoder)(nil)

type goymlEncoder struct{}

func NewEncoder() Encoder {
	return &goymlEncoder{}
}

func (e *goymlEncoder) Encode(ctx context.Context, files []*File) ([]byte, error) {
	var b bytes.Buffer
	for i, f := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if i > 0 {
			b.WriteString("---\n")
		}
		fmt.Fprintf(&b, "module: %s\nfile: %s\n", f.Module, f.Path)
		b.WriteString("imports:\n")
		writeList(&b, f.Imports)
		b.WriteString("types:\n")
		if len(f.Types) > 0 {
			for _, t := range f.Types {
				b.WriteString("---\n" + t + "\n")
			}
			b.WriteString("---\n\n")
		}
		b.WriteString("exports:\n")
		writeList(&b, f.Exports)
		b.WriteString("private:\n")
		writeList(&b, f.Private)
		b.WriteString("\n")
	}
	return b.Bytes(), nil
}

func writeList(b *bytes.Buffer, items []string) {
	for _, it := range items {
		b.WriteString("  - " + it + "\n")
	}
}

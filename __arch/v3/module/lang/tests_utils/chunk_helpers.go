package tests_utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

const (
	ChunkSep  = "\n---\n"
	ChunkMark = "// ### "
)

type Chunk struct {
	Name string
	Src  string
	Line int
	Want string
}

func ReadChunks(t *testing.T, path string) []Chunk {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read chunks: %v", err)
	}

	var chunks []Chunk
	first := 1
	for i, src := range strings.Split(string(data), ChunkSep) {
		c := Chunk{Name: fmt.Sprintf("chunk_%d_line_%d", i, first), Src: src}
		for n, l := range strings.Split(src, "\n") {
			j := strings.Index(l, ChunkMark)
			if j < 0 {
				continue
			}
			if c.Line > 0 {
				t.Fatalf("%s: one expected error per chunk", c.Name)
			}
			want, err := strconv.Unquote(strings.TrimSpace(l[j+len(ChunkMark):]))
			if err != nil {
				t.Fatalf("%s: bad marker %q: %v", c.Name, l, err)
			}
			c.Line, c.Want = n+1, want
		}
		chunks = append(chunks, c)
		first += strings.Count(src, "\n") + 2
	}

	return chunks
}

func RunChunks(t *testing.T, path string) {
	t.Helper()
	for _, c := range ReadChunks(t, path) {
		t.Run(c.Name, func(t *testing.T) {
			_, err := Parse(ParseFile, c.Src)
			switch {
			case err == nil && c.Line == 0:
			case err == nil:
				t.Fatalf("expected error at line %d: %s, got none", c.Line, c.Want)
			case c.Line == 0:
				t.Fatalf("unexpected error: %v", err)
			default:
				de := AsDiag(t, err)
				if de.Pos.Line != c.Line || !strings.Contains(de.Msg, c.Want) {
					t.Fatalf("error\n got: %v\nwant: line %d: %s", err, c.Line, c.Want)
				}
			}
		})
	}
}

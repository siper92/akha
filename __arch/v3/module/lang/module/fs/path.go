package fs

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
)

func clean(v module.IValue) (string, error) {
	s, ok := v.(eval.String)
	if !ok {
		return "", fmt.Errorf("%w: not a string", ErrPath)
	}

	p := strings.ReplaceAll(string(s), `\`, "/")
	switch {
	case p == "":
		return "", fmt.Errorf("%w: empty", ErrPath)
	case strings.ContainsRune(p, 0):
		return "", fmt.Errorf("%w: null byte", ErrPath)
	case volume(p):
		return "", fmt.Errorf("%w: volume names are not allowed", ErrPath)
	default:
		p = path.Clean(strings.TrimLeft(p, "/"))
	}

	if p == ".." || strings.HasPrefix(p, "../") {
		return "", fmt.Errorf("%w: escapes root", ErrPath)
	}

	return filepath.FromSlash(p), nil
}

func volume(p string) bool {
	if len(p) < 2 || p[1] != ':' {
		return false
	}
	c := p[0] | 0x20
	return c >= 'a' && c <= 'z' && (len(p) == 2 || p[2] == '/')
}

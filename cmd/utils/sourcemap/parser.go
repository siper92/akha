package sourcemap

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var _ Parser = (*treeParser)(nil)

type treeParser struct{}

func NewParser() Parser {
	return &treeParser{}
}

func (p *treeParser) Parse(ctx context.Context, root string, opts Options) ([]*File, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: %s", ErrNotDir, root)
	}
	modRoot, modPath, err := findModule(abs)
	if err != nil {
		return nil, err
	}
	dirs, err := collectDirs(abs)
	if err != nil {
		return nil, err
	}
	var out []*File
	for _, dir := range dirs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		files, err := parseDir(dir, modRoot, modPath, opts)
		if err != nil {
			return nil, err
		}
		out = append(out, files...)
	}
	return out, nil
}

var modLine = regexp.MustCompile(`(?m)^module\s+(\S+)`)

func findModule(dir string) (string, string, error) {
	for cur := dir; ; {
		data, err := os.ReadFile(filepath.Join(cur, "go.mod"))
		if err == nil {
			m := modLine.FindSubmatch(data)
			if m == nil {
				return "", "", fmt.Errorf("%w: no module line in %s", ErrNoModule, cur)
			}
			return cur, string(m[1]), nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return "", "", fmt.Errorf("%w: %s", ErrNoModule, dir)
		}
		cur = parent
	}
}

func collectDirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root {
			name := d.Name()
			if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || name == "testdata" || name == "vendor" {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				return filepath.SkipDir
			}
		}
		dirs = append(dirs, path)
		return nil
	})
	return dirs, err
}

func parseDir(dir string, modRoot string, modPath string, opts Options) ([]*File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	var paths []string
	var asts []*ast.File
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		path := filepath.Join(dir, name)
		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		paths = append(paths, path)
		asts = append(asts, f)
	}
	r := &renderer{fset: fset, opts: opts}
	validations := map[string][]string{}
	for _, f := range asts {
		r.collectValidations(f, validations)
	}
	out := make([]*File, 0, len(asts))
	for i, f := range asts {
		rel, err := filepath.Rel(modRoot, paths[i])
		if err != nil {
			return nil, err
		}
		out = append(out, r.file(f, modPath, filepath.ToSlash(rel), validations))
	}
	return out, nil
}

type namedLine struct {
	name string
	line string
}

func (r *renderer) file(f *ast.File, modPath string, rel string, validations map[string][]string) *File {
	out := &File{Module: modPath, Path: rel}
	for _, imp := range f.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			path = imp.Path.Value
		}
		out.Imports = append(out.Imports, path)
	}
	sort.Strings(out.Imports)

	var types, exports, private []namedLine
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				s := spec.(*ast.TypeSpec)
				if !r.visible(s.Name.Name) {
					continue
				}
				unit := strings.Join(append(append([]string{}, validations[s.Name.Name]...), r.typeSpec(s)), "\n")
				types = append(types, namedLine{s.Name.Name, unit})
			}
		case *ast.FuncDecl:
			name := d.Name.Name
			if name == "_" {
				continue
			}
			line := namedLine{name, r.funcDecl(d)}
			if ast.IsExported(name) {
				exports = append(exports, line)
			} else if r.opts.WithPrivate {
				private = append(private, line)
			}
		}
	}
	out.Types = sortedLines(types)
	out.Exports = sortedLines(exports)
	out.Private = sortedLines(private)
	return out
}

func sortedLines(in []namedLine) []string {
	sort.SliceStable(in, func(i, j int) bool { return in[i].name < in[j].name })
	out := make([]string, len(in))
	for i, l := range in {
		out[i] = l.line
	}
	return out
}

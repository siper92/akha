package sourcemap

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

type renderer struct {
	fset *token.FileSet
	opts Options
}

func (r *renderer) visible(name string) bool {
	return r.opts.WithPrivate || ast.IsExported(name)
}

func (r *renderer) expr(e ast.Expr) string {
	var b bytes.Buffer
	if err := printer.Fprint(&b, r.fset, e); err != nil {
		return ""
	}
	return b.String()
}

func (r *renderer) collectValidations(f *ast.File, into map[string][]string) {
	for _, decl := range f.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok || d.Tok != token.VAR {
			continue
		}
		for _, spec := range d.Specs {
			s := spec.(*ast.ValueSpec)
			if len(s.Names) != 1 || s.Names[0].Name != "_" || s.Type == nil || len(s.Values) != 1 {
				continue
			}
			target := assertTarget(s.Values[0])
			if target == "" {
				continue
			}
			into[target] = append(into[target], "var _ "+r.expr(s.Type)+" = "+r.expr(s.Values[0]))
		}
	}
}

func assertTarget(e ast.Expr) string {
	switch v := ast.Unparen(e).(type) {
	case *ast.CallExpr:
		fn := ast.Unparen(v.Fun)
		if star, ok := fn.(*ast.StarExpr); ok {
			fn = star.X
		}
		return baseName(fn)
	case *ast.CompositeLit:
		return baseName(v.Type)
	case *ast.UnaryExpr:
		return assertTarget(v.X)
	}
	return ""
}

func baseName(e ast.Expr) string {
	switch v := ast.Unparen(e).(type) {
	case *ast.Ident:
		return v.Name
	case *ast.IndexExpr:
		return baseName(v.X)
	case *ast.IndexListExpr:
		return baseName(v.X)
	}
	return ""
}

func (r *renderer) typeSpec(s *ast.TypeSpec) string {
	head := "type " + s.Name.Name + r.typeParams(s.TypeParams)
	if s.Assign.IsValid() {
		head += " ="
	}
	switch t := s.Type.(type) {
	case *ast.StructType:
		return head + " struct " + r.block(r.structFields(t.Fields))
	case *ast.InterfaceType:
		return head + " interface " + r.block(r.interfaceMethods(t.Methods))
	}
	return head + " " + r.expr(s.Type)
}

func (r *renderer) block(lines []string) string {
	if len(lines) == 0 {
		return "{}"
	}
	return "{\n\t" + strings.Join(lines, "\n\t") + "\n}"
}

func (r *renderer) structFields(fl *ast.FieldList) []string {
	var lines []string
	for _, f := range fl.List {
		typ := r.expr(f.Type)
		if len(f.Names) == 0 {
			lines = append(lines, typ)
			continue
		}
		for _, n := range f.Names {
			if r.visible(n.Name) {
				lines = append(lines, n.Name+" "+typ)
			}
		}
	}
	return lines
}

func (r *renderer) interfaceMethods(fl *ast.FieldList) []string {
	var lines []string
	for _, f := range fl.List {
		ft, ok := f.Type.(*ast.FuncType)
		if !ok || len(f.Names) == 0 {
			lines = append(lines, r.expr(f.Type))
			continue
		}
		for _, n := range f.Names {
			if r.visible(n.Name) {
				lines = append(lines, n.Name+r.signature(ft))
			}
		}
	}
	return lines
}

func (r *renderer) funcDecl(d *ast.FuncDecl) string {
	recv := ""
	if d.Recv != nil && len(d.Recv.List) > 0 {
		recv = "(" + r.params(d.Recv) + ") "
	}
	return "func " + recv + d.Name.Name + r.typeParams(d.Type.TypeParams) + r.signature(d.Type)
}

func (r *renderer) signature(ft *ast.FuncType) string {
	return "(" + r.params(ft.Params) + ")" + r.results(ft.Results)
}

func (r *renderer) typeParams(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	return "[" + r.params(fl) + "]"
}

func (r *renderer) params(fl *ast.FieldList) string {
	if fl == nil {
		return ""
	}
	var parts []string
	for _, f := range fl.List {
		typ := r.expr(f.Type)
		if len(f.Names) == 0 {
			parts = append(parts, typ)
			continue
		}
		for _, n := range f.Names {
			parts = append(parts, n.Name+" "+typ)
		}
	}
	return strings.Join(parts, ", ")
}

func (r *renderer) results(fl *ast.FieldList) string {
	if fl == nil || len(fl.List) == 0 {
		return ""
	}
	if len(fl.List) == 1 && len(fl.List[0].Names) == 0 {
		return " " + r.expr(fl.List[0].Type)
	}
	return " (" + r.params(fl) + ")"
}

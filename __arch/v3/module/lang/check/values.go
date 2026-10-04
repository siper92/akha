package check

import (
	"context"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/module/std"
)

var (
	_ ValueChecker = (*MemberChecker)(nil)
	_ ValueChecker = (*DivisorChecker)(nil)
)

type MemberChecker struct{}

type DivisorChecker struct{}

func Values() []ValueChecker {
	return []ValueChecker{&MemberChecker{}, &DivisorChecker{}, &CallChecker{Modules: std.Default()}}
}

func Literal(x ast.Expr) ast.Expr {
	return Resolve(x, func(string) ast.Expr { return nil })
}

func Resolve(x ast.Expr, lookup func(name string) ast.Expr) ast.Expr {
	switch x := x.(type) {
	case *ast.ObjectLit, *ast.ArrayLit, *ast.NumberLit, *ast.StringLit, *ast.BoolLit, *ast.NullLit:
		return x
	case *ast.Ident:
		return lookup(x.Name)
	case *ast.MemberExpr:
		obj, ok := Resolve(x.X, lookup).(*ast.ObjectLit)
		if !ok {
			return nil
		}
		if e, ok := entry(obj, x.Name); ok {
			return Resolve(e.Value, lookup)
		}
	}
	return nil
}

func (*MemberChecker) CheckValue(_ context.Context, x ast.Expr, known Known) error {
	m, ok := x.(*ast.MemberExpr)
	if !ok {
		return nil
	}

	switch k := known(m.X).(type) {
	case nil:
		return nil
	case *ast.ObjectLit:
		if _, ok := entry(k, m.Name); !ok {
			return newError(m.Pos, lexer.CodeMemberMissing, "use [\"key\"] for optional keys", "member %q does not exist on the object", m.Name)
		}
		return nil
	}

	return newError(m.Pos, lexer.CodeMemberKind, "use . on objects only", "member access works only on objects")
}

func (*DivisorChecker) CheckValue(_ context.Context, x ast.Expr, known Known) error {
	b, ok := x.(*ast.BinaryExpr)
	if !ok || (b.Op != lexer.Slash && b.Op != lexer.Percent) {
		return nil
	}

	if n, ok := known(b.Right).(*ast.NumberLit); ok && n.Value == 0 {
		return newError(b.Pos, lexer.CodeDivZero, "", "%s by zero", opName(b.Op))
	}

	return nil
}

func entry(obj *ast.ObjectLit, key string) (ast.Entry, bool) {
	for _, e := range obj.Entries {
		if e.Key == key {
			return e, true
		}
	}

	return ast.Entry{}, false
}

func opName(op lexer.Kind) string {
	if op == lexer.Percent {
		return "modulo"
	}

	return "division"
}

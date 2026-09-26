package tests

import (
	"fmt"
	"testing"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/token"
)

type StringCase struct {
	Name string
	Node fmt.Stringer
	Want string
}

func RunString(t *testing.T, cases []StringCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.Node.String()
			if got != c.Want {
				t.Fatalf("string\n got: %s\nwant: %s", got, c.Want)
			}
			if printed := fmt.Sprint(c.Node); printed != got {
				t.Fatalf("fmt does not use String\n got: %s\nwant: %s", printed, got)
			}
		})
	}
}

func RunStringParses(t *testing.T, cases []StringCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			parse := MustParse
			if _, ok := c.Node.(ast.Expr); ok {
				parse = MustParseExpr
			}
			if got := parse(t, c.Node.String()); got != c.Want {
				t.Fatalf("reparsed\n got: %s\nwant: %s", got, c.Want)
			}
		})
	}
}

func Id(name string) *ast.Ident {
	return &ast.Ident{Name: name}
}

func Num(v float64) *ast.NumberLit {
	return &ast.NumberLit{Value: v}
}

func Str(v string) *ast.StringLit {
	return &ast.StringLit{Value: v}
}

func Bin(op token.Kind, left, right ast.Expr) *ast.BinaryExpr {
	return &ast.BinaryExpr{Op: op, Left: left, Right: right}
}

func Un(op token.Kind, x ast.Expr) *ast.UnaryExpr {
	return &ast.UnaryExpr{Op: op, X: x}
}

func Member(x ast.Expr, name string) *ast.MemberExpr {
	return &ast.MemberExpr{X: x, Name: name}
}

func Index(x, index ast.Expr) *ast.IndexExpr {
	return &ast.IndexExpr{X: x, Index: index}
}

func Call(fn ast.Expr, args ...ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{Fn: fn, Args: args}
}

func Obj(entries ...ast.Entry) *ast.ObjectLit {
	return &ast.ObjectLit{Entries: entries}
}

func Arr(elems ...ast.Expr) *ast.ArrayLit {
	return &ast.ArrayLit{Elems: elems}
}

func Blk(stmts ...ast.Stmt) *ast.Block {
	return &ast.Block{Stmts: stmts}
}

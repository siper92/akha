package ast

import (
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestToString(t *testing.T) {
	cases := []tests_utils.Case[Node, string]{
		{
			Name: "Let statement",
			Input: &LetStmt{
				Name:  "x",
				Value: &NumberLit{Value: 42},
			},
			Expected: `let x = 42`,
		},
		{
			Name: "If statement with else",
			Input: &IfStmt{
				Cond: &Ident{Name: "x"},
				Then: &Block{
					Stmts: []Stmt{
						&ExprStmt{X: &Ident{Name: "doSomething"}},
					},
				},
				Else: &Block{
					Stmts: []Stmt{
						&ExprStmt{X: &Ident{Name: "doSomethingElse"}},
					},
				},
			},
			Expected: "if x {\n    doSomething\n} else {\n    doSomethingElse\n}",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := c.Input.String()
			if got != c.Expected {
				t.Errorf("Expected: %s, got: %s", c.Expected, got)
			}
		})
	}
}

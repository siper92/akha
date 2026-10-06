package ast

import (
	"fmt"
	"testing"

	"github.com/siper92/akha/lang/lexer"
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

func TestStmtToString(t *testing.T) {
	cases := []tests_utils.Case[fmt.Stringer, string]{
		// --- bindings
		{
			Name:     "var_without_value",
			Input:    &VarStmt{Name: "x"},
			Expected: "var x",
		},
		{
			Name:     "var_with_value",
			Input:    &VarStmt{Name: "x", Value: &StringLit{Value: "a"}},
			Expected: `var x = "a"`,
		},
		{
			Name: "assign_index",
			Input: &AssignStmt{
				Target: &IndexExpr{X: &Ident{Name: "a"}, Index: &StringLit{Value: "k"}},
				Value:  &BoolLit{Value: false},
			},
			Expected: `a["k"] = false`,
		},
		// --- flow
		{
			Name:     "break",
			Input:    &BreakStmt{},
			Expected: "break",
		},
		{
			Name:     "continue",
			Input:    &ContinueStmt{},
			Expected: "continue",
		},
		{
			Name:     "return_without_value",
			Input:    &ReturnStmt{},
			Expected: "return",
		},
		{
			Name:     "return_null",
			Input:    &ReturnStmt{Value: &NullLit{}},
			Expected: "return null",
		},
		{
			Name:     "exit_with_value",
			Input:    &ReturnStmt{Exit: true, Value: &NumberLit{Value: 1}},
			Expected: "exit 1",
		},
		// --- if chains and nesting
		{
			Name: "else_if_else",
			Input: &IfStmt{
				Cond: &Ident{Name: "a"},
				Then: &Block{},
				Else: &IfStmt{
					Cond: &Ident{Name: "b"},
					Then: &Block{},
					Else: &Block{},
				},
			},
			Expected: "if a {\n} else if b {\n} else {\n}",
		},
		{
			Name: "nested_indent",
			Input: &IfStmt{
				Cond: &Ident{Name: "a"},
				Then: &Block{Stmts: []Stmt{
					&IfStmt{
						Cond: &Ident{Name: "b"},
						Then: &Block{Stmts: []Stmt{
							&ExprStmt{X: &CallExpr{Fn: &Ident{Name: "f"}}},
						}},
					},
				}},
			},
			Expected: "if a {\n    if b {\n        f()\n    }\n}",
		},
		{
			Name: "header_object_is_wrapped",
			Input: &IfStmt{
				Cond: &MemberExpr{X: &ObjectLit{}, Name: "a"},
				Then: &Block{},
			},
			Expected: "if ({}.a) {\n}",
		},
		// --- loops
		{
			Name: "for_in_key_value",
			Input: &ForInStmt{
				Key:   "k",
				Value: "v",
				Iter:  &Ident{Name: "o"},
				Body:  &Block{},
			},
			Expected: "for k, v in o {\n}",
		},
		{
			Name: "for_in_var_object_nil_body",
			Input: &ForInStmt{
				Mutable: true,
				Value:   "v",
				Iter:    &ObjectLit{},
			},
			Expected: "for var v in ({}) {\n}",
		},
		{
			Name: "for_range",
			Input: &ForRangeStmt{
				Name:  "i",
				Start: &NumberLit{Value: 0},
				End: &BinaryExpr{
					Op:    lexer.Plus,
					Left:  &Ident{Name: "n"},
					Right: &NumberLit{Value: 1},
				},
				Body: &Block{Stmts: []Stmt{&BreakStmt{}}},
			},
			Expected: "for i range [0..n + 1] {\n    break\n}",
		},
		// --- containers
		{
			Name:     "block",
			Input:    &Block{Stmts: []Stmt{&BreakStmt{}}},
			Expected: "{\n    break\n}",
		},
		{
			Name:     "empty_script",
			Input:    &Script{},
			Expected: "",
		},
		{
			Name: "script",
			Input: &Script{Stmts: []Stmt{
				&LetStmt{Name: "a", Value: &NumberLit{Value: 1}},
				&ExprStmt{X: &CallExpr{Fn: &Ident{Name: "f"}}},
			}},
			Expected: "let a = 1\nf()",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateToString(t, c)
		})
	}
}

func TestExprToString(t *testing.T) {
	a, b, c := &Ident{Name: "a"}, &Ident{Name: "b"}, &Ident{Name: "c"}
	cases := []tests_utils.Case[fmt.Stringer, string]{
		// --- unary
		{
			Name:     "double_minus",
			Input:    &UnaryExpr{Op: lexer.Minus, X: &UnaryExpr{Op: lexer.Minus, X: a}},
			Expected: "- -a",
		},
		{
			Name:     "minus_negative_number",
			Input:    &UnaryExpr{Op: lexer.Minus, X: &NumberLit{Value: -1}},
			Expected: "- -1",
		},
		{
			Name:     "minus_binary",
			Input:    &UnaryExpr{Op: lexer.Minus, X: &BinaryExpr{Op: lexer.Plus, Left: a, Right: b}},
			Expected: "-(a + b)",
		},
		{
			Name:     "minus_not",
			Input:    &UnaryExpr{Op: lexer.Minus, X: &UnaryExpr{Op: lexer.Not, X: a}},
			Expected: "-(not a)",
		},
		{
			Name:     "not_and",
			Input:    &UnaryExpr{Op: lexer.Not, X: &BinaryExpr{Op: lexer.And, Left: a, Right: b}},
			Expected: "not (a and b)",
		},
		{
			Name:     "not_comparison",
			Input:    &UnaryExpr{Op: lexer.Not, X: &BinaryExpr{Op: lexer.Eq, Left: a, Right: b}},
			Expected: "not a == b",
		},
		{
			Name:     "not_not",
			Input:    &UnaryExpr{Op: lexer.Not, X: &UnaryExpr{Op: lexer.Not, X: a}},
			Expected: "not not a",
		},
		// --- binary grouping
		{
			Name: "left_assoc_no_parens",
			Input: &BinaryExpr{
				Op:    lexer.Minus,
				Left:  &BinaryExpr{Op: lexer.Minus, Left: a, Right: b},
				Right: c,
			},
			Expected: "a - b - c",
		},
		{
			Name: "right_nested_parens",
			Input: &BinaryExpr{
				Op:    lexer.Minus,
				Left:  a,
				Right: &BinaryExpr{Op: lexer.Minus, Left: b, Right: c},
			},
			Expected: "a - (b - c)",
		},
		{
			Name: "lower_prec_left",
			Input: &BinaryExpr{
				Op:    lexer.Star,
				Left:  &BinaryExpr{Op: lexer.Plus, Left: a, Right: b},
				Right: c,
			},
			Expected: "(a + b) * c",
		},
		{
			Name: "or_inside_and",
			Input: &BinaryExpr{
				Op:    lexer.And,
				Left:  &BinaryExpr{Op: lexer.Or, Left: a, Right: b},
				Right: c,
			},
			Expected: "(a or b) and c",
		},
		{
			Name: "comparison_left_is_grouped",
			Input: &BinaryExpr{
				Op:    lexer.Eq,
				Left:  &BinaryExpr{Op: lexer.Lt, Left: a, Right: b},
				Right: c,
			},
			Expected: "(a < b) == c",
		},
		{
			Name: "in_left_is_grouped",
			Input: &BinaryExpr{
				Op:    lexer.In,
				Left:  &BinaryExpr{Op: lexer.In, Left: a, Right: b},
				Right: c,
			},
			Expected: "(a in b) in c",
		},
		{
			Name: "comparison_inside_in",
			Input: &BinaryExpr{
				Op:    lexer.In,
				Left:  &BinaryExpr{Op: lexer.Eq, Left: a, Right: b},
				Right: c,
			},
			Expected: "a == b in c",
		},
		{
			Name:     "not_in",
			Input:    &BinaryExpr{Op: lexer.NotIn, Left: a, Right: b},
			Expected: "a not in b",
		},
		// --- postfix
		{
			Name:     "member_of_unary",
			Input:    &MemberExpr{X: &UnaryExpr{Op: lexer.Minus, X: a}, Name: "b"},
			Expected: "(-a).b",
		},
		{
			Name:     "index_of_binary",
			Input:    &IndexExpr{X: &BinaryExpr{Op: lexer.Plus, Left: a, Right: b}, Index: &NumberLit{Value: 0}},
			Expected: "(a + b)[0]",
		},
		{
			Name:     "call_on_member",
			Input:    &CallExpr{Fn: &MemberExpr{X: a, Name: "b"}, Args: []Expr{&NumberLit{Value: 1}, c}},
			Expected: "a.b(1, c)",
		},
		{
			Name:     "call_without_args",
			Input:    &CallExpr{Fn: a},
			Expected: "a()",
		},
		// --- numbers
		{
			Name:     "whole_float",
			Input:    &NumberLit{Value: 3.0},
			Expected: "3",
		},
		{
			Name:     "fraction",
			Input:    &NumberLit{Value: 0.5},
			Expected: "0.5",
		},
		{
			Name:     "large_number_no_exponent",
			Input:    &NumberLit{Value: 1e21},
			Expected: "1000000000000000000000",
		},
		// --- strings
		{
			Name:     "string_escapes",
			Input:    &StringLit{Value: "a\"b\\c\n\t\r"},
			Expected: `"a\"b\\c\n\t\r"`,
		},
		{
			Name:     "string_interp_is_escaped",
			Input:    &StringLit{Value: "${x}"},
			Expected: `"\${x}"`,
		},
		{
			Name:     "string_plain_dollar",
			Input:    &StringLit{Value: "$x $"},
			Expected: `"$x $"`,
		},
		{
			Name: "template",
			Input: &TemplateLit{Parts: []TemplatePart{
				{Text: "hi "},
				{Expr: &MemberExpr{X: &Ident{Name: "user"}, Name: "name"}},
				{Text: "!"},
			}},
			Expected: `"hi ${user.name}!"`,
		},
		{
			Name: "template_text_is_escaped",
			Input: &TemplateLit{Parts: []TemplatePart{
				{Text: `"${`},
				{Expr: a},
			}},
			Expected: `"\"\${${a}"`,
		},
		// --- literals
		{
			Name:     "bools_and_null",
			Input:    &ArrayLit{Elems: []Expr{&BoolLit{Value: true}, &BoolLit{Value: false}, &NullLit{}}},
			Expected: "[true, false, null]",
		},
		{
			Name:     "empty_array",
			Input:    &ArrayLit{},
			Expected: "[]",
		},
		{
			Name:     "empty_object",
			Input:    &ObjectLit{},
			Expected: "{}",
		},
		{
			Name: "object_keys",
			Input: &ObjectLit{Entries: []Entry{
				{Key: "a", Value: &NumberLit{Value: 1}},
				{Key: "a b", Value: &NumberLit{Value: 1}},
				{Key: "if", Value: &NumberLit{Value: 1}},
				{Key: "Null", Value: &NumberLit{Value: 1}},
				{Key: "fn", Value: &NumberLit{Value: 1}},
				{Key: "", Value: &NumberLit{Value: 1}},
				{Key: "_", Value: &NumberLit{Value: 1}},
				{Key: "x1", Value: &NumberLit{Value: 1}},
				{Key: "1x", Value: &NumberLit{Value: 1}},
				{Key: "é", Value: &NumberLit{Value: 1}},
				{Key: `a"b`, Value: &NumberLit{Value: 1}},
			}},
			Expected: `{a: 1, "a b": 1, "if": 1, "Null": 1, "fn": 1, "": 1, _: 1, x1: 1, "1x": 1, "é": 1, "a\"b": 1}`,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateToString(t, c)
		})
	}
}

func validateToString(t *testing.T, c tests_utils.Case[fmt.Stringer, string]) {
	t.Helper()
	if got := c.Input.String(); got != c.Expected {
		t.Errorf("\n got: %q\nwant: %q", got, c.Expected)
	}
}

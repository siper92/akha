package ast_test

import (
	"fmt"
	"testing"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/tests"
	"github.com/siper92/akha/lang/token"
)

var (
	_ fmt.Stringer = (*ast.Script)(nil)
	_ fmt.Stringer = (*ast.Block)(nil)
	_ fmt.Stringer = ast.Entry{}
	_ fmt.Stringer = ast.TemplatePart{}

	_ fmt.Stringer = (*ast.LetStmt)(nil)
	_ fmt.Stringer = (*ast.VarStmt)(nil)
	_ fmt.Stringer = (*ast.AssignStmt)(nil)
	_ fmt.Stringer = (*ast.ExprStmt)(nil)
	_ fmt.Stringer = (*ast.IfStmt)(nil)
	_ fmt.Stringer = (*ast.ForInStmt)(nil)
	_ fmt.Stringer = (*ast.ForRangeStmt)(nil)
	_ fmt.Stringer = (*ast.BreakStmt)(nil)
	_ fmt.Stringer = (*ast.ContinueStmt)(nil)
	_ fmt.Stringer = (*ast.ReturnStmt)(nil)

	_ fmt.Stringer = (*ast.Ident)(nil)
	_ fmt.Stringer = (*ast.NumberLit)(nil)
	_ fmt.Stringer = (*ast.StringLit)(nil)
	_ fmt.Stringer = (*ast.TemplateLit)(nil)
	_ fmt.Stringer = (*ast.BoolLit)(nil)
	_ fmt.Stringer = (*ast.NullLit)(nil)
	_ fmt.Stringer = (*ast.ArrayLit)(nil)
	_ fmt.Stringer = (*ast.ObjectLit)(nil)
	_ fmt.Stringer = (*ast.UnaryExpr)(nil)
	_ fmt.Stringer = (*ast.BinaryExpr)(nil)
	_ fmt.Stringer = (*ast.MemberExpr)(nil)
	_ fmt.Stringer = (*ast.IndexExpr)(nil)
	_ fmt.Stringer = (*ast.CallExpr)(nil)
)

var (
	a = tests.Id("a")
	b = tests.Id("b")
	c = tests.Id("c")
)

func primaryCases() []tests.StringCase {
	return []tests.StringCase{
		{
			Name: "ident",
			Node: tests.Id("Name"),
			Want: "Name",
		},
		{
			Name: "number_integral",
			Node: &ast.NumberLit{Value: 3, Raw: "3.0"},
			Want: "3",
		},
		{
			Name: "number_zero",
			Node: tests.Num(0),
			Want: "0",
		},
		{
			Name: "number_float",
			Node: tests.Num(3.2),
			Want: "3.2",
		},
		{
			Name: "number_fraction",
			Node: tests.Num(0.5),
			Want: "0.5",
		},
		{
			Name: "string_plain",
			Node: tests.Str("hello/"),
			Want: `"hello/"`,
		},
		{
			Name: "string_empty",
			Node: tests.Str(""),
			Want: `""`,
		},
		{
			Name: "string_escapes",
			Node: tests.Str("say \"hi\"\n\ttab \\ back\r"),
			Want: `"say \"hi\"\n\ttab \\ back\r"`,
		},
		{
			Name: "string_dollar_brace_escaped",
			Node: tests.Str("costs ${price}"),
			Want: `"costs \${price}"`,
		},
		{
			Name: "string_dollar_alone_kept",
			Node: tests.Str("cost $5 and $"),
			Want: `"cost $5 and $"`,
		},
		{
			Name: "string_unicode",
			Node: tests.Str("héllo 日本語"),
			Want: `"héllo 日本語"`,
		},
		{
			Name: "template_ident",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{{Text: "hi "}, {Expr: tests.Id("name")}}},
			Want: `"hi ${name}"`,
		},
		{
			Name: "template_value_refs",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{
				{Text: "hello "},
				{Expr: tests.Member(tests.Id("user"), "name")},
				{Text: ", first is "},
				{Expr: tests.Index(tests.Id("arr"), tests.Num(0))},
			}},
			Want: `"hello ${user.name}, first is ${arr[0]}"`,
		},
		{
			Name: "template_adjacent",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{{Expr: tests.Id("k")}, {Text: "="}, {Expr: tests.Id("v")}}},
			Want: `"${k}=${v}"`,
		},
		{
			Name: "template_string_index",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{{Expr: tests.Index(a, tests.Str("k"))}}},
			Want: `"${a["k"]}"`,
		},
		{
			Name: "template_text_escapes",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{{Text: "\"q\"\n${x} "}, {Expr: a}}},
			Want: `"\"q\"\n\${x} ${a}"`,
		},
		{
			Name: "true",
			Node: &ast.BoolLit{Value: true},
			Want: "true",
		},
		{
			Name: "false",
			Node: &ast.BoolLit{},
			Want: "false",
		},
		{
			Name: "null",
			Node: &ast.NullLit{},
			Want: "null",
		},
	}
}

func arrayObjectCases() []tests.StringCase {
	return []tests.StringCase{
		{
			Name: "array_empty",
			Node: tests.Arr(),
			Want: "[]",
		},
		{
			Name: "array_mixed",
			Node: tests.Arr(tests.Num(1), tests.Str("a"), &ast.BoolLit{Value: true}, &ast.NullLit{}),
			Want: `[1, "a", true, null]`,
		},
		{
			Name: "array_nested",
			Node: tests.Arr(tests.Arr(tests.Num(1), tests.Num(2)), tests.Arr(tests.Num(3), tests.Num(4))),
			Want: "[[1, 2], [3, 4]]",
		},
		{
			Name: "array_expr_element",
			Node: tests.Arr(tests.Bin(token.Plus, a, tests.Num(1))),
			Want: "[a + 1]",
		},
		{
			Name: "object_empty",
			Node: tests.Obj(),
			Want: "{}",
		},
		{
			Name: "object_ident_keys",
			Node: tests.Obj(ast.Entry{Key: "name", Value: tests.Id("name")}, ast.Entry{Key: "total", Value: tests.Id("n")}),
			Want: "{name: name, total: n}",
		},
		{
			Name: "object_string_keys",
			Node: tests.Obj(
				ast.Entry{Key: "content-type", Value: tests.Str("json")},
				ast.Entry{Key: "1", Value: &ast.BoolLit{Value: true}},
				ast.Entry{Key: "", Value: tests.Num(0)},
			),
			Want: `{"content-type": "json", "1": true, "": 0}`,
		},
		{
			Name: "object_keyword_keys_quoted",
			Node: tests.Obj(
				ast.Entry{Key: "if", Value: tests.Num(1)},
				ast.Entry{Key: "True", Value: tests.Num(2)},
				ast.Entry{Key: "fn", Value: tests.Num(3)},
			),
			Want: `{"if": 1, "True": 2, "fn": 3}`,
		},
		{
			Name: "object_underscore_keys",
			Node: tests.Obj(ast.Entry{Key: "_", Value: tests.Num(1)}, ast.Entry{Key: "_a1", Value: tests.Num(2)}),
			Want: "{_: 1, _a1: 2}",
		},
		{
			Name: "object_nested",
			Node: tests.Obj(ast.Entry{Key: "a", Value: tests.Obj(ast.Entry{Key: "b", Value: tests.Arr(tests.Num(1))})}),
			Want: "{a: {b: [1]}}",
		},
		{
			Name: "entry",
			Node: ast.Entry{Key: "a", Value: tests.Bin(token.Plus, tests.Num(1), tests.Num(2))},
			Want: "a: 1 + 2",
		},
		{
			Name: "template_part_text",
			Node: ast.TemplatePart{Text: "a\"${"},
			Want: `a\"\${`,
		},
		{
			Name: "template_part_expr",
			Node: ast.TemplatePart{Expr: tests.Member(a, "b")},
			Want: "${a.b}",
		},
	}
}

func operatorCases() []tests.StringCase {
	return []tests.StringCase{
		{
			Name: "or_left_assoc",
			Node: tests.Bin(token.Or, tests.Bin(token.Or, a, b), c),
			Want: "a or b or c",
		},
		{
			Name: "or_right_grouped",
			Node: tests.Bin(token.Or, a, tests.Bin(token.Or, b, c)),
			Want: "a or (b or c)",
		},
		{
			Name: "and_under_or",
			Node: tests.Bin(token.Or, a, tests.Bin(token.And, b, c)),
			Want: "a or b and c",
		},
		{
			Name: "or_under_and_grouped",
			Node: tests.Bin(token.And, tests.Bin(token.Or, a, b), c),
			Want: "(a or b) and c",
		},
		{
			Name: "not_under_and",
			Node: tests.Bin(token.And, tests.Un(token.Not, a), b),
			Want: "not a and b",
		},
		{
			Name: "not_over_cmp",
			Node: tests.Un(token.Not, tests.Bin(token.Eq, a, b)),
			Want: "not a == b",
		},
		{
			Name: "not_over_in",
			Node: tests.Un(token.Not, tests.Bin(token.In, a, b)),
			Want: "not a in b",
		},
		{
			Name: "not_over_and_grouped",
			Node: tests.Un(token.Not, tests.Bin(token.And, a, b)),
			Want: "not (a and b)",
		},
		{
			Name: "not_not",
			Node: tests.Un(token.Not, tests.Un(token.Not, a)),
			Want: "not not a",
		},
		{
			Name: "not_left_of_cmp_grouped",
			Node: tests.Bin(token.Eq, tests.Un(token.Not, a), b),
			Want: "(not a) == b",
		},
		{
			Name: "not_right_of_cmp_grouped",
			Node: tests.Bin(token.Eq, a, tests.Un(token.Not, b)),
			Want: "a == (not b)",
		},
		{
			Name: "in",
			Node: tests.Bin(token.In, tests.Num(2), tests.Id("arr")),
			Want: "2 in arr",
		},
		{
			Name: "not_in",
			Node: tests.Bin(token.NotIn, tests.Str("x"), tests.Id("obj")),
			Want: `"x" not in obj`,
		},
		{
			Name: "in_left_chain_grouped",
			Node: tests.Bin(token.In, tests.Bin(token.In, a, b), c),
			Want: "(a in b) in c",
		},
		{
			Name: "in_right_chain_grouped",
			Node: tests.Bin(token.In, a, tests.Bin(token.NotIn, b, c)),
			Want: "a in (b not in c)",
		},
		{
			Name: "cmp_under_in",
			Node: tests.Bin(token.In, tests.Bin(token.Eq, a, b), c),
			Want: "a == b in c",
		},
		{
			Name: "cmp_all_ops",
			Node: tests.Bin(token.And,
				tests.Bin(token.And, tests.Bin(token.NotEq, a, b), tests.Bin(token.Lt, a, b)),
				tests.Bin(token.And, tests.Bin(token.LtEq, a, b), tests.Bin(token.Or, tests.Bin(token.Gt, a, b), tests.Bin(token.GtEq, a, b))),
			),
			Want: "a != b and a < b and (a <= b and (a > b or a >= b))",
		},
		{
			Name: "cmp_left_chain_grouped",
			Node: tests.Bin(token.Lt, tests.Bin(token.Lt, a, b), c),
			Want: "(a < b) < c",
		},
		{
			Name: "cmp_right_chain_grouped",
			Node: tests.Bin(token.Eq, a, tests.Bin(token.Eq, b, c)),
			Want: "a == (b == c)",
		},
		{
			Name: "add_under_cmp",
			Node: tests.Bin(token.Lt, tests.Bin(token.Plus, a, tests.Num(1)), tests.Bin(token.Star, b, tests.Num(2))),
			Want: "a + 1 < b * 2",
		},
		{
			Name: "sub_left_assoc",
			Node: tests.Bin(token.Minus, tests.Bin(token.Minus, a, b), c),
			Want: "a - b - c",
		},
		{
			Name: "sub_right_grouped",
			Node: tests.Bin(token.Minus, a, tests.Bin(token.Minus, b, c)),
			Want: "a - (b - c)",
		},
		{
			Name: "mul_under_add",
			Node: tests.Bin(token.Plus, tests.Num(1), tests.Bin(token.Star, tests.Num(2), tests.Num(3))),
			Want: "1 + 2 * 3",
		},
		{
			Name: "add_under_mul_grouped",
			Node: tests.Bin(token.Star, tests.Bin(token.Plus, tests.Num(1), tests.Num(2)), tests.Num(3)),
			Want: "(1 + 2) * 3",
		},
		{
			Name: "mul_div_mod_left",
			Node: tests.Bin(token.Percent, tests.Bin(token.Slash, tests.Bin(token.Star, a, b), c), tests.Id("d")),
			Want: "a * b / c % d",
		},
		{
			Name: "div_right_grouped",
			Node: tests.Bin(token.Slash, a, tests.Bin(token.Slash, b, c)),
			Want: "a / (b / c)",
		},
		{
			Name: "neg",
			Node: tests.Un(token.Minus, tests.Num(3.2)),
			Want: "-3.2",
		},
		{
			Name: "neg_neg_spaced",
			Node: tests.Un(token.Minus, tests.Un(token.Minus, a)),
			Want: "- -a",
		},
		{
			Name: "neg_under_mul",
			Node: tests.Bin(token.Star, tests.Un(token.Minus, a), b),
			Want: "-a * b",
		},
		{
			Name: "neg_right_of_sub",
			Node: tests.Bin(token.Minus, a, tests.Un(token.Minus, b)),
			Want: "a - -b",
		},
		{
			Name: "neg_over_add_grouped",
			Node: tests.Un(token.Minus, tests.Bin(token.Plus, a, b)),
			Want: "-(a + b)",
		},
		{
			Name: "neg_over_not_grouped",
			Node: tests.Un(token.Minus, tests.Un(token.Not, a)),
			Want: "-(not a)",
		},
		{
			Name: "neg_over_index",
			Node: tests.Un(token.Minus, tests.Index(a, tests.Num(0))),
			Want: "-a[0]",
		},
	}
}

func postfixCases() []tests.StringCase {
	return []tests.StringCase{
		{
			Name: "member",
			Node: tests.Member(tests.Id("user"), "name"),
			Want: "user.name",
		},
		{
			Name: "member_chain",
			Node: tests.Member(tests.Member(a, "b"), "c"),
			Want: "a.b.c",
		},
		{
			Name: "index_number",
			Node: tests.Index(tests.Id("arr"), tests.Num(0)),
			Want: "arr[0]",
		},
		{
			Name: "index_string",
			Node: tests.Index(tests.Id("user"), tests.Str("content-type")),
			Want: `user["content-type"]`,
		},
		{
			Name: "index_expr",
			Node: tests.Index(a, tests.Bin(token.Plus, tests.Id("i"), tests.Num(1))),
			Want: "a[i + 1]",
		},
		{
			Name: "index_chain",
			Node: tests.Index(tests.Index(tests.Index(tests.Id("arr3"), tests.Num(4)), tests.Num(1)), tests.Str("2")),
			Want: `arr3[4][1]["2"]`,
		},
		{
			Name: "call_no_args",
			Node: tests.Call(tests.Id("f")),
			Want: "f()",
		},
		{
			Name: "call_args",
			Node: tests.Call(tests.Id("f"), tests.Bin(token.Plus, a, tests.Num(1)), tests.Arr(tests.Num(1)), tests.Obj(ast.Entry{Key: "k", Value: tests.Num(1)})),
			Want: "f(a + 1, [1], {k: 1})",
		},
		{
			Name: "call_chain",
			Node: tests.Call(tests.Call(tests.Id("f"), tests.Num(1)), tests.Num(2)),
			Want: "f(1)(2)",
		},
		{
			Name: "mixed_postfix",
			Node: tests.Member(tests.Call(tests.Index(tests.Member(a, "b"), tests.Num(0)), tests.Num(1)), "c"),
			Want: "a.b[0](1).c",
		},
		{
			Name: "index_on_neg_grouped",
			Node: tests.Index(tests.Un(token.Minus, a), tests.Num(0)),
			Want: "(-a)[0]",
		},
		{
			Name: "member_on_binary_grouped",
			Node: tests.Member(tests.Bin(token.Plus, a, b), "c"),
			Want: "(a + b).c",
		},
		{
			Name: "call_on_not_grouped",
			Node: tests.Call(tests.Un(token.Not, a)),
			Want: "(not a)()",
		},
		{
			Name: "member_on_object",
			Node: tests.Member(tests.Obj(ast.Entry{Key: "a", Value: tests.Num(1)}), "a"),
			Want: "{a: 1}.a",
		},
	}
}

func stmtCases() []tests.StringCase {
	return []tests.StringCase{
		{
			Name: "let",
			Node: &ast.LetStmt{Name: "name", Value: tests.Str("ak")},
			Want: `let name = "ak"`,
		},
		{
			Name: "let_discard",
			Node: &ast.LetStmt{Name: "_", Value: tests.Bin(token.Star, tests.Id("count"), tests.Num(2))},
			Want: "let _ = count * 2",
		},
		{
			Name: "var_with_init",
			Node: &ast.VarStmt{Name: "n", Value: tests.Num(0)},
			Want: "var n = 0",
		},
		{
			Name: "var_without_init",
			Node: &ast.VarStmt{Name: "last"},
			Want: "var last",
		},
		{
			Name: "assign_ident",
			Node: &ast.AssignStmt{Target: tests.Id("n"), Value: tests.Bin(token.Plus, tests.Id("n"), tests.Num(1))},
			Want: "n = n + 1",
		},
		{
			Name: "assign_nested_target",
			Node: &ast.AssignStmt{
				Target: tests.Member(tests.Index(tests.Member(tests.Id("data"), "items"), tests.Num(0)), "name"),
				Value:  tests.Str("x"),
			},
			Want: `data.items[0].name = "x"`,
		},
		{
			Name: "call_stmt",
			Node: &ast.ExprStmt{X: tests.Call(tests.Member(a, "b"), tests.Num(1))},
			Want: "a.b(1)",
		},
		{
			Name: "return_bare",
			Node: &ast.ReturnStmt{},
			Want: "return",
		},
		{
			Name: "return_value",
			Node: &ast.ReturnStmt{Value: tests.Obj(ast.Entry{Key: "total", Value: tests.Id("total")})},
			Want: "return {total: total}",
		},
		{
			Name: "exit_bare",
			Node: &ast.ReturnStmt{Exit: true},
			Want: "exit",
		},
		{
			Name: "exit_value",
			Node: &ast.ReturnStmt{Exit: true, Value: tests.Bin(token.Plus, a, b)},
			Want: "exit a + b",
		},
		{
			Name: "if_empty",
			Node: &ast.IfStmt{Cond: a, Then: tests.Blk()},
			Want: "if a {\n}",
		},
		{
			Name: "if_else",
			Node: &ast.IfStmt{
				Cond: tests.Bin(token.And, tests.Bin(token.Lt, tests.Id("n"), tests.Num(3)), tests.Un(token.Not, &ast.BoolLit{})),
				Then: tests.Blk(&ast.AssignStmt{Target: tests.Id("n"), Value: tests.Bin(token.Plus, tests.Id("n"), tests.Num(1))}),
				Else: tests.Blk(&ast.ReturnStmt{}),
			},
			Want: "if n < 3 and not false {\n    n = n + 1\n} else {\n    return\n}",
		},
		{
			Name: "if_else_if_else",
			Node: &ast.IfStmt{
				Cond: a,
				Then: tests.Blk(&ast.AssignStmt{Target: tests.Id("x"), Value: tests.Num(1)}),
				Else: &ast.IfStmt{
					Cond: b,
					Then: tests.Blk(&ast.AssignStmt{Target: tests.Id("x"), Value: tests.Num(2)}),
					Else: tests.Blk(&ast.AssignStmt{Target: tests.Id("x"), Value: tests.Num(3)}),
				},
			},
			Want: "if a {\n    x = 1\n} else if b {\n    x = 2\n} else {\n    x = 3\n}",
		},
		{
			Name: "if_object_header_grouped",
			Node: &ast.IfStmt{Cond: tests.Obj(ast.Entry{Key: "a", Value: tests.Num(1)}), Then: tests.Blk()},
			Want: "if ({a: 1}) {\n}",
		},
		{
			Name: "if_object_first_operand_grouped",
			Node: &ast.IfStmt{Cond: tests.Bin(token.Eq, tests.Member(tests.Obj(ast.Entry{Key: "a", Value: tests.Num(1)}), "a"), tests.Num(1)), Then: tests.Blk()},
			Want: "if ({a: 1}.a == 1) {\n}",
		},
		{
			Name: "if_object_right_operand_kept",
			Node: &ast.IfStmt{Cond: tests.Bin(token.In, tests.Str("name"), tests.Obj(ast.Entry{Key: "name", Value: tests.Num(1)})), Then: tests.Blk()},
			Want: "if \"name\" in {name: 1} {\n}",
		},
		{
			Name: "for_single",
			Node: &ast.ForInStmt{Value: "x", Iter: a, Body: tests.Blk()},
			Want: "for x in a {\n}",
		},
		{
			Name: "for_pair",
			Node: &ast.ForInStmt{Key: "i", Value: "v", Iter: a, Body: tests.Blk()},
			Want: "for i, v in a {\n}",
		},
		{
			Name: "for_var_pair",
			Node: &ast.ForInStmt{Mutable: true, Key: "i", Value: "v", Iter: tests.Arr(tests.Num(1), tests.Num(2)), Body: tests.Blk(&ast.ContinueStmt{})},
			Want: "for var i, v in [1, 2] {\n    continue\n}",
		},
		{
			Name: "for_discard_index",
			Node: &ast.ForInStmt{Key: "_", Value: "v", Iter: a, Body: tests.Blk(&ast.BreakStmt{})},
			Want: "for _, v in a {\n    break\n}",
		},
		{
			Name: "for_object_header_grouped",
			Node: &ast.ForInStmt{Key: "k", Value: "v", Iter: tests.Obj(ast.Entry{Key: "a", Value: tests.Num(1)}), Body: tests.Blk()},
			Want: "for k, v in ({a: 1}) {\n}",
		},
		{
			Name: "for_range",
			Node: &ast.ForRangeStmt{Name: "i", Start: tests.Num(0), End: tests.Num(10), Body: tests.Blk()},
			Want: "for i range [0..10] {\n}",
		},
		{
			Name: "for_var_range_expr_bounds",
			Node: &ast.ForRangeStmt{
				Mutable: true,
				Name:    "i",
				Start:   tests.Bin(token.Plus, a, tests.Num(1)),
				End:     tests.Bin(token.Star, b, tests.Num(2)),
				Body:    tests.Blk(&ast.BreakStmt{}),
			},
			Want: "for var i range [a + 1..b * 2] {\n    break\n}",
		},
		{
			Name: "nested_blocks_indent",
			Node: &ast.ForInStmt{Value: "x", Iter: a, Body: tests.Blk(
				&ast.IfStmt{Cond: tests.Id("x"), Then: tests.Blk(&ast.BreakStmt{}), Else: tests.Blk(
					&ast.ForRangeStmt{Name: "i", Start: tests.Num(0), End: tests.Num(3), Body: tests.Blk(&ast.ContinueStmt{})},
				)},
				&ast.LetStmt{Name: "y", Value: tests.Id("x")},
			)},
			Want: "for x in a {\n    if x {\n        break\n    } else {\n        for i range [0..3] {\n            continue\n        }\n    }\n    let y = x\n}",
		},
	}
}

func scriptCases() []tests.StringCase {
	return []tests.StringCase{
		{
			Name: "block_empty",
			Node: tests.Blk(),
			Want: "{\n}",
		},
		{
			Name: "block_stmts",
			Node: tests.Blk(&ast.LetStmt{Name: "a", Value: tests.Num(1)}, &ast.VarStmt{Name: "b"}),
			Want: "{\n    let a = 1\n    var b\n}",
		},
		{
			Name: "script_empty",
			Node: &ast.Script{},
			Want: "",
		},
		{
			Name: "script_lines",
			Node: &ast.Script{Stmts: []ast.Stmt{
				&ast.LetStmt{Name: "a", Value: tests.Num(1)},
				&ast.IfStmt{Cond: tests.Id("a"), Then: tests.Blk(&ast.LetStmt{Name: "b", Value: tests.Str("x")})},
				&ast.ReturnStmt{Value: tests.Id("a")},
			}},
			Want: "let a = 1\nif a {\n    let b = \"x\"\n}\nreturn a",
		},
	}
}

func TestString(t *testing.T) {
	// --- primary
	tests.RunString(t, primaryCases())

	// --- array, object, entry, template part
	tests.RunString(t, arrayObjectCases())

	// --- operators, minimal grouping by precedence
	tests.RunString(t, operatorCases())

	// --- postfix_expr
	tests.RunString(t, postfixCases())

	// --- stmt
	tests.RunString(t, stmtCases())

	// --- block and script
	tests.RunString(t, scriptCases())
}

func TestStringIsEbnf(t *testing.T) {
	// --- expressions reparse to the same string
	var exprs []tests.StringCase
	for _, group := range [][]tests.StringCase{primaryCases(), arrayObjectCases(), operatorCases(), postfixCases()} {
		for _, c := range group {
			if _, ok := c.Node.(ast.Expr); ok {
				exprs = append(exprs, c)
			}
		}
	}
	tests.RunStringParses(t, exprs)

	// --- statements reparse to the same string
	var stmts []tests.StringCase
	for _, c := range append(stmtCases(), scriptCases()...) {
		switch c.Node.(type) {
		case *ast.Script, *ast.LetStmt, *ast.VarStmt, *ast.AssignStmt, *ast.ExprStmt, *ast.IfStmt, *ast.ForInStmt, *ast.ForRangeStmt, *ast.ReturnStmt:
			stmts = append(stmts, c)
		}
	}
	tests.RunStringParses(t, stmts)
}

func TestStringScripts(t *testing.T) {
	// --- script parses to its canonical ebnf form
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "sample",
			Src:  tests.Sample,
			Want: tests.SampleCanonical,
		},
		{
			Name: "sample_crlf",
			Src:  tests.SampleCRLF,
			Want: tests.SampleCRLFCanonical,
		},
	})

	// --- canonical form is a fixed point
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "spec_def",
			Src:  tests.SpecDef,
		},
	})
}

func TestPos(t *testing.T) {
	// --- statement lines
	tests.RunPos(t, []tests.PosCase{
		{
			Name:  "sample",
			Src:   tests.Sample,
			Lines: []int{2, 3, 5, 13, 17},
		},
		{
			Name:  "sample_crlf",
			Src:   tests.SampleCRLF,
			Lines: []int{1, 3},
		},
		{
			Name:  "blank_lines_and_comments",
			Src:   "\n// c\n\nlet a = 1\n\n\nvar b\n",
			Lines: []int{4, 7},
		},
		{
			Name:  "multi_line_literal",
			Src:   "let a = [\n    1,\n    2,\n]\nlet b = 1\n",
			Lines: []int{1, 5},
		},
	})
}

func TestPosLineOnly(t *testing.T) {
	// --- every node carries a line only Pos
	tests.RunPosLineOnly(t, []tests.PosCase{
		{
			Name: "sample",
			Src:  tests.Sample,
		},
		{
			Name: "spec_def",
			Src:  tests.SpecDef,
		},
	})
}

package _arch_test

import (
	"fmt"
	"testing"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/tests_utils"
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
	a = tests_utils.Id("a")
	b = tests_utils.Id("b")
	c = tests_utils.Id("c")
)

func primaryCases() []tests_utils.StringCase {
	return []tests_utils.StringCase{
		{
			Name: "ident",
			Node: tests_utils.Id("Name"),
			Want: "Name",
		},
		{
			Name: "number_integral",
			Node: &ast.NumberLit{Value: 3, Raw: "3.0"},
			Want: "3",
		},
		{
			Name: "number_zero",
			Node: tests_utils.Num(0),
			Want: "0",
		},
		{
			Name: "number_float",
			Node: tests_utils.Num(3.2),
			Want: "3.2",
		},
		{
			Name: "number_fraction",
			Node: tests_utils.Num(0.5),
			Want: "0.5",
		},
		{
			Name: "string_plain",
			Node: tests_utils.Str("hello/"),
			Want: `"hello/"`,
		},
		{
			Name: "string_empty",
			Node: tests_utils.Str(""),
			Want: `""`,
		},
		{
			Name: "string_escapes",
			Node: tests_utils.Str("say \"hi\"\n\ttab \\ back\r"),
			Want: `"say \"hi\"\n\ttab \\ back\r"`,
		},
		{
			Name: "string_dollar_brace_escaped",
			Node: tests_utils.Str("costs ${price}"),
			Want: `"costs \${price}"`,
		},
		{
			Name: "string_dollar_alone_kept",
			Node: tests_utils.Str("cost $5 and $"),
			Want: `"cost $5 and $"`,
		},
		{
			Name: "string_unicode",
			Node: tests_utils.Str("héllo 日本語"),
			Want: `"héllo 日本語"`,
		},
		{
			Name: "template_ident",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{{Text: "hi "}, {Expr: tests_utils.Id("name")}}},
			Want: `"hi ${name}"`,
		},
		{
			Name: "template_value_refs",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{
				{Text: "hello "},
				{Expr: tests_utils.Member(tests_utils.Id("user"), "name")},
				{Text: ", first is "},
				{Expr: tests_utils.Index(tests_utils.Id("arr"), tests_utils.Num(0))},
			}},
			Want: `"hello ${user.name}, first is ${arr[0]}"`,
		},
		{
			Name: "template_adjacent",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{{Expr: tests_utils.Id("k")}, {Text: "="}, {Expr: tests_utils.Id("v")}}},
			Want: `"${k}=${v}"`,
		},
		{
			Name: "template_string_index",
			Node: &ast.TemplateLit{Parts: []ast.TemplatePart{{Expr: tests_utils.Index(a, tests_utils.Str("k"))}}},
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

func arrayObjectCases() []tests_utils.StringCase {
	return []tests_utils.StringCase{
		{
			Name: "array_empty",
			Node: tests_utils.Arr(),
			Want: "[]",
		},
		{
			Name: "array_mixed",
			Node: tests_utils.Arr(tests_utils.Num(1), tests_utils.Str("a"), &ast.BoolLit{Value: true}, &ast.NullLit{}),
			Want: `[1, "a", true, null]`,
		},
		{
			Name: "array_nested",
			Node: tests_utils.Arr(tests_utils.Arr(tests_utils.Num(1), tests_utils.Num(2)), tests_utils.Arr(tests_utils.Num(3), tests_utils.Num(4))),
			Want: "[[1, 2], [3, 4]]",
		},
		{
			Name: "array_expr_element",
			Node: tests_utils.Arr(tests_utils.Bin(lexer.Plus, a, tests_utils.Num(1))),
			Want: "[a + 1]",
		},
		{
			Name: "object_empty",
			Node: tests_utils.Obj(),
			Want: "{}",
		},
		{
			Name: "object_ident_keys",
			Node: tests_utils.Obj(ast.Entry{Key: "name", Value: tests_utils.Id("name")}, ast.Entry{Key: "total", Value: tests_utils.Id("n")}),
			Want: "{name: name, total: n}",
		},
		{
			Name: "object_string_keys",
			Node: tests_utils.Obj(
				ast.Entry{Key: "content-type", Value: tests_utils.Str("json")},
				ast.Entry{Key: "1", Value: &ast.BoolLit{Value: true}},
				ast.Entry{Key: "", Value: tests_utils.Num(0)},
			),
			Want: `{"content-type": "json", "1": true, "": 0}`,
		},
		{
			Name: "object_keyword_keys_quoted",
			Node: tests_utils.Obj(
				ast.Entry{Key: "if", Value: tests_utils.Num(1)},
				ast.Entry{Key: "True", Value: tests_utils.Num(2)},
				ast.Entry{Key: "fn", Value: tests_utils.Num(3)},
			),
			Want: `{"if": 1, "True": 2, "fn": 3}`,
		},
		{
			Name: "object_underscore_keys",
			Node: tests_utils.Obj(ast.Entry{Key: "_", Value: tests_utils.Num(1)}, ast.Entry{Key: "_a1", Value: tests_utils.Num(2)}),
			Want: "{_: 1, _a1: 2}",
		},
		{
			Name: "object_nested",
			Node: tests_utils.Obj(ast.Entry{Key: "a", Value: tests_utils.Obj(ast.Entry{Key: "b", Value: tests_utils.Arr(tests_utils.Num(1))})}),
			Want: "{a: {b: [1]}}",
		},
		{
			Name: "entry",
			Node: ast.Entry{Key: "a", Value: tests_utils.Bin(lexer.Plus, tests_utils.Num(1), tests_utils.Num(2))},
			Want: "a: 1 + 2",
		},
		{
			Name: "template_part_text",
			Node: ast.TemplatePart{Text: "a\"${"},
			Want: `a\"\${`,
		},
		{
			Name: "template_part_expr",
			Node: ast.TemplatePart{Expr: tests_utils.Member(a, "b")},
			Want: "${a.b}",
		},
	}
}

func operatorCases() []tests_utils.StringCase {
	return []tests_utils.StringCase{
		{
			Name: "or_left_assoc",
			Node: tests_utils.Bin(lexer.Or, tests_utils.Bin(lexer.Or, a, b), c),
			Want: "a or b or c",
		},
		{
			Name: "or_right_grouped",
			Node: tests_utils.Bin(lexer.Or, a, tests_utils.Bin(lexer.Or, b, c)),
			Want: "a or (b or c)",
		},
		{
			Name: "and_under_or",
			Node: tests_utils.Bin(lexer.Or, a, tests_utils.Bin(lexer.And, b, c)),
			Want: "a or b and c",
		},
		{
			Name: "or_under_and_grouped",
			Node: tests_utils.Bin(lexer.And, tests_utils.Bin(lexer.Or, a, b), c),
			Want: "(a or b) and c",
		},
		{
			Name: "not_under_and",
			Node: tests_utils.Bin(lexer.And, tests_utils.Un(lexer.Not, a), b),
			Want: "not a and b",
		},
		{
			Name: "not_over_cmp",
			Node: tests_utils.Un(lexer.Not, tests_utils.Bin(lexer.Eq, a, b)),
			Want: "not a == b",
		},
		{
			Name: "not_over_in",
			Node: tests_utils.Un(lexer.Not, tests_utils.Bin(lexer.In, a, b)),
			Want: "not a in b",
		},
		{
			Name: "not_over_and_grouped",
			Node: tests_utils.Un(lexer.Not, tests_utils.Bin(lexer.And, a, b)),
			Want: "not (a and b)",
		},
		{
			Name: "not_not",
			Node: tests_utils.Un(lexer.Not, tests_utils.Un(lexer.Not, a)),
			Want: "not not a",
		},
		{
			Name: "not_left_of_cmp_grouped",
			Node: tests_utils.Bin(lexer.Eq, tests_utils.Un(lexer.Not, a), b),
			Want: "(not a) == b",
		},
		{
			Name: "not_right_of_cmp_grouped",
			Node: tests_utils.Bin(lexer.Eq, a, tests_utils.Un(lexer.Not, b)),
			Want: "a == (not b)",
		},
		{
			Name: "in",
			Node: tests_utils.Bin(lexer.In, tests_utils.Num(2), tests_utils.Id("arr")),
			Want: "2 in arr",
		},
		{
			Name: "not_in",
			Node: tests_utils.Bin(lexer.NotIn, tests_utils.Str("x"), tests_utils.Id("obj")),
			Want: `"x" not in obj`,
		},
		{
			Name: "in_left_chain_grouped",
			Node: tests_utils.Bin(lexer.In, tests_utils.Bin(lexer.In, a, b), c),
			Want: "(a in b) in c",
		},
		{
			Name: "in_right_chain_grouped",
			Node: tests_utils.Bin(lexer.In, a, tests_utils.Bin(lexer.NotIn, b, c)),
			Want: "a in (b not in c)",
		},
		{
			Name: "cmp_under_in",
			Node: tests_utils.Bin(lexer.In, tests_utils.Bin(lexer.Eq, a, b), c),
			Want: "a == b in c",
		},
		{
			Name: "cmp_all_ops",
			Node: tests_utils.Bin(lexer.And,
				tests_utils.Bin(lexer.And, tests_utils.Bin(lexer.NotEq, a, b), tests_utils.Bin(lexer.Lt, a, b)),
				tests_utils.Bin(lexer.And, tests_utils.Bin(lexer.LtEq, a, b), tests_utils.Bin(lexer.Or, tests_utils.Bin(lexer.Gt, a, b), tests_utils.Bin(lexer.GtEq, a, b))),
			),
			Want: "a != b and a < b and (a <= b and (a > b or a >= b))",
		},
		{
			Name: "cmp_left_chain_grouped",
			Node: tests_utils.Bin(lexer.Lt, tests_utils.Bin(lexer.Lt, a, b), c),
			Want: "(a < b) < c",
		},
		{
			Name: "cmp_right_chain_grouped",
			Node: tests_utils.Bin(lexer.Eq, a, tests_utils.Bin(lexer.Eq, b, c)),
			Want: "a == (b == c)",
		},
		{
			Name: "add_under_cmp",
			Node: tests_utils.Bin(lexer.Lt, tests_utils.Bin(lexer.Plus, a, tests_utils.Num(1)), tests_utils.Bin(lexer.Star, b, tests_utils.Num(2))),
			Want: "a + 1 < b * 2",
		},
		{
			Name: "sub_left_assoc",
			Node: tests_utils.Bin(lexer.Minus, tests_utils.Bin(lexer.Minus, a, b), c),
			Want: "a - b - c",
		},
		{
			Name: "sub_right_grouped",
			Node: tests_utils.Bin(lexer.Minus, a, tests_utils.Bin(lexer.Minus, b, c)),
			Want: "a - (b - c)",
		},
		{
			Name: "mul_under_add",
			Node: tests_utils.Bin(lexer.Plus, tests_utils.Num(1), tests_utils.Bin(lexer.Star, tests_utils.Num(2), tests_utils.Num(3))),
			Want: "1 + 2 * 3",
		},
		{
			Name: "add_under_mul_grouped",
			Node: tests_utils.Bin(lexer.Star, tests_utils.Bin(lexer.Plus, tests_utils.Num(1), tests_utils.Num(2)), tests_utils.Num(3)),
			Want: "(1 + 2) * 3",
		},
		{
			Name: "mul_div_mod_left",
			Node: tests_utils.Bin(lexer.Percent, tests_utils.Bin(lexer.Slash, tests_utils.Bin(lexer.Star, a, b), c), tests_utils.Id("d")),
			Want: "a * b / c % d",
		},
		{
			Name: "div_right_grouped",
			Node: tests_utils.Bin(lexer.Slash, a, tests_utils.Bin(lexer.Slash, b, c)),
			Want: "a / (b / c)",
		},
		{
			Name: "neg",
			Node: tests_utils.Un(lexer.Minus, tests_utils.Num(3.2)),
			Want: "-3.2",
		},
		{
			Name: "neg_neg_spaced",
			Node: tests_utils.Un(lexer.Minus, tests_utils.Un(lexer.Minus, a)),
			Want: "- -a",
		},
		{
			Name: "neg_under_mul",
			Node: tests_utils.Bin(lexer.Star, tests_utils.Un(lexer.Minus, a), b),
			Want: "-a * b",
		},
		{
			Name: "neg_right_of_sub",
			Node: tests_utils.Bin(lexer.Minus, a, tests_utils.Un(lexer.Minus, b)),
			Want: "a - -b",
		},
		{
			Name: "neg_over_add_grouped",
			Node: tests_utils.Un(lexer.Minus, tests_utils.Bin(lexer.Plus, a, b)),
			Want: "-(a + b)",
		},
		{
			Name: "neg_over_not_grouped",
			Node: tests_utils.Un(lexer.Minus, tests_utils.Un(lexer.Not, a)),
			Want: "-(not a)",
		},
		{
			Name: "neg_over_index",
			Node: tests_utils.Un(lexer.Minus, tests_utils.Index(a, tests_utils.Num(0))),
			Want: "-a[0]",
		},
	}
}

func postfixCases() []tests_utils.StringCase {
	return []tests_utils.StringCase{
		{
			Name: "member",
			Node: tests_utils.Member(tests_utils.Id("user"), "name"),
			Want: "user.name",
		},
		{
			Name: "member_chain",
			Node: tests_utils.Member(tests_utils.Member(a, "b"), "c"),
			Want: "a.b.c",
		},
		{
			Name: "index_number",
			Node: tests_utils.Index(tests_utils.Id("arr"), tests_utils.Num(0)),
			Want: "arr[0]",
		},
		{
			Name: "index_string",
			Node: tests_utils.Index(tests_utils.Id("user"), tests_utils.Str("content-type")),
			Want: `user["content-type"]`,
		},
		{
			Name: "index_expr",
			Node: tests_utils.Index(a, tests_utils.Bin(lexer.Plus, tests_utils.Id("i"), tests_utils.Num(1))),
			Want: "a[i + 1]",
		},
		{
			Name: "index_chain",
			Node: tests_utils.Index(tests_utils.Index(tests_utils.Index(tests_utils.Id("arr3"), tests_utils.Num(4)), tests_utils.Num(1)), tests_utils.Str("2")),
			Want: `arr3[4][1]["2"]`,
		},
		{
			Name: "call_no_args",
			Node: tests_utils.Call(tests_utils.Id("f")),
			Want: "f()",
		},
		{
			Name: "call_args",
			Node: tests_utils.Call(tests_utils.Id("f"), tests_utils.Bin(lexer.Plus, a, tests_utils.Num(1)), tests_utils.Arr(tests_utils.Num(1)), tests_utils.Obj(ast.Entry{Key: "k", Value: tests_utils.Num(1)})),
			Want: "f(a + 1, [1], {k: 1})",
		},
		{
			Name: "call_chain",
			Node: tests_utils.Call(tests_utils.Call(tests_utils.Id("f"), tests_utils.Num(1)), tests_utils.Num(2)),
			Want: "f(1)(2)",
		},
		{
			Name: "mixed_postfix",
			Node: tests_utils.Member(tests_utils.Call(tests_utils.Index(tests_utils.Member(a, "b"), tests_utils.Num(0)), tests_utils.Num(1)), "c"),
			Want: "a.b[0](1).c",
		},
		{
			Name: "index_on_neg_grouped",
			Node: tests_utils.Index(tests_utils.Un(lexer.Minus, a), tests_utils.Num(0)),
			Want: "(-a)[0]",
		},
		{
			Name: "member_on_binary_grouped",
			Node: tests_utils.Member(tests_utils.Bin(lexer.Plus, a, b), "c"),
			Want: "(a + b).c",
		},
		{
			Name: "call_on_not_grouped",
			Node: tests_utils.Call(tests_utils.Un(lexer.Not, a)),
			Want: "(not a)()",
		},
		{
			Name: "member_on_object",
			Node: tests_utils.Member(tests_utils.Obj(ast.Entry{Key: "a", Value: tests_utils.Num(1)}), "a"),
			Want: "{a: 1}.a",
		},
	}
}

func stmtCases() []tests_utils.StringCase {
	return []tests_utils.StringCase{
		{
			Name: "let",
			Node: &ast.LetStmt{Name: "name", Value: tests_utils.Str("ak")},
			Want: `let name = "ak"`,
		},
		{
			Name: "let_discard",
			Node: &ast.LetStmt{Name: "_", Value: tests_utils.Bin(lexer.Star, tests_utils.Id("count"), tests_utils.Num(2))},
			Want: "let _ = count * 2",
		},
		{
			Name: "var_with_init",
			Node: &ast.VarStmt{Name: "n", Value: tests_utils.Num(0)},
			Want: "var n = 0",
		},
		{
			Name: "var_without_init",
			Node: &ast.VarStmt{Name: "last"},
			Want: "var last",
		},
		{
			Name: "assign_ident",
			Node: &ast.AssignStmt{Target: tests_utils.Id("n"), Value: tests_utils.Bin(lexer.Plus, tests_utils.Id("n"), tests_utils.Num(1))},
			Want: "n = n + 1",
		},
		{
			Name: "assign_nested_target",
			Node: &ast.AssignStmt{
				Target: tests_utils.Member(tests_utils.Index(tests_utils.Member(tests_utils.Id("data"), "items"), tests_utils.Num(0)), "name"),
				Value:  tests_utils.Str("x"),
			},
			Want: `data.items[0].name = "x"`,
		},
		{
			Name: "call_stmt",
			Node: &ast.ExprStmt{X: tests_utils.Call(tests_utils.Member(a, "b"), tests_utils.Num(1))},
			Want: "a.b(1)",
		},
		{
			Name: "return_bare",
			Node: &ast.ReturnStmt{},
			Want: "return",
		},
		{
			Name: "return_value",
			Node: &ast.ReturnStmt{Value: tests_utils.Obj(ast.Entry{Key: "total", Value: tests_utils.Id("total")})},
			Want: "return {total: total}",
		},
		{
			Name: "exit_bare",
			Node: &ast.ReturnStmt{Exit: true},
			Want: "exit",
		},
		{
			Name: "exit_value",
			Node: &ast.ReturnStmt{Exit: true, Value: tests_utils.Bin(lexer.Plus, a, b)},
			Want: "exit a + b",
		},
		{
			Name: "if_empty",
			Node: &ast.IfStmt{Cond: a, Then: tests_utils.Blk()},
			Want: "if a {\n}",
		},
		{
			Name: "if_else",
			Node: &ast.IfStmt{
				Cond: tests_utils.Bin(lexer.And, tests_utils.Bin(lexer.Lt, tests_utils.Id("n"), tests_utils.Num(3)), tests_utils.Un(lexer.Not, &ast.BoolLit{})),
				Then: tests_utils.Blk(&ast.AssignStmt{Target: tests_utils.Id("n"), Value: tests_utils.Bin(lexer.Plus, tests_utils.Id("n"), tests_utils.Num(1))}),
				Else: tests_utils.Blk(&ast.ReturnStmt{}),
			},
			Want: "if n < 3 and not false {\n    n = n + 1\n} else {\n    return\n}",
		},
		{
			Name: "if_else_if_else",
			Node: &ast.IfStmt{
				Cond: a,
				Then: tests_utils.Blk(&ast.AssignStmt{Target: tests_utils.Id("x"), Value: tests_utils.Num(1)}),
				Else: &ast.IfStmt{
					Cond: b,
					Then: tests_utils.Blk(&ast.AssignStmt{Target: tests_utils.Id("x"), Value: tests_utils.Num(2)}),
					Else: tests_utils.Blk(&ast.AssignStmt{Target: tests_utils.Id("x"), Value: tests_utils.Num(3)}),
				},
			},
			Want: "if a {\n    x = 1\n} else if b {\n    x = 2\n} else {\n    x = 3\n}",
		},
		{
			Name: "if_object_header_grouped",
			Node: &ast.IfStmt{Cond: tests_utils.Obj(ast.Entry{Key: "a", Value: tests_utils.Num(1)}), Then: tests_utils.Blk()},
			Want: "if ({a: 1}) {\n}",
		},
		{
			Name: "if_object_first_operand_grouped",
			Node: &ast.IfStmt{Cond: tests_utils.Bin(lexer.Eq, tests_utils.Member(tests_utils.Obj(ast.Entry{Key: "a", Value: tests_utils.Num(1)}), "a"), tests_utils.Num(1)), Then: tests_utils.Blk()},
			Want: "if ({a: 1}.a == 1) {\n}",
		},
		{
			Name: "if_object_right_operand_kept",
			Node: &ast.IfStmt{Cond: tests_utils.Bin(lexer.In, tests_utils.Str("name"), tests_utils.Obj(ast.Entry{Key: "name", Value: tests_utils.Num(1)})), Then: tests_utils.Blk()},
			Want: "if \"name\" in {name: 1} {\n}",
		},
		{
			Name: "for_single",
			Node: &ast.ForInStmt{Value: "x", Iter: a, Body: tests_utils.Blk()},
			Want: "for x in a {\n}",
		},
		{
			Name: "for_pair",
			Node: &ast.ForInStmt{Key: "i", Value: "v", Iter: a, Body: tests_utils.Blk()},
			Want: "for i, v in a {\n}",
		},
		{
			Name: "for_var_pair",
			Node: &ast.ForInStmt{Mutable: true, Key: "i", Value: "v", Iter: tests_utils.Arr(tests_utils.Num(1), tests_utils.Num(2)), Body: tests_utils.Blk(&ast.ContinueStmt{})},
			Want: "for var i, v in [1, 2] {\n    continue\n}",
		},
		{
			Name: "for_discard_index",
			Node: &ast.ForInStmt{Key: "_", Value: "v", Iter: a, Body: tests_utils.Blk(&ast.BreakStmt{})},
			Want: "for _, v in a {\n    break\n}",
		},
		{
			Name: "for_object_header_grouped",
			Node: &ast.ForInStmt{Key: "k", Value: "v", Iter: tests_utils.Obj(ast.Entry{Key: "a", Value: tests_utils.Num(1)}), Body: tests_utils.Blk()},
			Want: "for k, v in ({a: 1}) {\n}",
		},
		{
			Name: "for_range",
			Node: &ast.ForRangeStmt{Name: "i", Start: tests_utils.Num(0), End: tests_utils.Num(10), Body: tests_utils.Blk()},
			Want: "for i range [0..10] {\n}",
		},
		{
			Name: "for_var_range_expr_bounds",
			Node: &ast.ForRangeStmt{
				Mutable: true,
				Name:    "i",
				Start:   tests_utils.Bin(lexer.Plus, a, tests_utils.Num(1)),
				End:     tests_utils.Bin(lexer.Star, b, tests_utils.Num(2)),
				Body:    tests_utils.Blk(&ast.BreakStmt{}),
			},
			Want: "for var i range [a + 1..b * 2] {\n    break\n}",
		},
		{
			Name: "nested_blocks_indent",
			Node: &ast.ForInStmt{Value: "x", Iter: a, Body: tests_utils.Blk(
				&ast.IfStmt{Cond: tests_utils.Id("x"), Then: tests_utils.Blk(&ast.BreakStmt{}), Else: tests_utils.Blk(
					&ast.ForRangeStmt{Name: "i", Start: tests_utils.Num(0), End: tests_utils.Num(3), Body: tests_utils.Blk(&ast.ContinueStmt{})},
				)},
				&ast.LetStmt{Name: "y", Value: tests_utils.Id("x")},
			)},
			Want: "for x in a {\n    if x {\n        break\n    } else {\n        for i range [0..3] {\n            continue\n        }\n    }\n    let y = x\n}",
		},
	}
}

func scriptCases() []tests_utils.StringCase {
	return []tests_utils.StringCase{
		{
			Name: "block_empty",
			Node: tests_utils.Blk(),
			Want: "{\n}",
		},
		{
			Name: "block_stmts",
			Node: tests_utils.Blk(&ast.LetStmt{Name: "a", Value: tests_utils.Num(1)}, &ast.VarStmt{Name: "b"}),
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
				&ast.LetStmt{Name: "a", Value: tests_utils.Num(1)},
				&ast.IfStmt{Cond: tests_utils.Id("a"), Then: tests_utils.Blk(&ast.LetStmt{Name: "b", Value: tests_utils.Str("x")})},
				&ast.ReturnStmt{Value: tests_utils.Id("a")},
			}},
			Want: "let a = 1\nif a {\n    let b = \"x\"\n}\nreturn a",
		},
	}
}

func TestString(t *testing.T) {
	// --- primary
	tests_utils.RunString(t, primaryCases())

	// --- array, object, entry, template part
	tests_utils.RunString(t, arrayObjectCases())

	// --- operators, minimal grouping by precedence
	tests_utils.RunString(t, operatorCases())

	// --- postfix_expr
	tests_utils.RunString(t, postfixCases())

	// --- stmt
	tests_utils.RunString(t, stmtCases())

	// --- block and script
	tests_utils.RunString(t, scriptCases())
}

func TestStringIsEbnf(t *testing.T) {
	// --- expressions reparse to the same string
	var exprs []tests_utils.StringCase
	for _, group := range [][]tests_utils.StringCase{primaryCases(), arrayObjectCases(), operatorCases(), postfixCases()} {
		for _, c := range group {
			if _, ok := c.Node.(ast.Expr); ok {
				exprs = append(exprs, c)
			}
		}
	}
	tests_utils.RunStringParses(t, exprs)

	// --- statements reparse to the same string
	var stmts []tests_utils.StringCase
	for _, c := range append(stmtCases(), scriptCases()...) {
		switch c.Node.(type) {
		case *ast.Script, *ast.LetStmt, *ast.VarStmt, *ast.AssignStmt, *ast.ExprStmt, *ast.IfStmt, *ast.ForInStmt, *ast.ForRangeStmt, *ast.ReturnStmt:
			stmts = append(stmts, c)
		}
	}
	tests_utils.RunStringParses(t, stmts)
}

func TestStringScripts(t *testing.T) {
	// --- script parses to its canonical ebnf form
	tests_utils.RunAst(t, []tests_utils.AstCase{
		{
			Name: "sample",
			Src:  tests_utils.Sample,
			Want: tests_utils.SampleCanonical,
		},
		{
			Name: "sample_crlf",
			Src:  tests_utils.SampleCRLF,
			Want: tests_utils.SampleCRLFCanonical,
		},
	})

	// --- canonical form is a fixed point
	tests_utils.RunOK(t, []tests_utils.ParseOkCase{
		{
			Name: "spec_def",
			Src:  tests_utils.SpecDef,
		},
	})
}

func TestPos(t *testing.T) {
	// --- statement lines
	tests_utils.RunPos(t, []tests_utils.PosCase{
		{
			Name:  "sample",
			Src:   tests_utils.Sample,
			Lines: []int{2, 3, 5, 13, 17},
		},
		{
			Name:  "sample_crlf",
			Src:   tests_utils.SampleCRLF,
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
	tests_utils.RunPosLineOnly(t, []tests_utils.PosCase{
		{
			Name: "sample",
			Src:  tests_utils.Sample,
		},
		{
			Name: "spec_def",
			Src:  tests_utils.SpecDef,
		},
	})
}

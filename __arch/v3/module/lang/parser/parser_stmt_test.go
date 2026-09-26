package parser_test

import (
	"testing"

	"github.com/siper92/akha/lang/tests"
)

func TestScriptLayout(t *testing.T) {
	// --- script and line
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "empty_source",
			Src:  "",
			Want: "",
		},
		{
			Name: "only_comments",
			Src:  "// a\n// b\n",
			Want: "",
		},
		{
			Name: "blank_lines_ignored",
			Src:  "\n\nlet a = 1\n\n\n",
			Want: "let a = 1",
		},
		{
			Name: "no_trailing_newline",
			Src:  "let a = 1",
			Want: "let a = 1",
		},
		{
			Name: "trailing_comment",
			Src:  "let a = 1 // c\n",
			Want: "let a = 1",
		},
		{
			Name: "two_statements",
			Src:  "let a = 1\nvar b = 2\n",
			Want: "let a = 1\nvar b = 2",
		},
		{
			Name: "indentation_ignored",
			Src:  "    let a = 1\n\t\tvar b\n",
			Want: "let a = 1\nvar b",
		},
		{
			Name: "blank_line_between_statements_dropped",
			Src:  "let a = 1\n\n\nvar b\n",
			Want: "let a = 1\nvar b",
		},
	})

	// --- block
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "empty_block",
			Src:  "if a {\n}\n",
			Want: "if a {\n}",
		},
		{
			Name: "block_blank_lines",
			Src:  "if a {\n\n\n}\n",
			Want: "if a {\n}",
		},
		{
			Name: "block_only_comment",
			Src:  "if a {\n    // c\n}\n",
			Want: "if a {\n}",
		},
		{
			Name: "block_one_stmt",
			Src:  "if a {\n    let b = 1\n}\n",
			Want: "if a {\n    let b = 1\n}",
		},
		{
			Name: "block_comment_after_open",
			Src:  "if a { // c\n}\n",
			Want: "if a {\n}",
		},
		{
			Name: "block_indent_normalized",
			Src:  "if a {\nlet b = 1\n\t\tvar c\n}\n",
			Want: "if a {\n    let b = 1\n    var c\n}",
		},
		{
			Name: "nested_block_indent",
			Src:  "if a {\n  if b {\n let c = 1\n  }\n}\n",
			Want: "if a {\n    if b {\n        let c = 1\n    }\n}",
		},
	})

	// --- layout errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "two_statements_one_line",
			Src:  "let a = 1 let b = 2",
			Line: 1,
		},
		{
			Name: "open_brace_next_line",
			Src:  "if a\n{\n}\n",
		},
		{
			Name: "close_brace_after_stmt",
			Src:  "if a {\nlet b = 1 }\n",
			Line: 2,
		},
		{
			Name: "block_on_one_line",
			Src:  "if a {}",
			Line: 1,
		},
		{
			Name: "stmt_on_header_line",
			Src:  "if a { let b = 1\n}\n",
			Line: 1,
		},
		{
			Name: "unterminated_block",
			Src:  "if a {\nlet b = 1\n",
		},
		{
			Name: "stray_close_brace",
			Src:  "}",
			Line: 1,
		},
		{
			Name: "stmt_after_close_brace",
			Src:  "if a {\n} let b = 1\n",
			Line: 2,
		},
		{
			Name: "newline_outside_brackets",
			Src:  "let a = 1 +\n2\n",
		},
	})
}

func TestLet(t *testing.T) {
	// --- let_stmt
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "let_number",
			Src:  "let a = 1",
			Want: "let a = 1",
		},
		{
			Name: "let_string",
			Src:  `let name = "ak"`,
			Want: `let name = "ak"`,
		},
		{
			Name: "let_ident",
			Src:  "let a = b",
			Want: "let a = b",
		},
		{
			Name: "let_discard",
			Src:  "let _ = count * 2",
			Want: "let _ = count * 2",
		},
		{
			Name: "let_case_sensitive_name",
			Src:  "let Name = 1",
			Want: "let Name = 1",
		},
		{
			Name: "let_spacing_normalized",
			Src:  "let   a=1",
			Want: "let a = 1",
		},
	})

	// --- let errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "let_without_init",
			Src:  "let x",
			Line: 1,
		},
		{
			Name: "let_missing_expr",
			Src:  "let x =",
			Line: 1,
		},
		{
			Name: "let_missing_name",
			Src:  "let = 1",
			Line: 1,
		},
		{
			Name: "let_missing_eq",
			Src:  "let a 1",
			Line: 1,
		},
		{
			Name: "let_two_names",
			Src:  "let a, b = 1",
			Line: 1,
		},
		{
			Name: "let_two_decls",
			Src:  "let a = 1, b = 2",
			Line: 1,
		},
		{
			Name: "let_number_name",
			Src:  "let 1 = 2",
			Line: 1,
		},
		{
			Name: "let_member_name",
			Src:  "let a.b = 1",
			Line: 1,
		},
		{
			Name: "let_without_init_second_line",
			Src:  "let a = 1\nlet x\n",
			Line: 2,
		},
	})
}

func TestVar(t *testing.T) {
	// --- var_stmt
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "var_with_init",
			Src:  "var n = 0",
			Want: "var n = 0",
		},
		{
			Name: "var_without_init",
			Src:  "var last",
			Want: "var last",
		},
		{
			Name: "var_array",
			Src:  "var a = [1, 2]",
			Want: "var a = [1, 2]",
		},
		{
			Name: "var_without_init_comment",
			Src:  "var last // null",
			Want: "var last",
		},
	})

	// --- var errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "var_missing_expr",
			Src:  "var x =",
			Line: 1,
		},
		{
			Name: "var_two_names",
			Src:  "var a, b",
			Line: 1,
		},
		{
			Name: "var_two_decls",
			Src:  "var a = 1, b = 2",
			Line: 1,
		},
		{
			Name: "var_missing_name",
			Src:  "var = 1",
			Line: 1,
		},
	})
}

func TestAssign(t *testing.T) {
	// --- assign_stmt
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "assign_ident",
			Src:  "x = 1",
			Want: "x = 1",
		},
		{
			Name: "assign_expr",
			Src:  "n = n + 1",
			Want: "n = n + 1",
		},
		{
			Name: "assign_nested_target",
			Src:  `data.items[0]["name"] = "x"`,
			Want: `data.items[0]["name"] = "x"`,
		},
		{
			Name: "assign_multi_line_value",
			Src:  "x = [\n    1,\n    2,\n]\n",
			Want: "x = [1, 2]",
		},
	})

	// --- target
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "target_member",
			Src:  "state.count = state.count + 1",
		},
		{
			Name: "target_index",
			Src:  "a[0] = 1",
		},
		{
			Name: "target_string_index",
			Src:  `state["count"] = 2`,
		},
		{
			Name: "target_expr_index",
			Src:  "a[i + 1] = b",
		},
		{
			Name: "target_nested",
			Src:  "data.items[0].name = x",
		},
		{
			Name: "target_deep_member",
			Src:  "a.b.c = f(1)",
		},
	})

	// --- invalid targets
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "number_target",
			Src:  "1 = 2",
			Line: 1,
		},
		{
			Name: "string_target",
			Src:  `"a" = 1`,
			Line: 1,
		},
		{
			Name: "call_target",
			Src:  "f() = 1",
			Line: 1,
		},
		{
			Name: "method_call_target",
			Src:  "a.b() = 1",
			Line: 1,
		},
		{
			Name: "binary_target",
			Src:  "a + b = 1",
			Line: 1,
		},
		{
			Name: "grouped_target",
			Src:  "(a) = 1",
			Line: 1,
		},
		{
			Name: "array_literal_target",
			Src:  "[1][0] = 2",
			Line: 1,
		},
		{
			Name: "discard_target",
			Src:  "_ = 1",
			Line: 1,
		},
		{
			Name: "missing_value",
			Src:  "x =",
			Line: 1,
		},
		{
			Name: "chained_assign",
			Src:  "a = b = 1",
			Line: 1,
		},
		{
			Name: "slice_target",
			Src:  "a[1:2] = x",
			Line: 1,
			Msg:  "slices are not supported in v1",
		},
	})
}

func TestCallStmt(t *testing.T) {
	// --- call_stmt
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "call_no_args",
			Src:  "f()",
		},
		{
			Name: "call_args",
			Src:  "f(1, 2)",
		},
		{
			Name: "call_trailing_comma",
			Src:  "f(1, 2,)",
		},
		{
			Name: "call_member",
			Src:  "a.b(1)",
		},
		{
			Name: "call_index",
			Src:  "a[0](1)",
		},
		{
			Name: "call_chain",
			Src:  "f(1)(2)",
		},
		{
			Name: "call_mixed_args",
			Src:  "f(a + 1, [1, 2], {k: 1})",
		},
		{
			Name: "call_comment",
			Src:  "f() // c",
		},
	})

	tests.RunAst(t, []tests.AstCase{
		{
			Name: "call_canonical",
			Src:  "f( a+1 , [1,2] , {k:1} )",
			Want: "f(a + 1, [1, 2], {k: 1})",
		},
		{
			Name: "call_split_lines_canonical",
			Src:  "f(\n    1,\n    2,\n)\n",
			Want: "f(1, 2)",
		},
	})

	// --- call multi line
	tests.RunSame(t, []tests.ParseSameCase{
		{
			Name: "call_split_lines",
			Src:  "f(\n    1,\n    2,\n)\n",
			Same: "f(1, 2)",
		},
		{
			Name: "call_trailing_comma_same",
			Src:  "f(1,)",
			Same: "f(1)",
		},
	})

	// --- unused value
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "unused_binary",
			Src:  "1 + 2",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_ident",
			Src:  "x",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_member",
			Src:  "a.b",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_index",
			Src:  "a[0]",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_string",
			Src:  `"s"`,
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_array",
			Src:  "[1, 2]",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_index_after_call",
			Src:  "f()[0]",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_member_after_call",
			Src:  "f().x",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_call_in_binary",
			Src:  "f() + 1",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_negated_call",
			Src:  "-f()",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_not_call",
			Src:  "not f()",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_comparison",
			Src:  "a == b",
			Line: 1,
			Msg:  "unused value",
		},
		{
			Name: "unused_second_line",
			Src:  "let a = 1\na\n",
			Line: 2,
			Msg:  "unused value",
		},
	})

	// --- call syntax errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "call_only_comma",
			Src:  "f(,)",
			Line: 1,
		},
		{
			Name: "call_double_comma",
			Src:  "f(1,,2)",
			Line: 1,
		},
		{
			Name: "call_missing_comma",
			Src:  "f(1 2)",
			Line: 1,
		},
		{
			Name: "call_unclosed",
			Src:  "f(1",
		},
	})
}

func TestIf(t *testing.T) {
	// --- if_stmt
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "if_empty",
			Src:  "if a {\n}\n",
			Want: "if a {\n}",
		},
		{
			Name: "if_else",
			Src:  "if a {\n    let b = 1\n} else {\n    return\n}\n",
			Want: "if a {\n    let b = 1\n} else {\n    return\n}",
		},
		{
			Name: "if_sample_header",
			Src:  "if n < 3 and not false {\n    n = n + 1\n} else {\n    return\n}\n",
			Want: "if n < 3 and not false {\n    n = n + 1\n} else {\n    return\n}",
		},
		{
			Name: "if_else_if_else",
			Src:  "if a {\n    x = 1\n} else if b {\n    x = 2\n} else {\n    x = 3\n}\n",
			Want: "if a {\n    x = 1\n} else if b {\n    x = 2\n} else {\n    x = 3\n}",
		},
		{
			Name: "if_header_group_dropped",
			Src:  "if (a) {\n}\n",
			Want: "if a {\n}",
		},
		{
			Name: "if_header_object_group_kept",
			Src:  "if ({a: 1}) {\n}\n",
			Want: "if ({a: 1}) {\n}",
		},
		{
			Name: "if_header_object_operand_group_dropped",
			Src:  "if (\"name\" in {name: 1}) {\n}\n",
			Want: "if \"name\" in {name: 1} {\n}",
		},
		{
			Name: "if_header_multi_line_in_parens",
			Src:  "if (a and\n    b) {\n}\n",
			Want: "if a and b {\n}",
		},
	})

	// --- else if
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "else_if",
			Src:  "if a {\n} else if b {\n}\n",
		},
		{
			Name: "else_if_else",
			Src:  "if a {\n    x = 1\n} else if b {\n    x = 2\n} else {\n    x = 3\n}\n",
		},
		{
			Name: "many_else_if",
			Src:  "if a {\n} else if b {\n} else if c {\n} else if d {\n} else {\n}\n",
		},
		{
			Name: "nested_if",
			Src:  "if a {\n    if b {\n    } else {\n    }\n}\n",
		},
	})

	// --- header
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "header_grouped_object_in",
			Src:  "if (\"name\" in {name: 1}) {\n}\n",
		},
		{
			Name: "header_grouped_object",
			Src:  "if ({a: 1}) {\n}\n",
		},
		{
			Name: "header_call",
			Src:  "if f(1) {\n}\n",
		},
		{
			Name: "header_array",
			Src:  "if [1] {\n}\n",
		},
		{
			Name: "header_multi_line_in_parens",
			Src:  "if (a and\n    b) {\n}\n",
		},
	})

	tests.RunSame(t, []tests.ParseSameCase{
		{
			Name: "header_grouping_dropped",
			Src:  "if (a) {\n}\n",
			Same: "if a {\n}\n",
		},
	})

	// --- if errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "else_on_new_line",
			Src:  "if a {\n}\nelse {\n}\n",
			Line: 3,
		},
		{
			Name: "else_brace_next_line",
			Src:  "if a {\n} else\n{\n}\n",
		},
		{
			Name: "two_else",
			Src:  "if a {\n} else {\n} else {\n}\n",
			Line: 3,
		},
		{
			Name: "dangling_else",
			Src:  "else {\n}\n",
			Line: 1,
		},
		{
			Name: "missing_header",
			Src:  "if {\n}\n",
			Line: 1,
		},
		{
			Name: "object_literal_header",
			Src:  "if {a: 1} {\n}\n",
			Line: 1,
		},
		{
			Name: "missing_block",
			Src:  "if a",
		},
		{
			Name: "header_split_without_parens",
			Src:  "if a and\nb {\n}\n",
		},
		{
			Name: "else_if_missing_header",
			Src:  "if a {\n} else if {\n}\n",
			Line: 2,
		},
	})
}

func TestFor(t *testing.T) {
	// --- for_stmt in
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "for_var_pair_sample",
			Src:  "for var i, v in [1, 2] {\n    continue\n}\n",
			Want: "for var i, v in [1, 2] {\n    continue\n}",
		},
		{
			Name: "for_single",
			Src:  "for x in a {\n}\n",
			Want: "for x in a {\n}",
		},
		{
			Name: "for_pair",
			Src:  "for i, v in a {\n}\n",
			Want: "for i, v in a {\n}",
		},
		{
			Name: "for_var_single",
			Src:  "for var x in a {\n}\n",
			Want: "for var x in a {\n}",
		},
		{
			Name: "for_discard_index",
			Src:  "for _, v in a {\n}\n",
			Want: "for _, v in a {\n}",
		},
		{
			Name: "for_header_object_group_kept",
			Src:  "for k, v in ({a: 1}) {\n}\n",
			Want: "for k, v in ({a: 1}) {\n}",
		},
		{
			Name: "for_multi_line_array",
			Src:  "for x in [\n    1,\n    2,\n] {\n}\n",
			Want: "for x in [1, 2] {\n}",
		},
	})

	// --- for_stmt range
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "range_literal",
			Src:  "for i range [0..10] {\n}\n",
			Want: "for i range [0..10] {\n}",
		},
		{
			Name: "range_var",
			Src:  "for var i range [0..10] {\n}\n",
			Want: "for var i range [0..10] {\n}",
		},
		{
			Name: "range_expr_bounds",
			Src:  "for i range [a + 1..b * 2] {\n}\n",
			Want: "for i range [a + 1..b * 2] {\n}",
		},
		{
			Name: "range_grouped_bound_dropped",
			Src:  "for i range [(0)..(n + 1)] {\n    break\n}\n",
			Want: "for i range [0..n + 1] {\n    break\n}",
		},
	})

	// --- for_stmt range
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "range_literal",
			Src:  "for i range [0..10] {\n}\n",
		},
		{
			Name: "range_var",
			Src:  "for var i range [0..10] {\n}\n",
		},
		{
			Name: "range_ident_end",
			Src:  "for n range [0..count] {\n    total = total + n\n}\n",
		},
		{
			Name: "range_expr_bounds",
			Src:  "for i range [a + 1..b * 2] {\n}\n",
		},
		{
			Name: "range_discard",
			Src:  "for _ range [0..3] {\n}\n",
		},
	})

	// --- for header
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "for_multi_line_array",
			Src:  "for x in [\n    1,\n    2,\n] {\n}\n",
		},
		{
			Name: "for_grouped_object",
			Src:  "for k, v in ({a: 1}) {\n}\n",
		},
		{
			Name: "for_member_collection",
			Src:  "for x in input.items {\n}\n",
		},
		{
			Name: "for_nested",
			Src:  "for x in a {\n    for y in x {\n    }\n}\n",
		},
	})

	tests.RunSame(t, []tests.ParseSameCase{
		{
			Name: "range_bound_precedence",
			Src:  "for i range [0..n + 1] {\n}\n",
			Same: "for i range [0..(n + 1)] {\n}\n",
		},
	})

	tests.RunDiff(t, []tests.ParseSameCase{
		{
			Name: "range_var_is_kept",
			Src:  "for var i range [0..3] {\n}\n",
			Same: "for i range [0..3] {\n}\n",
		},
		{
			Name: "in_var_is_kept",
			Src:  "for var x in a {\n}\n",
			Same: "for x in a {\n}\n",
		},
		{
			Name: "range_differs_from_in",
			Src:  "for i range [0..3] {\n}\n",
			Same: "for i in [0, 3] {\n}\n",
		},
	})

	// --- for errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "for_let",
			Src:  "for let x in a {\n}\n",
			Line: 1,
		},
		{
			Name: "for_let_range",
			Src:  "for let i range [0..3] {\n}\n",
			Line: 1,
		},
		{
			Name: "for_missing_in",
			Src:  "for x a {\n}\n",
			Line: 1,
		},
		{
			Name: "for_three_vars",
			Src:  "for a, b, c in x {\n}\n",
			Line: 1,
		},
		{
			Name: "for_number_var",
			Src:  "for 1 in a {\n}\n",
			Line: 1,
		},
		{
			Name: "for_missing_collection",
			Src:  "for x in {\n}\n",
			Line: 1,
		},
		{
			Name: "for_object_literal_header",
			Src:  "for x in {a: 1} {\n}\n",
			Line: 1,
		},
		{
			Name: "for_missing_block",
			Src:  "for x in a",
		},
		{
			Name: "range_two_vars",
			Src:  "for i, v range [0..3] {\n}\n",
			Line: 1,
		},
		{
			Name: "range_no_brackets",
			Src:  "for i range 0..10 {\n}\n",
			Line: 1,
		},
		{
			Name: "range_missing_end",
			Src:  "for i range [0..] {\n}\n",
			Line: 1,
		},
		{
			Name: "range_missing_start",
			Src:  "for i range [..3] {\n}\n",
			Line: 1,
		},
		{
			Name: "range_comma",
			Src:  "for i range [0, 10] {\n}\n",
			Line: 1,
		},
	})
}

func TestBreakContinue(t *testing.T) {
	// --- break_stmt and continue_stmt
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "break_in_for",
			Src:  "for x in a {\n    break\n}\n",
			Want: "for x in a {\n    break\n}",
		},
		{
			Name: "continue_in_for",
			Src:  "for x in a {\n    continue\n}\n",
			Want: "for x in a {\n    continue\n}",
		},
	})

	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "break_in_if_in_loop",
			Src:  "for x in a {\n    if x {\n        break\n    }\n}\n",
		},
		{
			Name: "continue_in_else_in_loop",
			Src:  "for x in a {\n    if x {\n    } else if y {\n        continue\n    } else {\n        continue\n    }\n}\n",
		},
		{
			Name: "break_in_range",
			Src:  "for i range [0..3] {\n    break\n}\n",
		},
		{
			Name: "break_in_inner_loop",
			Src:  "if a {\n    for x in a {\n        for y in x {\n            break\n        }\n        continue\n    }\n}\n",
		},
	})

	// --- outside a loop
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "break_top_level",
			Src:  "break",
			Line: 1,
		},
		{
			Name: "continue_top_level",
			Src:  "continue",
			Line: 1,
		},
		{
			Name: "break_in_if",
			Src:  "if a {\n    break\n}\n",
			Line: 2,
		},
		{
			Name: "continue_in_else",
			Src:  "if a {\n} else {\n    continue\n}\n",
			Line: 3,
		},
		{
			Name: "break_after_loop",
			Src:  "for x in a {\n}\nbreak\n",
			Line: 3,
		},
		{
			Name: "break_with_value",
			Src:  "for x in a {\n    break 1\n}\n",
			Line: 2,
		},
	})
}

func TestReturn(t *testing.T) {
	// --- return_stmt
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "return_bare",
			Src:  "return",
			Want: "return",
		},
		{
			Name: "return_number",
			Src:  "return 1",
			Want: "return 1",
		},
		{
			Name: "return_expr",
			Src:  "return a + 1",
			Want: "return a + 1",
		},
		{
			Name: "return_object",
			Src:  "return {name: name, total: n}",
			Want: "return {name: name, total: n}",
		},
	})

	// --- exit alias keeps its keyword
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "exit_bare",
			Src:  "exit",
			Want: "exit",
		},
		{
			Name: "exit_value",
			Src:  "exit 1",
			Want: "exit 1",
		},
		{
			Name: "exit_expr",
			Src:  "exit a + b",
			Want: "exit a + b",
		},
	})

	tests.RunDiff(t, []tests.ParseSameCase{
		{
			Name: "exit_is_not_printed_as_return",
			Src:  "exit",
			Same: "return",
		},
	})

	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "return_in_loop",
			Src:  "for x in a {\n    if x {\n        return x\n    }\n}\n",
		},
		{
			Name: "exit_in_range",
			Src:  "for i range [0..3] {\n    exit\n}\n",
		},
		{
			Name: "return_multi_line_array",
			Src:  "return [\n    1,\n]\n",
		},
	})

	// --- return errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "return_two_values",
			Src:  "return 1 2",
			Line: 1,
		},
		{
			Name: "return_comma_values",
			Src:  "return 1, 2",
			Line: 1,
		},
		{
			Name: "exit_exit",
			Src:  "exit exit",
			Line: 1,
		},
	})
}

func TestNames(t *testing.T) {
	// --- keywords and reserved words as names
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "let_if",
			Src:  "let if = 1",
			Line: 1,
		},
		{
			Name: "let_fn",
			Src:  "let fn = 1",
			Line: 1,
		},
		{
			Name: "let_try",
			Src:  "let try = 1",
			Line: 1,
		},
		{
			Name: "let_catch",
			Src:  "let catch = 1",
			Line: 1,
		},
		{
			Name: "let_let",
			Src:  "let let = 1",
			Line: 1,
		},
		{
			Name: "let_true",
			Src:  "let true = 1",
			Line: 1,
		},
		{
			Name: "let_null",
			Src:  "let null = 1",
			Line: 1,
		},
		{
			Name: "let_and",
			Src:  "let and = 1",
			Line: 1,
		},
		{
			Name: "let_in",
			Src:  "let in = 1",
			Line: 1,
		},
		{
			Name: "let_range",
			Src:  "let range = 1",
			Line: 1,
		},
		{
			Name: "let_exit",
			Src:  "let exit = 1",
			Line: 1,
		},
		{
			Name: "var_for",
			Src:  "var for",
			Line: 1,
		},
		{
			Name: "var_fn",
			Src:  "var fn = 1",
			Line: 1,
		},
		{
			Name: "for_fn",
			Src:  "for fn in a {\n}\n",
			Line: 1,
		},
		{
			Name: "for_second_try",
			Src:  "for i, try in a {\n}\n",
			Line: 1,
		},
		{
			Name: "range_catch",
			Src:  "for catch range [0..3] {\n}\n",
			Line: 1,
		},
		{
			Name: "assign_fn",
			Src:  "fn = 1",
			Line: 1,
		},
		{
			Name: "keyword_name_second_line",
			Src:  "let a = 1\nlet if = 2\n",
			Line: 2,
		},
	})

	// --- keywords are case insensitive
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "let_title_true",
			Src:  "let True = 1",
			Line: 1,
		},
		{
			Name: "let_title_if",
			Src:  "let If = 1",
			Line: 1,
		},
		{
			Name: "let_upper_null",
			Src:  "let NULL = 1",
			Line: 1,
		},
		{
			Name: "var_mixed_for",
			Src:  "var fOr",
			Line: 1,
		},
		{
			Name: "let_upper_fn",
			Src:  "let FN = 1",
			Line: 1,
		},
		{
			Name: "for_title_try",
			Src:  "for Try in a {\n}\n",
			Line: 1,
		},
		{
			Name: "assign_upper_catch",
			Src:  "CATCH = 1",
			Line: 1,
		},
		{
			Name: "upper_break_outside_loop",
			Src:  "BREAK",
			Line: 1,
		},
	})

	tests.RunSame(t, []tests.ParseSameCase{
		{
			Name: "upper_let",
			Src:  "LET a = 1",
			Same: "let a = 1",
		},
		{
			Name: "title_var",
			Src:  "Var a",
			Same: "var a",
		},
		{
			Name: "upper_if_else_if_else",
			Src:  "IF a {\n} ELSE IF b {\n} Else {\n}\n",
			Same: "if a {\n} else if b {\n} else {\n}\n",
		},
		{
			Name: "mixed_for_var_in",
			Src:  "FOR VAR x IN a {\n    Continue\n}\n",
			Same: "for var x in a {\n    continue\n}\n",
		},
		{
			Name: "upper_range",
			Src:  "for i RANGE [0..3] {\n    BREAK\n}\n",
			Same: "for i range [0..3] {\n    break\n}\n",
		},
		{
			Name: "title_return",
			Src:  "Return 1",
			Same: "return 1",
		},
		{
			Name: "upper_exit",
			Src:  "EXIT",
			Same: "exit",
		},
	})

	// --- valid names
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "capital_ident",
			Src:  "let Name = 1",
		},
		{
			Name: "upper_keyword_prefix",
			Src:  "let LETTER = 1",
		},
		{
			Name: "keyword_prefix",
			Src:  "let fn1 = 1",
		},
		{
			Name: "underscore_prefix",
			Src:  "let _x = 1",
		},
		{
			Name: "keyword_suffix",
			Src:  "var letter = 1",
		},
	})
}

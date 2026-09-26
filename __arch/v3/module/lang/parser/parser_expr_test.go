package parser_test

import (
	"testing"

	"github.com/siper92/akha/lang/tests"
)

func TestExprAst(t *testing.T) {
	// --- expr: canonical ebnf form
	tests.RunExprAst(t, []tests.AstCase{
		{
			Name: "add_mul",
			Src:  "1 + 2 * 3",
			Want: "1 + 2 * 3",
		},
		{
			Name: "and_cmp_not",
			Src:  "n < 3 and not false",
			Want: "n < 3 and not false",
		},
		{
			Name: "sub_left",
			Src:  "a - b - c",
			Want: "a - b - c",
		},
		{
			Name: "mul_div_mod_left",
			Src:  "a * b / c % d",
			Want: "a * b / c % d",
		},
		{
			Name: "or",
			Src:  "a or b",
			Want: "a or b",
		},
		{
			Name: "eq",
			Src:  "a == b",
			Want: "a == b",
		},
		{
			Name: "ne",
			Src:  "a != b",
			Want: "a != b",
		},
		{
			Name: "lt",
			Src:  "a < b",
			Want: "a < b",
		},
		{
			Name: "le",
			Src:  "a <= b",
			Want: "a <= b",
		},
		{
			Name: "gt",
			Src:  "a > b",
			Want: "a > b",
		},
		{
			Name: "ge",
			Src:  "a >= b",
			Want: "a >= b",
		},
		{
			Name: "in",
			Src:  "a in b",
			Want: "a in b",
		},
		{
			Name: "not_in",
			Src:  "a not in b",
			Want: "a not in b",
		},
		{
			Name: "mod",
			Src:  "10 % 3",
			Want: "10 % 3",
		},
		{
			Name: "neg",
			Src:  "-3.2",
			Want: "-3.2",
		},
	})

	// --- expr: redundant grouping is dropped
	tests.RunExprAst(t, []tests.AstCase{
		{
			Name: "group_left_assoc_dropped",
			Src:  "(a - b) - c",
			Want: "a - b - c",
		},
		{
			Name: "group_higher_prec_dropped",
			Src:  "1 + (2 * 3)",
			Want: "1 + 2 * 3",
		},
		{
			Name: "group_not_operand_dropped",
			Src:  "not (a == b)",
			Want: "not a == b",
		},
		{
			Name: "double_group_dropped",
			Src:  "((a))",
			Want: "a",
		},
		{
			Name: "group_neg_operand_dropped",
			Src:  "-(a[0])",
			Want: "-a[0]",
		},
		{
			Name: "group_postfix_dropped",
			Src:  "(((a.b)[0])(1)).c",
			Want: "a.b[0](1).c",
		},
	})

	// --- expr: needed grouping is kept
	tests.RunExprAst(t, []tests.AstCase{
		{
			Name: "group_right_sub_kept",
			Src:  "a - (b - c)",
			Want: "a - (b - c)",
		},
		{
			Name: "group_add_under_mul_kept",
			Src:  "(1 + 2) * 3",
			Want: "(1 + 2) * 3",
		},
		{
			Name: "group_not_left_of_eq_kept",
			Src:  "(not a) == b",
			Want: "(not a) == b",
		},
		{
			Name: "group_or_under_and_kept",
			Src:  "(a or b) and c",
			Want: "(a or b) and c",
		},
		{
			Name: "group_neg_then_index_kept",
			Src:  "(-a)[0]",
			Want: "(-a)[0]",
		},
		{
			Name: "group_cmp_chain_kept",
			Src:  "(a < b) < c",
			Want: "(a < b) < c",
		},
		{
			Name: "group_in_chain_kept",
			Src:  "a in (b in c)",
			Want: "a in (b in c)",
		},
		{
			Name: "group_neg_over_add_kept",
			Src:  "-(a + b)",
			Want: "-(a + b)",
		},
	})

	// --- expr: aliases and case normalize to the ebnf keywords
	tests.RunExprAst(t, []tests.AstCase{
		{
			Name: "and_alias",
			Src:  "a && b",
			Want: "a and b",
		},
		{
			Name: "or_alias",
			Src:  "a || b",
			Want: "a or b",
		},
		{
			Name: "not_alias",
			Src:  "!a",
			Want: "not a",
		},
		{
			Name: "aliases_mixed",
			Src:  "true && 1 < 2 || !b",
			Want: "true and 1 < 2 or not b",
		},
		{
			Name: "upper_keyword_ops",
			Src:  "a AND b OR Not c In d",
			Want: "a and b or not c in d",
		},
		{
			Name: "mixed_not_in",
			Src:  "a Not In b",
			Want: "a not in b",
		},
		{
			Name: "neg_neg_spaced",
			Src:  "- -a",
			Want: "- -a",
		},
		{
			Name: "whitespace_normalized",
			Src:  "a+b*c",
			Want: "a + b * c",
		},
	})
}

func TestPrecedence(t *testing.T) {
	// --- or_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "or_left_assoc",
			Src:  "a or b or c",
			Same: "(a or b) or c",
		},
		{
			Name: "or_lower_than_and",
			Src:  "a or b and c",
			Same: "a or (b and c)",
		},
		{
			Name: "or_alias",
			Src:  "a || b",
			Same: "a or b",
		},
	})

	// --- and_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "and_left_assoc",
			Src:  "a and b and c",
			Same: "(a and b) and c",
		},
		{
			Name: "and_lower_than_not",
			Src:  "not a and b",
			Same: "(not a) and b",
		},
		{
			Name: "and_alias",
			Src:  "a && b",
			Same: "a and b",
		},
		{
			Name: "aliases_mixed",
			Src:  "true && 1 < 2 || !b",
			Same: "(true and (1 < 2)) or (not b)",
		},
	})

	// --- not_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "not_over_eq",
			Src:  "not a == b",
			Same: "not (a == b)",
		},
		{
			Name: "not_over_in",
			Src:  "not a in b",
			Same: "not (a in b)",
		},
		{
			Name: "not_not",
			Src:  "not not a",
			Same: "not (not a)",
		},
		{
			Name: "bang_alias",
			Src:  "!a",
			Same: "not a",
		},
		{
			Name: "bang_over_eq",
			Src:  "!a == b",
			Same: "not (a == b)",
		},
		{
			Name: "bang_bang",
			Src:  "!!a",
			Same: "not (not a)",
		},
	})

	// --- in_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "in_lower_than_cmp",
			Src:  "a == b in c",
			Same: "(a == b) in c",
		},
		{
			Name: "in_lower_than_add",
			Src:  "a + 1 in b",
			Same: "(a + 1) in b",
		},
		{
			Name: "not_in_rhs_add",
			Src:  "a not in b + c",
			Same: "a not in (b + c)",
		},
		{
			Name: "in_under_and",
			Src:  "a in b and c not in d",
			Same: "(a in b) and (c not in d)",
		},
		{
			Name: "not_in_under_or",
			Src:  "x not in a or y",
			Same: "(x not in a) or y",
		},
	})

	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "grouped_in_chain_left",
			Src:  "let x = (a in b) in c",
		},
		{
			Name: "grouped_in_chain_right",
			Src:  "let x = a in (b in c)",
		},
		{
			Name: "not_in_string",
			Src:  `let x = "x" not in obj`,
		},
	})

	// --- cmp_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "cmp_lower_than_add",
			Src:  "a + 1 < b * 2",
			Same: "(a + 1) < (b * 2)",
		},
		{
			Name: "cmp_under_and",
			Src:  "a == b and c != d",
			Same: "(a == b) and (c != d)",
		},
		{
			Name: "cmp_ge_under_and_not",
			Src:  "count >= 3 and not (i == 1)",
			Same: "(count >= 3) and (not (i == 1))",
		},
	})

	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "grouped_cmp_chain_left",
			Src:  "let x = (a < b) < c",
		},
		{
			Name: "grouped_cmp_chain_right",
			Src:  "let x = a == (b == c)",
		},
	})

	// --- non associative levels
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "lt_chain",
			Src:  "let x = a < b < c",
			Line: 1,
		},
		{
			Name: "eq_chain",
			Src:  "let x = a == b == c",
			Line: 1,
		},
		{
			Name: "mixed_cmp_chain",
			Src:  "let x = a < b == c",
			Line: 1,
		},
		{
			Name: "ge_le_chain",
			Src:  "let x = a >= b <= c",
			Line: 1,
		},
		{
			Name: "in_chain",
			Src:  "let x = a in b in c",
			Line: 1,
		},
		{
			Name: "not_in_chain",
			Src:  "let x = a not in b in c",
			Line: 1,
		},
		{
			Name: "in_not_in_chain",
			Src:  "let x = a in b not in c",
			Line: 1,
		},
		{
			Name: "chain_in_if_header",
			Src:  "if a < b < c {\n}\n",
			Line: 1,
		},
	})

	// --- add_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "add_left_assoc",
			Src:  "a - b - c",
			Same: "(a - b) - c",
		},
		{
			Name: "add_sub_mixed",
			Src:  "a - b + c",
			Same: "(a - b) + c",
		},
		{
			Name: "add_lower_than_mul",
			Src:  "1 + 2 * 3",
			Same: "1 + (2 * 3)",
		},
	})

	// --- mul_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "div_left_assoc",
			Src:  "a / b / c",
			Same: "(a / b) / c",
		},
		{
			Name: "mod_mul_left",
			Src:  "a % b * c",
			Same: "(a % b) * c",
		},
		{
			Name: "mul_mod_left",
			Src:  "a * b % c",
			Same: "(a * b) % c",
		},
	})

	// --- unary_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "neg_over_mul",
			Src:  "-a * b",
			Same: "(-a) * b",
		},
		{
			Name: "neg_neg",
			Src:  "- -a",
			Same: "-(-a)",
		},
		{
			Name: "neg_under_index",
			Src:  "-a[0]",
			Same: "-(a[0])",
		},
		{
			Name: "neg_under_member",
			Src:  "-a.b",
			Same: "-(a.b)",
		},
		{
			Name: "neg_under_call",
			Src:  "-f(1)",
			Same: "-(f(1))",
		},
		{
			Name: "sub_neg",
			Src:  "a - -b",
			Same: "a - (-b)",
		},
	})

	// --- grouping changes the tree
	tests.RunDiffExpr(t, []tests.ParseSameCase{
		{
			Name: "grouped_add_then_mul",
			Src:  "(1 + 2) * 3",
			Same: "1 + 2 * 3",
		},
		{
			Name: "grouped_right_sub",
			Src:  "a - (b - c)",
			Same: "a - b - c",
		},
		{
			Name: "grouped_right_div",
			Src:  "a / (b / c)",
			Same: "a / b / c",
		},
		{
			Name: "grouped_neg_then_index",
			Src:  "(-a)[0]",
			Same: "-a[0]",
		},
		{
			Name: "grouped_not_then_eq",
			Src:  "(not a) == b",
			Same: "not a == b",
		},
		{
			Name: "grouped_or_then_and",
			Src:  "(a or b) and c",
			Same: "a or b and c",
		},
	})
}

func TestPostfix(t *testing.T) {
	// --- postfix_expr
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "member_left_assoc",
			Src:  "a.b.c",
			Same: "(a.b).c",
		},
		{
			Name: "index_left_assoc",
			Src:  "a[0][1]",
			Same: "(a[0])[1]",
		},
		{
			Name: "call_left_assoc",
			Src:  "f(1)(2)",
			Same: "(f(1))(2)",
		},
		{
			Name: "mixed_postfix",
			Src:  "a.b[0](1).c",
			Same: "(((a.b)[0])(1)).c",
		},
		{
			Name: "index_expr",
			Src:  "a[i + 1]",
			Same: "a[(i + 1)]",
		},
	})

	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "string_index",
			Src:  `let x = user["content-type"]`,
		},
		{
			Name: "deep_access",
			Src:  `let x = arr3[4][1]["2"]`,
		},
		{
			Name: "ident_index",
			Src:  "let x = user[key]",
		},
		{
			Name: "input_member",
			Src:  "let x = input.items",
		},
		{
			Name: "multi_line_index",
			Src:  "let x = a[\n    0\n]\n",
		},
	})

	// --- slices
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "slice_both",
			Src:  "let x = a[1:2]",
			Line: 1,
			Msg:  "slices are not supported in v1",
		},
		{
			Name: "slice_idents",
			Src:  "let x = a[i:j]",
			Line: 1,
			Msg:  "slices are not supported in v1",
		},
		{
			Name: "slice_open_end",
			Src:  "let x = a[1:]",
			Line: 1,
			Msg:  "slices are not supported in v1",
		},
		{
			Name: "slice_open_start",
			Src:  "let x = a[:2]",
			Line: 1,
		},
	})

	// --- postfix errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "member_missing_name",
			Src:  "let x = a.",
			Line: 1,
		},
		{
			Name: "member_bracket",
			Src:  "let x = a.[0]",
			Line: 1,
		},
		{
			Name: "empty_index",
			Src:  "let x = a[]",
			Line: 1,
		},
		{
			Name: "two_indexes",
			Src:  "let x = a[1, 2]",
			Line: 1,
		},
		{
			Name: "unclosed_index",
			Src:  "let x = a[0",
		},
	})
}

func TestPrimary(t *testing.T) {
	// --- primary
	tests.RunExprAst(t, []tests.AstCase{
		{
			Name: "integer",
			Src:  "42",
			Want: "42",
		},
		{
			Name: "zero",
			Src:  "0",
			Want: "0",
		},
		{
			Name: "float",
			Src:  "3.2",
			Want: "3.2",
		},
		{
			Name: "fraction",
			Src:  "0.5",
			Want: "0.5",
		},
		{
			Name: "float_zero_fraction",
			Src:  "1.0",
			Want: "1",
		},
		{
			Name: "float_trailing_zero",
			Src:  "1.50",
			Want: "1.5",
		},
		{
			Name: "string",
			Src:  `"ak"`,
			Want: `"ak"`,
		},
		{
			Name: "string_escapes",
			Src:  `"say \"hi\"\n\ttab \\ back \r"`,
			Want: `"say \"hi\"\n\ttab \\ back \r"`,
		},
		{
			Name: "string_escaped_dollar_brace",
			Src:  `"costs \${price}"`,
			Want: `"costs \${price}"`,
		},
		{
			Name: "string_escaped_dollar_alone",
			Src:  `"cost \$5"`,
			Want: `"cost $5"`,
		},
		{
			Name: "template",
			Src:  `"hello ${user.name}, first is ${arr[0]}"`,
			Want: `"hello ${user.name}, first is ${arr[0]}"`,
		},
		{
			Name: "true",
			Src:  "true",
			Want: "true",
		},
		{
			Name: "false",
			Src:  "false",
			Want: "false",
		},
		{
			Name: "null",
			Src:  "null",
			Want: "null",
		},
		{
			Name: "ident",
			Src:  "name",
			Want: "name",
		},
		{
			Name: "ident_case_sensitive",
			Src:  "Name",
			Want: "Name",
		},
		{
			Name: "title_true_is_literal",
			Src:  "True",
			Want: "true",
		},
		{
			Name: "upper_false_is_literal",
			Src:  "FALSE",
			Want: "false",
		},
		{
			Name: "mixed_null_is_literal",
			Src:  "nUll",
			Want: "null",
		},
	})

	// --- keyword operators are case insensitive
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "upper_and_or",
			Src:  "a AND b OR c",
			Same: "a and b or c",
		},
		{
			Name: "title_not",
			Src:  "Not a == b",
			Same: "not (a == b)",
		},
		{
			Name: "upper_in",
			Src:  "a IN b",
			Same: "a in b",
		},
		{
			Name: "mixed_not_in",
			Src:  "a Not In b",
			Same: "a not in b",
		},
	})

	tests.RunDiffExpr(t, []tests.ParseSameCase{
		{
			Name: "ident_case_kept",
			Src:  "Name",
			Same: "name",
		},
	})

	// --- grouping
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "double_group",
			Src:  "((a))",
			Same: "a",
		},
		{
			Name: "group_expr",
			Src:  "(1 + 2)",
			Same: "1 + 2",
		},
		{
			Name: "group_multi_line",
			Src:  "(1 +\n    2 +\n    3)",
			Same: "1 + 2 + 3",
		},
	})

	// --- primary errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "empty_group",
			Src:  "let x = ()",
			Line: 1,
		},
		{
			Name: "unclosed_group",
			Src:  "let x = (1 + 2",
		},
		{
			Name: "extra_close",
			Src:  "let x = (1 + 2))",
			Line: 1,
		},
		{
			Name: "dangling_operator",
			Src:  "let x = 1 +",
			Line: 1,
		},
		{
			Name: "binary_missing_left",
			Src:  "let x = * 2",
			Line: 1,
		},
		{
			Name: "keyword_operand",
			Src:  "let x = if",
			Line: 1,
		},
		{
			Name: "upper_keyword_operand",
			Src:  "let x = IF",
			Line: 1,
		},
	})
}

func TestArray(t *testing.T) {
	// --- array
	tests.RunExprAst(t, []tests.AstCase{
		{
			Name: "array_numbers",
			Src:  "[1, 2]",
			Want: "[1, 2]",
		},
		{
			Name: "array_empty",
			Src:  "[]",
			Want: "[]",
		},
		{
			Name: "array_mixed",
			Src:  `[1, "a", true, null]`,
			Want: `[1, "a", true, null]`,
		},
		{
			Name: "array_nested",
			Src:  "[[1, 2], [3, 4]]",
			Want: "[[1, 2], [3, 4]]",
		},
		{
			Name: "array_expr_element",
			Src:  "[a + 1]",
			Want: "[a + 1]",
		},
		{
			Name: "array_trailing_comma_dropped",
			Src:  "[\n    1,\n    2,\n]",
			Want: "[1, 2]",
		},
	})

	// --- array multi line and trailing comma
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "array_split_lines",
			Src:  "[\n    1,\n    2,\n]",
			Same: "[1, 2]",
		},
		{
			Name: "array_trailing_comma",
			Src:  "[1, 2,]",
			Same: "[1, 2]",
		},
		{
			Name: "array_empty_split",
			Src:  "[\n]",
			Same: "[]",
		},
		{
			Name: "array_nested_split",
			Src:  "[\n    [1, 2],\n    [3, 4],\n]",
			Same: "[[1, 2], [3, 4]]",
		},
	})

	// --- array errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "array_only_comma",
			Src:  "let x = [,]",
			Line: 1,
		},
		{
			Name: "array_double_comma",
			Src:  "let x = [1,,2]",
			Line: 1,
		},
		{
			Name: "array_missing_comma",
			Src:  "let x = [1 2]",
			Line: 1,
		},
		{
			Name: "array_unclosed",
			Src:  "let x = [1, 2",
		},
		{
			Name: "array_wrong_close",
			Src:  "let x = [1, 2)",
			Line: 1,
		},
	})
}

func TestObject(t *testing.T) {
	// --- object and entry
	tests.RunExprAst(t, []tests.AstCase{
		{
			Name: "object_ident_keys",
			Src:  "{name: name, total: n}",
			Want: "{name: name, total: n}",
		},
		{
			Name: "object_string_key",
			Src:  `{"content-type": "json"}`,
			Want: `{"content-type": "json"}`,
		},
		{
			Name: "object_ident_like_string_key",
			Src:  `{"a": 1}`,
			Want: "{a: 1}",
		},
		{
			Name: "object_numeric_string_key",
			Src:  `{"1": true}`,
			Want: `{"1": true}`,
		},
		{
			Name: "object_keyword_string_key",
			Src:  `{"if": 1}`,
			Want: `{"if": 1}`,
		},
		{
			Name: "object_array_value",
			Src:  "{a: [1, 2]}",
			Want: "{a: [1, 2]}",
		},
		{
			Name: "object_nested",
			Src:  "{a: {b: 1}}",
			Want: "{a: {b: 1}}",
		},
		{
			Name: "object_expr_value",
			Src:  "{a: 1 + 2}",
			Want: "{a: 1 + 2}",
		},
		{
			Name: "object_empty",
			Src:  "{}",
			Want: "{}",
		},
	})

	// --- object multi line and trailing comma
	tests.RunSameExpr(t, []tests.ParseSameCase{
		{
			Name: "object_trailing_comma",
			Src:  "{a: 1, b: 2,}",
			Same: "{a: 1, b: 2}",
		},
		{
			Name: "object_split_lines",
			Src:  "{\n    retries: 3,\n    endpoint: {host: \"localhost\", port: 8080},\n}",
			Same: "{retries: 3, endpoint: {host: \"localhost\", port: 8080}}",
		},
		{
			Name: "object_empty_split",
			Src:  "{\n}",
			Same: "{}",
		},
		{
			Name: "object_ident_key_is_string",
			Src:  "{a: 1}",
			Same: `{"a": 1}`,
		},
	})

	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "object_same_key_other_level",
			Src:  "let x = {a: {a: 1}}",
		},
		{
			Name: "object_numeric_string_keys",
			Src:  `let x = {"1": true, "2": 2, "3": 3.2}`,
		},
		{
			Name: "object_as_call_arg",
			Src:  "f({a: 1})",
		},
		{
			Name: "object_in_array",
			Src:  `let x = [1, {"1": 1, "2": 2}]`,
		},
	})

	// --- duplicate keys
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "duplicate_ident_key",
			Src:  "let x = {a: 1, a: 2}",
			Line: 1,
		},
		{
			Name: "duplicate_ident_and_string_key",
			Src:  `let x = {a: 1, "a": 2}`,
			Line: 1,
		},
		{
			Name: "duplicate_string_key",
			Src:  `let x = {"k": 1, "k": 2}`,
			Line: 1,
		},
		{
			Name: "duplicate_nested_key",
			Src:  "let x = {a: {b: 1, b: 2}}",
			Line: 1,
		},
		{
			Name: "duplicate_key_multi_line",
			Src:  "let x = {\n    a: 1,\n    a: 2,\n}\n",
			Line: 3,
		},
	})

	// --- entry errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "number_key",
			Src:  "let x = {1: 2}",
			Line: 1,
		},
		{
			Name: "interp_key",
			Src:  `let x = {"${k}": 1}`,
			Line: 1,
		},
		{
			Name: "computed_key",
			Src:  "let x = {[a]: 1}",
			Line: 1,
		},
		{
			Name: "member_key",
			Src:  "let x = {a.b: 1}",
			Line: 1,
		},
		{
			Name: "missing_colon",
			Src:  "let x = {a 1}",
			Line: 1,
		},
		{
			Name: "missing_value",
			Src:  "let x = {a: }",
			Line: 1,
		},
		{
			Name: "missing_comma",
			Src:  "let x = {a: 1 b: 2}",
			Line: 1,
		},
		{
			Name: "only_comma",
			Src:  "let x = {,}",
			Line: 1,
		},
		{
			Name: "shorthand_entry",
			Src:  "let x = {a}",
			Line: 1,
		},
		{
			Name: "object_unclosed",
			Src:  "let x = {a: 1",
		},
	})
}

func TestList(t *testing.T) {
	// --- list
	tests.RunSame(t, []tests.ParseSameCase{
		{
			Name: "list_call_trailing_comma",
			Src:  "f(a, b,)",
			Same: "f(a, b)",
		},
		{
			Name: "list_call_split_no_trailing",
			Src:  "f(\n    a,\n    b\n)\n",
			Same: "f(a, b)",
		},
		{
			Name: "list_array_trailing_comma",
			Src:  "let x = [a, b,]",
			Same: "let x = [a, b]",
		},
		{
			Name: "list_single_trailing_comma",
			Src:  "let x = [a,]",
			Same: "let x = [a]",
		},
	})

	tests.RunDiff(t, []tests.ParseSameCase{
		{
			Name: "list_call_arity_kept",
			Src:  "f(a, b)",
			Same: "f(a)",
		},
	})
}

func TestInterpolation(t *testing.T) {
	// --- value_ref in string interpolation
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "interp_ident",
			Src:  `let s = "hi ${name}"`,
		},
		{
			Name: "interp_member",
			Src:  `let s = "hello ${user.name}"`,
		},
		{
			Name: "interp_number_index",
			Src:  `let s = "first is ${arr[0]}"`,
		},
		{
			Name: "interp_string_index",
			Src:  `let s = "${a["k"]}"`,
		},
		{
			Name: "interp_value_ref_index",
			Src:  `let s = "${a[b.c]}"`,
		},
		{
			Name: "interp_nested_value_ref_index",
			Src:  `let s = "${a[b[0]]}"`,
		},
		{
			Name: "interp_two_parts",
			Src:  `let s = "${k}=${v}"`,
		},
		{
			Name: "interp_escaped",
			Src:  `let s = "costs \${price}"`,
		},
		{
			Name: "interp_in_block",
			Src:  "for k, v in obj {\n    let pair = \"${k}=${v}\"\n}\n",
		},
	})

	tests.RunDiff(t, []tests.ParseSameCase{
		{
			Name: "interp_differs_from_escaped",
			Src:  `let s = "${price}"`,
			Same: `let s = "\${price}"`,
		},
	})

	// --- value_ref errors
	tests.RunParseErr(t, []tests.ParseErrCase{
		{
			Name: "interp_binary",
			Src:  `let s = "${1 + 2}"`,
			Line: 1,
		},
		{
			Name: "interp_ident_binary",
			Src:  `let s = "${a + b}"`,
			Line: 1,
		},
		{
			Name: "interp_call",
			Src:  `let s = "${f()}"`,
			Line: 1,
		},
		{
			Name: "interp_neg",
			Src:  `let s = "${-a}"`,
			Line: 1,
		},
		{
			Name: "interp_empty",
			Src:  `let s = "${}"`,
			Line: 1,
		},
		{
			Name: "interp_literal",
			Src:  `let s = "${1}"`,
			Line: 1,
		},
		{
			Name: "interp_expr_index",
			Src:  `let s = "${a[1 + 1]}"`,
			Line: 1,
		},
		{
			Name: "interp_second_line",
			Src:  "let a = 1\nlet s = \"${a + 1}\"\n",
			Line: 2,
		},
	})
}

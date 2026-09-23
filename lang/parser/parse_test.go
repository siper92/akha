package parser_test

import (
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/parser"
	"github.com/siper92/akha/lang/token"
)

func TestParseCalls(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- empty input
		{
			Name:     "empty_source",
			Input:    "",
			Expected: result{Script: script()},
		},
		{
			Name:     "only_newlines",
			Input:    "\n\n\n",
			Expected: result{Script: script()},
		},
		{
			Name:     "only_spaces",
			Input:    "   \t ",
			Expected: result{Script: script()},
		},
		// --- single call
		{
			Name:     "no_args",
			Input:    "Ak.Allow()",
			Expected: result{Script: script(callStmt("Ak", "Allow", 1, 1, nil, nil))},
		},
		{
			Name:     "no_args_newline_inside",
			Input:    "Ak.Allow(\n)",
			Expected: result{Script: script(callStmt("Ak", "Allow", 1, 1, nil, nil))},
		},
		{
			Name:     "no_trailing_newline",
			Input:    `Ak.Log("a")`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
		},
		{
			Name:     "trailing_newline",
			Input:    "Ak.Log(\"a\")\n",
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
		},
		// --- many calls
		{
			Name:  "three_calls",
			Input: "Ak.Allow(FS...)\nFS.ReadFile(\"x\")\nAk.Exit(\"e\", code=1)",
			Expected: result{Script: script(
				callStmt("Ak", "Allow", 1, 1, args(spread("FS", 1, 10)), nil),
				callStmt("FS", "ReadFile", 2, 1, args(str("x", 2, 13)), nil),
				callStmt("Ak", "Exit", 3, 1, args(str("e", 3, 9)), kwargs(kw("code", num("1", 3, 19), 3, 14))),
			)},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseArgs(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- positional literals
		{
			Name:     "string",
			Input:    `Ak.Log("hello")`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(str("hello", 1, 8)), nil))},
		},
		{
			Name:     "string_escapes_decoded",
			Input:    `Ak.Log("a\tb\"c\\d\n")`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(str("a\tb\"c\\d\n", 1, 8)), nil))},
		},
		{
			Name:     "unicode_counts_runes",
			Input:    `Ak.Log("é", 1)`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(str("é", 1, 8), num("1", 1, 13)), nil))},
		},
		{
			Name:     "int",
			Input:    `Ak.Log(42)`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(num("42", 1, 8)), nil))},
		},
		{
			Name:     "bools",
			Input:    `Ak.Log(true, false)`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(boolean(true, 1, 8), boolean(false, 1, 14)), nil))},
		},
		{
			Name:     "mixed_literals",
			Input:    `Ak.Debug("dbg", 1, "two")`,
			Expected: result{Script: script(callStmt("Ak", "Debug", 1, 1, args(str("dbg", 1, 10), num("1", 1, 17), str("two", 1, 20)), nil))},
		},
		// --- variable references
		{
			Name:     "ident_positional",
			Input:    `Ak.Log(yes)`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(ident("yes", 1, 8)), nil))},
		},
		{
			Name:     "ident_kwarg_value",
			Input:    `Ak.Log(x=yes)`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, nil, kwargs(kw("x", ident("yes", 1, 10), 1, 8))))},
		},
		{
			Name:     "ident_then_expr_arg",
			Input:    `Ak.Log(x + 1)`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(binary(token.PLUS, ident("x", 1, 8), num("1", 1, 12), 1, 10)), nil))},
		},
		// --- keyword args
		{
			Name:     "kwarg_string",
			Input:    `Ak.Setup(log="a")`,
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 1, 14), 1, 10))))},
		},
		{
			Name:     "kwarg_bool",
			Input:    `Ak.Setup(debug=false)`,
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(kw("debug", boolean(false, 1, 16), 1, 10))))},
		},
		{
			Name:     "kwarg_spaces_around_assign",
			Input:    `Ak.Setup(log = "a")`,
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 1, 16), 1, 10))))},
		},
		{
			Name:     "kwarg_value_on_next_line",
			Input:    "Ak.Setup(log=\n\"a\")",
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 2, 1), 1, 10))))},
		},
		{
			Name:     "kwarg_expression_value",
			Input:    `Ak.Exit("x", code=1 + n)`,
			Expected: result{Script: script(callStmt("Ak", "Exit", 1, 1, args(str("x", 1, 9)), kwargs(kw("code", binary(token.PLUS, num("1", 1, 19), ident("n", 1, 23), 1, 21), 1, 14))))},
		},
		{
			Name:  "two_kwargs",
			Input: `Ak.Setup(log="a", debug="b")`,
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(
				kw("log", str("a", 1, 14), 1, 10),
				kw("debug", str("b", 1, 25), 1, 19),
			)))},
		},
		{
			Name:     "positional_then_kwarg",
			Input:    `Ak.Exit("x", code=3)`,
			Expected: result{Script: script(callStmt("Ak", "Exit", 1, 1, args(str("x", 1, 9)), kwargs(kw("code", num("3", 1, 19), 1, 14))))},
		},
		// --- nested calls
		{
			Name:     "call_as_positional",
			Input:    `Ak.Log(FS.ReadFile("f"))`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(call("FS", "ReadFile", 1, 8, args(str("f", 1, 20)), nil)), nil))},
		},
		{
			Name:     "call_as_kwarg_value",
			Input:    `Ak.Exit(code=Ak.Len(x))`,
			Expected: result{Script: script(callStmt("Ak", "Exit", 1, 1, nil, kwargs(kw("code", call("Ak", "Len", 1, 14, args(ident("x", 1, 21)), nil), 1, 9))))},
		},
		// --- spread
		{
			Name:     "one_spread",
			Input:    `Ak.Allow(FS...)`,
			Expected: result{Script: script(callStmt("Ak", "Allow", 1, 1, args(spread("FS", 1, 10)), nil))},
		},
		{
			Name:     "many_spreads",
			Input:    `Ak.Allow(FS..., HTTP...)`,
			Expected: result{Script: script(callStmt("Ak", "Allow", 1, 1, args(spread("FS", 1, 10), spread("HTTP", 1, 17)), nil))},
		},
		// --- trailing comma
		{
			Name:     "trailing_comma_spread",
			Input:    `Ak.Allow(FS...,)`,
			Expected: result{Script: script(callStmt("Ak", "Allow", 1, 1, args(spread("FS", 1, 10)), nil))},
		},
		{
			Name:     "trailing_comma_kwarg",
			Input:    `Ak.Setup(log="a",)`,
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 1, 14), 1, 10))))},
		},
		{
			Name:     "trailing_comma_positional",
			Input:    `Ak.Log("a",)`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseLet(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- declarations
		{
			Name:     "string_value",
			Input:    `let x = "a"`,
			Expected: result{Script: script(let("x", str("a", 1, 9), 1, 1))},
		},
		{
			Name:     "int_value",
			Input:    `let x = 1`,
			Expected: result{Script: script(let("x", num("1", 1, 9), 1, 1))},
		},
		{
			Name:     "bool_value",
			Input:    `let x = true`,
			Expected: result{Script: script(let("x", boolean(true, 1, 9), 1, 1))},
		},
		{
			Name:     "ident_value",
			Input:    `let x = y`,
			Expected: result{Script: script(let("x", ident("y", 1, 9), 1, 1))},
		},
		{
			Name:     "call_value",
			Input:    `let x = FS.ReadFile("f")`,
			Expected: result{Script: script(let("x", call("FS", "ReadFile", 1, 9, args(str("f", 1, 21)), nil), 1, 1))},
		},
		{
			Name:     "expression_value",
			Input:    `let x = 1 + 2`,
			Expected: result{Script: script(let("x", binary(token.PLUS, num("1", 1, 9), num("2", 1, 13), 1, 11), 1, 1))},
		},
		{
			Name:     "no_spaces",
			Input:    `let x=1`,
			Expected: result{Script: script(let("x", num("1", 1, 7), 1, 1))},
		},
		{
			Name:     "indented",
			Input:    "  let x = 1",
			Expected: result{Script: script(let("x", num("1", 1, 11), 1, 3))},
		},
		{
			Name:     "underscore_name",
			Input:    `let _x1 = 1`,
			Expected: result{Script: script(let("_x1", num("1", 1, 11), 1, 1))},
		},
		{
			Name:  "let_then_use",
			Input: "let x = 1\nAk.Log(x)",
			Expected: result{Script: script(
				let("x", num("1", 1, 9), 1, 1),
				callStmt("Ak", "Log", 2, 1, args(ident("x", 2, 8)), nil),
			)},
		},
		// --- errors
		{
			Name:     "missing_name",
			Input:    `let = 1`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 5, "expected variable name, got assign"))},
		},
		{
			Name:     "keyword_name",
			Input:    `let if = 1`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 5, "expected variable name, got if"))},
		},
		{
			Name:     "missing_assign",
			Input:    `let x 1`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 7, "expected '=', got int"))},
		},
		{
			Name:     "missing_value",
			Input:    `let x =`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected expression, got eof"))},
		},
		{
			Name:     "value_on_next_line",
			Input:    "let x =\n1",
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected expression, got newline"), perr(2, 1, "expression is not a statement"))},
		},
		{
			Name:     "let_alone",
			Input:    `let`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 4, "expected variable name, got eof"))},
		},
		{
			Name:  "let_error_then_next_line_parses",
			Input: "let = 1\nlet y = 2",
			Expected: result{
				Script: script(let("y", num("2", 2, 9), 2, 1)),
				Errs:   perrs(perr(1, 5, "expected variable name, got assign")),
			},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseAssign(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- assignments
		{
			Name:     "literal",
			Input:    `x = 1`,
			Expected: result{Script: script(assign("x", num("1", 1, 5), 1, 1))},
		},
		{
			Name:     "self_reference",
			Input:    `x = x + 1`,
			Expected: result{Script: script(assign("x", binary(token.PLUS, ident("x", 1, 5), num("1", 1, 9), 1, 7), 1, 1))},
		},
		{
			Name:     "call_value",
			Input:    `x = Ak.Str(1)`,
			Expected: result{Script: script(assign("x", call("Ak", "Str", 1, 5, args(num("1", 1, 12)), nil), 1, 1))},
		},
		{
			Name:     "no_spaces",
			Input:    `x=1`,
			Expected: result{Script: script(assign("x", num("1", 1, 3), 1, 1))},
		},
		// --- errors
		{
			Name:     "missing_value",
			Input:    `x =`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 4, "expected expression, got eof"))},
		},
		{
			Name:     "double_assign",
			Input:    `x = = 1`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 5, "expected expression, got assign"))},
		},
		{
			Name:     "eq_is_not_assign",
			Input:    `x == 1`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 3, "expected '.' or '=', got eq"))},
		},
		{
			Name:     "ident_alone",
			Input:    `x`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 2, "expected '.' or '=', got eof"))},
		},
		{
			Name:     "ident_then_newline",
			Input:    "x\n",
			Expected: result{Script: script(), Errs: perrs(perr(1, 2, "expected '.' or '=', got newline"))},
		},
		{
			Name:     "ident_then_expression",
			Input:    `x + 1`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 3, "expected '.' or '=', got plus"))},
		},
		{
			Name:     "trailing_tokens",
			Input:    `x = 1 2`,
			Expected: result{Script: script(assign("x", num("1", 1, 5), 1, 1)), Errs: perrs(perr(1, 7, "expected newline after statement, got int"))},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseExprs(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- precedence
		{
			Name:     "mul_binds_tighter_than_add",
			Input:    `let x = 1 + 2 * 3`,
			Expected: result{Script: script(let("x", binary(token.PLUS, num("1", 1, 9), binary(token.STAR, num("2", 1, 13), num("3", 1, 17), 1, 15), 1, 11), 1, 1))},
		},
		{
			Name:  "add_is_left_assoc",
			Input: `let x = 1 - 2 - 3`,
			Expected: result{Script: script(let("x", binary(
				token.MINUS,
				binary(token.MINUS, num("1", 1, 9), num("2", 1, 13), 1, 11), num("3", 1, 17), 1, 15), 1, 1))},
		},
		{
			Name:     "mul_is_left_assoc",
			Input:    `let x = 8 / 2 % 3`,
			Expected: result{Script: script(let("x", binary(token.PERCENT, binary(token.SLASH, num("8", 1, 9), num("2", 1, 13), 1, 11), num("3", 1, 17), 1, 15), 1, 1))},
		},
		{
			Name:     "cmp_below_add",
			Input:    `let x = 1 + 2 == 3`,
			Expected: result{Script: script(let("x", binary(token.EQ, binary(token.PLUS, num("1", 1, 9), num("2", 1, 13), 1, 11), num("3", 1, 18), 1, 15), 1, 1))},
		},
		{
			Name:     "and_below_cmp",
			Input:    `let x = a < b and c`,
			Expected: result{Script: script(let("x", binary(token.AND, binary(token.LT, ident("a", 1, 9), ident("b", 1, 13), 1, 11), ident("c", 1, 19), 1, 15), 1, 1))},
		},
		{
			Name:     "or_below_and",
			Input:    `let x = a or b and c`,
			Expected: result{Script: script(let("x", binary(token.OR, ident("a", 1, 9), binary(token.AND, ident("b", 1, 14), ident("c", 1, 20), 1, 16), 1, 11), 1, 1))},
		},
		{
			Name:     "not_above_and",
			Input:    `let x = not a and b`,
			Expected: result{Script: script(let("x", binary(token.AND, unary(token.NOT, ident("a", 1, 13), 1, 9), ident("b", 1, 19), 1, 15), 1, 1))},
		},
		{
			Name:     "not_below_cmp",
			Input:    `let x = not a == b`,
			Expected: result{Script: script(let("x", unary(token.NOT, binary(token.EQ, ident("a", 1, 13), ident("b", 1, 18), 1, 15), 1, 9), 1, 1))},
		},
		{
			Name:     "unary_minus_above_mul",
			Input:    `let x = -a * b`,
			Expected: result{Script: script(let("x", binary(token.STAR, unary(token.MINUS, ident("a", 1, 10), 1, 9), ident("b", 1, 14), 1, 12), 1, 1))},
		},
		{
			Name:     "double_negation",
			Input:    `let x = --1`,
			Expected: result{Script: script(let("x", unary(token.MINUS, unary(token.MINUS, num("1", 1, 11), 1, 10), 1, 9), 1, 1))},
		},
		{
			Name:     "double_not",
			Input:    `let x = not not a`,
			Expected: result{Script: script(let("x", unary(token.NOT, unary(token.NOT, ident("a", 1, 17), 1, 13), 1, 9), 1, 1))},
		},
		// --- grouping
		{
			Name:     "parens_override",
			Input:    `let x = (1 + 2) * 3`,
			Expected: result{Script: script(let("x", binary(token.STAR, binary(token.PLUS, num("1", 1, 10), num("2", 1, 14), 1, 12), num("3", 1, 19), 1, 17), 1, 1))},
		},
		{
			Name:     "nested_parens",
			Input:    `let x = ((a))`,
			Expected: result{Script: script(let("x", ident("a", 1, 11), 1, 1))},
		},
		{
			Name:     "minus_grouped",
			Input:    `let x = -(a + b)`,
			Expected: result{Script: script(let("x", unary(token.MINUS, binary(token.PLUS, ident("a", 1, 11), ident("b", 1, 15), 1, 13), 1, 9), 1, 1))},
		},
		// --- comparisons
		{
			Name:     "neq",
			Input:    `let x = a != b`,
			Expected: result{Script: script(let("x", binary(token.NEQ, ident("a", 1, 9), ident("b", 1, 14), 1, 11), 1, 1))},
		},
		{
			Name:     "lte",
			Input:    `let x = a <= b`,
			Expected: result{Script: script(let("x", binary(token.LTE, ident("a", 1, 9), ident("b", 1, 14), 1, 11), 1, 1))},
		},
		{
			Name:     "gt",
			Input:    `let x = a > b`,
			Expected: result{Script: script(let("x", binary(token.GT, ident("a", 1, 9), ident("b", 1, 13), 1, 11), 1, 1))},
		},
		{
			Name:     "gte",
			Input:    `let x = a >= b`,
			Expected: result{Script: script(let("x", binary(token.GTE, ident("a", 1, 9), ident("b", 1, 14), 1, 11), 1, 1))},
		},
		// --- calls inside expressions
		{
			Name:     "call_in_binary",
			Input:    `let x = Ak.Len(s) + 1`,
			Expected: result{Script: script(let("x", binary(token.PLUS, call("Ak", "Len", 1, 9, args(ident("s", 1, 16)), nil), num("1", 1, 21), 1, 19), 1, 1))},
		},
		{
			Name:     "call_with_expression_args",
			Input:    `let x = Ak.Str(a + 1, not b)`,
			Expected: result{Script: script(let("x", call("Ak", "Str", 1, 9, args(binary(token.PLUS, ident("a", 1, 16), num("1", 1, 20), 1, 18), unary(token.NOT, ident("b", 1, 27), 1, 23)), nil), 1, 1))},
		},
		{
			Name:     "call_in_call",
			Input:    `let x = Ak.Str(Ak.Len(s))`,
			Expected: result{Script: script(let("x", call("Ak", "Str", 1, 9, args(call("Ak", "Len", 1, 16, args(ident("s", 1, 23)), nil)), nil), 1, 1))},
		},
		// --- errors
		{
			Name:     "chained_comparison",
			Input:    `let x = 1 < 2 < 3`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 15, "expected newline after statement, got lt"))},
		},
		{
			Name:     "missing_right_operand",
			Input:    `let x = 1 +`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 12, "expected expression, got eof"))},
		},
		{
			Name:     "missing_right_operand_newline",
			Input:    "let x = 1 +\n2",
			Expected: result{Script: script(), Errs: perrs(perr(1, 12, "expected expression, got newline"), perr(2, 1, "expression is not a statement"))},
		},
		{
			Name:     "unclosed_paren",
			Input:    `let x = (1 + 2`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 15, "expected ')', got eof"))},
		},
		{
			Name:     "empty_parens",
			Input:    `let x = ()`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 10, "expected expression, got rparen"))},
		},
		{
			Name:     "two_operators",
			Input:    `let x = 1 * * 2`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 13, "expected expression, got star"))},
		},
		{
			Name:     "keyword_as_operand",
			Input:    `let x = if`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 9, "expected expression, got if"))},
		},
		{
			Name:     "spread_in_expression",
			Input:    "let x = FS...",
			Expected: result{Script: script(), Errs: perrs(perr(1, 11, "expected newline after statement, got ellipsis"))},
		},
	}

	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseStmtErrors(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- expressions cannot stand alone
		{
			Name:     "string_statement",
			Input:    `"a"`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expression is not a statement"))},
		},
		{
			Name:     "int_statement",
			Input:    `1 + 2`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expression is not a statement"))},
		},
		{
			Name:     "bool_statement",
			Input:    `true`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expression is not a statement"))},
		},
		{
			Name:     "paren_statement",
			Input:    `(x)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expression is not a statement"))},
		},
		{
			Name:     "minus_statement",
			Input:    `-x`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expression is not a statement"))},
		},
		{
			Name:     "not_statement",
			Input:    `not x`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expression is not a statement"))},
		},
		// --- keywords without a statement yet
		{
			Name:     "if_not_implemented",
			Input:    `if x {`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "if is not implemented"))},
		},
		{
			Name:     "for_not_implemented",
			Input:    `for x in y {`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "for is not implemented"))},
		},
		{
			Name:     "while_not_implemented",
			Input:    `while x {`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "while is not implemented"))},
		},
		{
			Name:     "break_not_implemented",
			Input:    `break`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "break is not implemented"))},
		},
		{
			Name:     "continue_not_implemented",
			Input:    `continue`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "continue is not implemented"))},
		},
		// --- other keywords and punctuation
		{
			Name:     "else_statement",
			Input:    `else`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expected statement, got else"))},
		},
		{
			Name:     "in_statement",
			Input:    `in`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expected statement, got in"))},
		},
		{
			Name:     "stray_rbrace",
			Input:    `}`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expected statement, got rbrace"))},
		},
		{
			Name:     "stray_lbrace",
			Input:    `{`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expected statement, got lbrace"))},
		},
	}

	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseLayout(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- newlines inside parens
		{
			Name:  "newlines_everywhere",
			Input: "Ak.Setup(\n\n log = \"a\"\n ,\n debug=\"b\"\n)",
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(
				kw("log", str("a", 3, 8), 3, 2),
				kw("debug", str("b", 5, 8), 5, 2),
			)))},
		},
		{
			Name:     "multi_line_positional",
			Input:    "Ak.Debug(\n  \"a\",\n  1,\n)",
			Expected: result{Script: script(callStmt("Ak", "Debug", 1, 1, args(str("a", 2, 3), num("1", 3, 3)), nil))},
		},
		{
			Name:     "multi_line_call_value",
			Input:    "let x = Ak.Str(\n  1,\n)",
			Expected: result{Script: script(let("x", call("Ak", "Str", 1, 9, args(num("1", 2, 3)), nil), 1, 1))},
		},
		// --- blank lines
		{
			Name:     "leading_and_trailing_blank_lines",
			Input:    "\n\nAk.Log(\"a\")\n\n",
			Expected: result{Script: script(callStmt("Ak", "Log", 3, 1, args(str("a", 3, 8)), nil))},
		},
		{
			Name:  "blank_lines_between_calls",
			Input: "Ak.Log(\"a\")\n\n\nAk.Log(\"b\")",
			Expected: result{Script: script(
				callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil),
				callStmt("Ak", "Log", 4, 1, args(str("b", 4, 8)), nil),
			)},
		},
		// --- whitespace
		{
			Name:     "tabs_count_one_col",
			Input:    "\tAk.Log(\t\"a\")",
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 2, args(str("a", 1, 10)), nil))},
		},
		{
			Name:  "crlf_line_endings",
			Input: "Ak.Log(\"a\")\r\nAk.Log(\"b\")\r\n",
			Expected: result{Script: script(
				callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil),
				callStmt("Ak", "Log", 2, 1, args(str("b", 2, 8)), nil),
			)},
		},
	}

	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseComments(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- comments give the same ast in both lexer modes
		{
			Name:     "only_comment",
			Input:    "// only a comment",
			Expected: result{Script: script()},
		},
		{
			Name:     "trailing_comment",
			Input:    `Ak.Log("a") // trailing`,
			Expected: result{Script: script(callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
		},
		{
			Name:     "head_and_tail_comments",
			Input:    "// head\nAk.Log(\"a\")\n// tail",
			Expected: result{Script: script(callStmt("Ak", "Log", 2, 1, args(str("a", 2, 8)), nil))},
		},
		{
			Name:  "comments_inside_parens",
			Input: "Ak.Setup(\n  log=\"a\", // c\n  debug=\"b\" // d\n)",
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(
				kw("log", str("a", 2, 7), 2, 3),
				kw("debug", str("b", 3, 9), 3, 3),
			)))},
		},
		{
			Name:     "comment_after_assign",
			Input:    "Ak.Setup(log= // c\n\"a\")",
			Expected: result{Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 2, 1), 1, 10))))},
		},
		{
			Name:     "comment_after_let",
			Input:    "let x = 1 // c",
			Expected: result{Script: script(let("x", num("1", 1, 9), 1, 1))},
		},
	}
	t.Run("comments_dropped", func(t *testing.T) {
		tu.Run(tu.New(t), cases, parseWith(), nil)
	})
	t.Run("comments_kept", func(t *testing.T) {
		tu.Run(tu.New(t), cases, parseWith(lexer.WithComments()), nil)
	})
}

func TestParseRuleErrors(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- spread outside Ak.Allow keeps the spread
		{
			Name:  "spread_in_ak_log",
			Input: `Ak.Log(FS...)`,
			Expected: result{
				Script: script(callStmt("Ak", "Log", 1, 1, args(spread("FS", 1, 8)), nil)),
				Errs:   perrs(perr(1, 8, "spread only allowed in Ak.Allow")),
			},
		},
		{
			Name:  "spread_in_fs_module",
			Input: `FS.ReadFile(FS...)`,
			Expected: result{
				Script: script(callStmt("FS", "ReadFile", 1, 1, args(spread("FS", 1, 13)), nil)),
				Errs:   perrs(perr(1, 13, "spread only allowed in Ak.Allow")),
			},
		},
		{
			Name:  "spread_in_nested_allow",
			Input: `let x = Ak.Allow(FS...)`,
			Expected: result{
				Script: script(let("x", call("Ak", "Allow", 1, 9, args(spread("FS", 1, 18)), nil), 1, 1)),
			},
		},
		// --- positional after keyword keeps both
		{
			Name:  "literal_after_kwarg",
			Input: `Ak.Log("a", code=1, "b")`,
			Expected: result{
				Script: script(callStmt("Ak", "Log", 1, 1,
					args(str("a", 1, 8), str("b", 1, 21)),
					kwargs(kw("code", num("1", 1, 18), 1, 13)),
				)),
				Errs: perrs(perr(1, 21, "positional argument after keyword argument")),
			},
		},
		{
			Name:  "expression_after_kwarg_reports_operator_pos",
			Input: `Ak.Log(code=1, a + b)`,
			Expected: result{
				Script: script(callStmt("Ak", "Log", 1, 1,
					args(binary(token.PLUS, ident("a", 1, 16), ident("b", 1, 20), 1, 18)),
					kwargs(kw("code", num("1", 1, 13), 1, 8)),
				)),
				Errs: perrs(perr(1, 18, "positional argument after keyword argument")),
			},
		},
		{
			Name:  "spread_after_kwarg",
			Input: `Ak.Allow(x=1, FS...)`,
			Expected: result{
				Script: script(callStmt("Ak", "Allow", 1, 1,
					args(spread("FS", 1, 15)),
					kwargs(kw("x", num("1", 1, 12), 1, 10)),
				)),
				Errs: perrs(perr(1, 15, "positional argument after keyword argument")),
			},
		},
		{
			Name:  "spread_after_kwarg_outside_allow",
			Input: `FS.ReadFile(x=1, FS...)`,
			Expected: result{
				Script: script(callStmt("FS", "ReadFile", 1, 1,
					args(spread("FS", 1, 18)),
					kwargs(kw("x", num("1", 1, 15), 1, 13)),
				)),
				Errs: perrs(
					perr(1, 18, "spread only allowed in Ak.Allow"),
					perr(1, 18, "positional argument after keyword argument"),
				),
			},
		},
		// --- duplicate kwarg keeps every kwarg
		{
			Name:  "duplicate_once",
			Input: `Ak.Setup(log="a", log="b")`,
			Expected: result{
				Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(
					kw("log", str("a", 1, 14), 1, 10),
					kw("log", str("b", 1, 23), 1, 19),
				))),
				Errs: perrs(perr(1, 19, "duplicate argument log")),
			},
		},
		{
			Name:  "duplicate_twice",
			Input: `Ak.Setup(a=1, a=2, a=3)`,
			Expected: result{
				Script: script(callStmt("Ak", "Setup", 1, 1, nil, kwargs(
					kw("a", num("1", 1, 12), 1, 10),
					kw("a", num("2", 1, 17), 1, 15),
					kw("a", num("3", 1, 22), 1, 20),
				))),
				Errs: perrs(
					perr(1, 15, "duplicate argument a"),
					perr(1, 20, "duplicate argument a"),
				),
			},
		},
		// --- rule error does not drop later statements
		{
			Name:  "rule_error_then_call",
			Input: "Ak.Log(FS...)\nAk.Log(\"b\")",
			Expected: result{
				Script: script(
					callStmt("Ak", "Log", 1, 1, args(spread("FS", 1, 8)), nil),
					callStmt("Ak", "Log", 2, 1, args(str("b", 2, 8)), nil),
				),
				Errs: perrs(perr(1, 8, "spread only allowed in Ak.Allow")),
			},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseSyntaxErrors(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- statement head
		{
			Name:     "illegal_start",
			Input:    `#`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, `expected statement, got illegal "#"`))},
		},
		{
			Name:     "assign_start",
			Input:    `=`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expected statement, got assign"))},
		},
		{
			Name:     "missing_dot",
			Input:    `Ak`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 3, "expected '.' or '=', got eof"))},
		},
		{
			Name:     "lparen_after_module",
			Input:    `Ak("a")`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 3, "expected '.' or '=', got lparen"))},
		},
		{
			Name:     "int_function_name",
			Input:    `Ak.1()`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 4, "expected function name, got int"))},
		},
		{
			Name:     "keyword_function_name",
			Input:    `Ak.if()`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 4, "expected function name, got if"))},
		},
		{
			Name:     "double_dot",
			Input:    `FS..ReadFile("x")`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 4, "expected function name, got dot"))},
		},
		{
			Name:     "missing_lparen_eof",
			Input:    `Ak.Log`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 7, "expected '(', got eof"))},
		},
		{
			Name:     "missing_lparen_string",
			Input:    `Ak.Log "a"`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected '(', got string"))},
		},
		// --- arguments
		{
			Name:     "leading_comma",
			Input:    `Ak.Log(,)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected expression, got comma"))},
		},
		{
			Name:     "double_comma",
			Input:    `Ak.Log("a",,"b")`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 12, "expected expression, got comma"))},
		},
		{
			Name:     "assign_without_name",
			Input:    `Ak.Log(=1)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected expression, got assign"))},
		},
		{
			Name:     "illegal_arg",
			Input:    `Ak.Log("a", #)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 13, `expected expression, got illegal "#"`))},
		},
		{
			Name:     "unterminated_string_arg",
			Input:    `Ak.Log("abc`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, `expected expression, got illegal "\"abc"`))},
		},
		{
			Name:     "missing_kwarg_value",
			Input:    `Ak.Log(x=)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 10, "expected expression, got rparen"))},
		},
		{
			Name:     "unterminated_string_kwarg_value",
			Input:    `Ak.Log(x="ab`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 10, `expected expression, got illegal "\"ab"`))},
		},
		{
			Name:     "missing_comma",
			Input:    `Ak.Log("a" "b")`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 12, "expected ')' or ',', got string"))},
		},
		{
			Name:     "ident_then_ident_drops_call",
			Input:    `Ak.Log(yes x)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 12, "expected ')' or ',', got ident"))},
		},
		{
			Name:     "brace_in_args",
			Input:    `Ak.Log({)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected expression, got lbrace"))},
		},
		// --- unclosed paren
		{
			Name:     "eof_after_lparen",
			Input:    `Ak.Log(`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "unexpected end of file, expected ')'"))},
		},
		{
			Name:     "eof_after_arg",
			Input:    `Ak.Log("a"`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 11, "unexpected end of file, expected ')'"))},
		},
		{
			Name:     "eof_after_comma_and_newline",
			Input:    "Ak.Setup(\n    log=\"a\",\n",
			Expected: result{Script: script(), Errs: perrs(perr(3, 1, "unexpected end of file, expected ')'"))},
		},
		// --- after statement
		{
			Name:  "second_call_same_line",
			Input: `Ak.Log("a") Ak.Log("b")`,
			Expected: result{
				Script: script(callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil)),
				Errs:   perrs(perr(1, 13, "expected newline after statement, got ident")),
			},
		},
		{
			Name:  "illegal_after_call",
			Input: `Ak.Log("a")#`,
			Expected: result{
				Script: script(callStmt("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil)),
				Errs:   perrs(perr(1, 12, `expected newline after statement, got illegal "#"`)),
			},
		},
		{
			Name:  "rbrace_after_let",
			Input: `let x = 1 }`,
			Expected: result{
				Script: script(let("x", num("1", 1, 9), 1, 1)),
				Errs:   perrs(perr(1, 11, "expected newline after statement, got rbrace")),
			},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseRecovery(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- sync to next line and keep parsing
		{
			Name:  "mixed_errors_across_lines",
			Input: "Ak.Log(\"ok\")\nAk.Log \"bad\"\nFS..ReadFile(\"x\")\nAk.Setup(log=\"a\", \"pos\", log=\"b\")\nAk.Exit(\"done\")",
			Expected: result{
				Script: script(
					callStmt("Ak", "Log", 1, 1, args(str("ok", 1, 8)), nil),
					callStmt("Ak", "Setup", 4, 1,
						args(str("pos", 4, 19)),
						kwargs(
							kw("log", str("a", 4, 14), 4, 10),
							kw("log", str("b", 4, 30), 4, 26),
						),
					),
					callStmt("Ak", "Exit", 5, 1, args(str("done", 5, 9)), nil),
				),
				Errs: perrs(
					perr(2, 8, "expected '(', got string"),
					perr(3, 4, "expected function name, got dot"),
					perr(4, 19, "positional argument after keyword argument"),
					perr(4, 26, "duplicate argument log"),
				),
			},
		},
		{
			Name:  "sync_after_missing_comma",
			Input: "Ak.Log(1 2)\nAk.Log(\"x\")",
			Expected: result{
				Script: script(callStmt("Ak", "Log", 2, 1, args(str("x", 2, 8)), nil)),
				Errs:   perrs(perr(1, 10, "expected ')' or ',', got int")),
			},
		},
		{
			Name:  "lparen_on_next_line",
			Input: "Ak.Log\n(\"a\")",
			Expected: result{Script: script(), Errs: perrs(
				perr(1, 7, "expected '(', got newline"),
				perr(2, 1, "expression is not a statement"),
			)},
		},
		{
			Name:  "errors_on_every_line_but_last",
			Input: "Ak\nFS.\nAk.Log(\"ok\")",
			Expected: result{
				Script: script(callStmt("Ak", "Log", 3, 1, args(str("ok", 3, 8)), nil)),
				Errs: perrs(
					perr(1, 3, "expected '.' or '=', got newline"),
					perr(2, 4, "expected function name, got newline"),
				),
			},
		},
		{
			Name:  "bad_let_between_good_statements",
			Input: "let a = 1\nlet b = +\na = 2",
			Expected: result{
				Script: script(
					let("a", num("1", 1, 9), 1, 1),
					assign("a", num("2", 3, 5), 3, 1),
				),
				Errs: perrs(perr(2, 9, "expected expression, got plus")),
			},
		},
		{
			Name:  "expression_statement_then_call",
			Input: "1 + 1\nAk.Log(\"x\")",
			Expected: result{
				Script: script(callStmt("Ak", "Log", 2, 1, args(str("x", 2, 8)), nil)),
				Errs:   perrs(perr(1, 1, "expression is not a statement")),
			},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseNilError(t *testing.T) {
	cases := []tu.Case[string, bool]{
		// --- clean parse returns an untyped nil error
		{
			Name:     "empty_source",
			Input:    "",
			Expected: true,
		},
		{
			Name:     "clean_script",
			Input:    "Ak.Allow(FS...)\nFS.ReadFile(\"x\")",
			Expected: true,
		},
		{
			Name:     "clean_let",
			Input:    "let x = 1\nx = x + 1",
			Expected: true,
		},
		// --- any error returns a non nil error
		{
			Name:     "syntax_error",
			Input:    `Ak`,
			Expected: false,
		},
		{
			Name:     "rule_error",
			Input:    `Ak.Log(FS...)`,
			Expected: false,
		},
	}
	tu.Run(tu.New(t), cases, func(src string) (bool, error) {
		_, err := parser.New(lexer.New(src)).Parse()
		return err == nil, nil
	}, nil)
}

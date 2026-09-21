package parser_test

import (
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/parser"
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
			Expected: result{Script: script(call("Ak", "Allow", 1, 1, nil, nil))},
		},
		{
			Name:     "no_args_newline_inside",
			Input:    "Ak.Allow(\n)",
			Expected: result{Script: script(call("Ak", "Allow", 1, 1, nil, nil))},
		},
		{
			Name:     "no_trailing_newline",
			Input:    `Ak.Log("a")`,
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
		},
		{
			Name:     "trailing_newline",
			Input:    "Ak.Log(\"a\")\n",
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
		},
		// --- many calls
		{
			Name:  "three_calls",
			Input: "Ak.Allow(FS...)\nFS.ReadFile(\"x\")\nAk.Exit(\"e\", code=1)",
			Expected: result{Script: script(
				call("Ak", "Allow", 1, 1, args(spread("FS", 1, 10)), nil),
				call("FS", "ReadFile", 2, 1, args(str("x", 2, 13)), nil),
				call("Ak", "Exit", 3, 1, args(str("e", 3, 9)), kwargs(kw("code", num("1", 3, 19), 3, 14))),
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
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(str("hello", 1, 8)), nil))},
		},
		{
			Name:     "string_escapes_decoded",
			Input:    `Ak.Log("a\tb\"c\\d\n")`,
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(str("a\tb\"c\\d\n", 1, 8)), nil))},
		},
		{
			Name:     "unicode_counts_runes",
			Input:    `Ak.Log("é", 1)`,
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(str("é", 1, 8), num("1", 1, 13)), nil))},
		},
		{
			Name:     "int",
			Input:    `Ak.Log(42)`,
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(num("42", 1, 8)), nil))},
		},
		{
			Name:     "bools",
			Input:    `Ak.Log(true, false)`,
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(ident("true", 1, 8), ident("false", 1, 14)), nil))},
		},
		{
			Name:     "mixed_literals",
			Input:    `Ak.Debug("dbg", 1, "two")`,
			Expected: result{Script: script(call("Ak", "Debug", 1, 1, args(str("dbg", 1, 10), num("1", 1, 17), str("two", 1, 20)), nil))},
		},
		// --- keyword args
		{
			Name:     "kwarg_string",
			Input:    `Ak.Setup(log="a")`,
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 1, 14), 1, 10))))},
		},
		{
			Name:     "kwarg_bool",
			Input:    `Ak.Setup(debug=false)`,
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(kw("debug", ident("false", 1, 16), 1, 10))))},
		},
		{
			Name:     "kwarg_spaces_around_assign",
			Input:    `Ak.Setup(log = "a")`,
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 1, 16), 1, 10))))},
		},
		{
			Name:     "kwarg_value_on_next_line",
			Input:    "Ak.Setup(log=\n\"a\")",
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 2, 1), 1, 10))))},
		},
		{
			Name:  "two_kwargs",
			Input: `Ak.Setup(log="a", debug="b")`,
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(
				kw("log", str("a", 1, 14), 1, 10),
				kw("debug", str("b", 1, 25), 1, 19),
			)))},
		},
		{
			Name:     "positional_then_kwarg",
			Input:    `Ak.Exit("x", code=3)`,
			Expected: result{Script: script(call("Ak", "Exit", 1, 1, args(str("x", 1, 9)), kwargs(kw("code", num("3", 1, 19), 1, 14))))},
		},
		// --- spread
		{
			Name:     "one_spread",
			Input:    `Ak.Allow(FS...)`,
			Expected: result{Script: script(call("Ak", "Allow", 1, 1, args(spread("FS", 1, 10)), nil))},
		},
		{
			Name:     "many_spreads",
			Input:    `Ak.Allow(FS..., HTTP...)`,
			Expected: result{Script: script(call("Ak", "Allow", 1, 1, args(spread("FS", 1, 10), spread("HTTP", 1, 17)), nil))},
		},
		// --- trailing comma
		{
			Name:     "trailing_comma_spread",
			Input:    `Ak.Allow(FS...,)`,
			Expected: result{Script: script(call("Ak", "Allow", 1, 1, args(spread("FS", 1, 10)), nil))},
		},
		{
			Name:     "trailing_comma_kwarg",
			Input:    `Ak.Setup(log="a",)`,
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 1, 14), 1, 10))))},
		},
		{
			Name:     "trailing_comma_positional",
			Input:    `Ak.Log("a",)`,
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
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
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(
				kw("log", str("a", 3, 8), 3, 2),
				kw("debug", str("b", 5, 8), 5, 2),
			)))},
		},
		{
			Name:     "multi_line_positional",
			Input:    "Ak.Debug(\n  \"a\",\n  1,\n)",
			Expected: result{Script: script(call("Ak", "Debug", 1, 1, args(str("a", 2, 3), num("1", 3, 3)), nil))},
		},
		// --- blank lines
		{
			Name:     "leading_and_trailing_blank_lines",
			Input:    "\n\nAk.Log(\"a\")\n\n",
			Expected: result{Script: script(call("Ak", "Log", 3, 1, args(str("a", 3, 8)), nil))},
		},
		{
			Name:  "blank_lines_between_calls",
			Input: "Ak.Log(\"a\")\n\n\nAk.Log(\"b\")",
			Expected: result{Script: script(
				call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil),
				call("Ak", "Log", 4, 1, args(str("b", 4, 8)), nil),
			)},
		},
		// --- whitespace
		{
			Name:     "tabs_count_one_col",
			Input:    "\tAk.Log(\t\"a\")",
			Expected: result{Script: script(call("Ak", "Log", 1, 2, args(str("a", 1, 10)), nil))},
		},
		{
			Name:  "crlf_line_endings",
			Input: "Ak.Log(\"a\")\r\nAk.Log(\"b\")\r\n",
			Expected: result{Script: script(
				call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil),
				call("Ak", "Log", 2, 1, args(str("b", 2, 8)), nil),
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
			Expected: result{Script: script(call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil))},
		},
		{
			Name:     "head_and_tail_comments",
			Input:    "// head\nAk.Log(\"a\")\n// tail",
			Expected: result{Script: script(call("Ak", "Log", 2, 1, args(str("a", 2, 8)), nil))},
		},
		{
			Name:  "comments_inside_parens",
			Input: "Ak.Setup(\n  log=\"a\", // c\n  debug=\"b\" // d\n)",
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(
				kw("log", str("a", 2, 7), 2, 3),
				kw("debug", str("b", 3, 9), 3, 3),
			)))},
		},
		{
			Name:     "comment_after_assign",
			Input:    "Ak.Setup(log= // c\n\"a\")",
			Expected: result{Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(kw("log", str("a", 2, 1), 1, 10))))},
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
		// --- unexpected identifier keeps the literal
		{
			Name:  "ident_positional",
			Input: `Ak.Log(yes)`,
			Expected: result{
				Script: script(call("Ak", "Log", 1, 1, args(ident("yes", 1, 8)), nil)),
				Errs:   perrs(perr(1, 8, "unexpected identifier")),
			},
		},
		{
			Name:  "ident_kwarg_value",
			Input: `Ak.Log(x=yes)`,
			Expected: result{
				Script: script(call("Ak", "Log", 1, 1, nil, kwargs(kw("x", ident("yes", 1, 10), 1, 8)))),
				Errs:   perrs(perr(1, 10, "unexpected identifier")),
			},
		},
		// --- spread outside Ak.Allow keeps the spread
		{
			Name:  "spread_in_ak_log",
			Input: `Ak.Log(FS...)`,
			Expected: result{
				Script: script(call("Ak", "Log", 1, 1, args(spread("FS", 1, 8)), nil)),
				Errs:   perrs(perr(1, 8, "spread only allowed in Ak.Allow")),
			},
		},
		{
			Name:  "spread_in_fs_module",
			Input: `FS.ReadFile(FS...)`,
			Expected: result{
				Script: script(call("FS", "ReadFile", 1, 1, args(spread("FS", 1, 13)), nil)),
				Errs:   perrs(perr(1, 13, "spread only allowed in Ak.Allow")),
			},
		},
		// --- positional after keyword keeps both
		{
			Name:  "literal_after_kwarg",
			Input: `Ak.Log("a", code=1, "b")`,
			Expected: result{
				Script: script(call("Ak", "Log", 1, 1,
					args(str("a", 1, 8), str("b", 1, 21)),
					kwargs(kw("code", num("1", 1, 18), 1, 13)),
				)),
				Errs: perrs(perr(1, 21, "positional argument after keyword argument")),
			},
		},
		{
			Name:  "spread_after_kwarg",
			Input: `Ak.Allow(x=1, FS...)`,
			Expected: result{
				Script: script(call("Ak", "Allow", 1, 1,
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
				Script: script(call("FS", "ReadFile", 1, 1,
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
				Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(
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
				Script: script(call("Ak", "Setup", 1, 1, nil, kwargs(
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
		// --- rule error does not drop later calls
		{
			Name:  "rule_error_then_call",
			Input: "Ak.Log(yes)\nAk.Log(\"b\")",
			Expected: result{
				Script: script(
					call("Ak", "Log", 1, 1, args(ident("yes", 1, 8)), nil),
					call("Ak", "Log", 2, 1, args(str("b", 2, 8)), nil),
				),
				Errs: perrs(perr(1, 8, "unexpected identifier")),
			},
		},
	}
	tu.Run(tu.New(t), cases, parseWith(), nil)
}

func TestParseSyntaxErrors(t *testing.T) {
	cases := []tu.Case[string, result]{
		// --- call head
		{
			Name:     "illegal_start",
			Input:    `-`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, `expected module name, got illegal "-"`))},
		},
		{
			Name:     "assign_start",
			Input:    `=`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 1, "expected module name, got assign"))},
		},
		{
			Name:     "missing_dot",
			Input:    `Ak`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 3, "expected '.', got eof"))},
		},
		{
			Name:     "lparen_after_module",
			Input:    `Ak("a")`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 3, "expected '.', got lparen"))},
		},
		{
			Name:     "int_function_name",
			Input:    `Ak.1()`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 4, "expected function name, got int"))},
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
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected argument, got comma"))},
		},
		{
			Name:     "double_comma",
			Input:    `Ak.Log("a",,"b")`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 12, "expected argument, got comma"))},
		},
		{
			Name:     "assign_without_name",
			Input:    `Ak.Log(=1)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, "expected argument, got assign"))},
		},
		{
			Name:     "illegal_arg",
			Input:    `Ak.Log("a", #)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 13, `expected argument, got illegal "#"`))},
		},
		{
			Name:     "unterminated_string_arg",
			Input:    `Ak.Log("abc`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 8, `expected argument, got illegal "\"abc"`))},
		},
		{
			Name:     "missing_kwarg_value",
			Input:    `Ak.Log(x=)`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 10, "expected value, got rparen"))},
		},
		{
			Name:     "unterminated_string_kwarg_value",
			Input:    `Ak.Log(x="ab`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 10, `expected value, got illegal "\"ab"`))},
		},
		{
			Name:     "missing_comma",
			Input:    `Ak.Log("a" "b")`,
			Expected: result{Script: script(), Errs: perrs(perr(1, 12, "expected ')' or ',', got string"))},
		},
		{
			Name:  "rule_error_then_syntax_error_drops_call",
			Input: `Ak.Log(yes x)`,
			Expected: result{Script: script(), Errs: perrs(
				perr(1, 8, "unexpected identifier"),
				perr(1, 12, "expected ')' or ',', got ident"),
			)},
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
		// --- after call
		{
			Name:  "second_call_same_line",
			Input: `Ak.Log("a") Ak.Log("b")`,
			Expected: result{
				Script: script(call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil)),
				Errs:   perrs(perr(1, 13, "expected newline after call, got ident")),
			},
		},
		{
			Name:  "illegal_after_call",
			Input: `Ak.Log("a")#`,
			Expected: result{
				Script: script(call("Ak", "Log", 1, 1, args(str("a", 1, 8)), nil)),
				Errs:   perrs(perr(1, 12, `expected newline after call, got illegal "#"`)),
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
					call("Ak", "Log", 1, 1, args(str("ok", 1, 8)), nil),
					call("Ak", "Setup", 4, 1,
						args(str("pos", 4, 19)),
						kwargs(
							kw("log", str("a", 4, 14), 4, 10),
							kw("log", str("b", 4, 30), 4, 26),
						),
					),
					call("Ak", "Exit", 5, 1, args(str("done", 5, 9)), nil),
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
				Script: script(call("Ak", "Log", 2, 1, args(str("x", 2, 8)), nil)),
				Errs:   perrs(perr(1, 10, "expected ')' or ',', got int")),
			},
		},
		{
			Name:  "lparen_on_next_line",
			Input: "Ak.Log\n(\"a\")",
			Expected: result{Script: script(), Errs: perrs(
				perr(1, 7, "expected '(', got newline"),
				perr(2, 1, "expected module name, got lparen"),
			)},
		},
		{
			Name:  "errors_on_every_line_but_last",
			Input: "Ak\nFS.\nAk.Log(\"ok\")",
			Expected: result{
				Script: script(call("Ak", "Log", 3, 1, args(str("ok", 3, 8)), nil)),
				Errs: perrs(
					perr(1, 3, "expected '.', got newline"),
					perr(2, 4, "expected function name, got newline"),
				),
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
		// --- any error returns a non nil error
		{
			Name:     "syntax_error",
			Input:    `Ak`,
			Expected: false,
		},
		{
			Name:     "rule_error",
			Input:    `Ak.Log(yes)`,
			Expected: false,
		},
	}
	tu.Run(tu.New(t), cases, func(src string) (bool, error) {
		_, err := parser.New(lexer.New(src)).Parse()
		return err == nil, nil
	}, nil)
}

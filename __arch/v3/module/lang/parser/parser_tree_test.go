package parser_test

import (
	"strings"
	"testing"

	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/tests_utils"
)

func TestExprParseTrees(t *testing.T) {
	// --- precedence and associativity, every nested operation grouped
	tests_utils.RunExprTree(t, []tests_utils.TreeCase{
		{
			Name: "add_mul",
			Src:  "1 + 2 * 3",
			Want: "1 + (2 * 3)",
		},
		{
			Name: "sub_left_assoc",
			Src:  "a - b - c",
			Want: "(a - b) - c",
		},
		{
			Name: "or_and",
			Src:  "a or b and c or d",
			Want: "(a or (b and c)) or d",
		},
		{
			Name: "and_or",
			Src:  "a and b or c and d",
			Want: "(a and b) or (c and d)",
		},
		{
			Name: "not_over_eq",
			Src:  "not a == b",
			Want: "not (a == b)",
		},
		{
			Name: "not_not",
			Src:  "not not a",
			Want: "not (not a)",
		},
		{
			Name: "bang_under_and",
			Src:  "!a and b",
			Want: "(not a) and b",
		},
		{
			Name: "cmp_under_in",
			Src:  "a == b in c",
			Want: "(a == b) in c",
		},
		{
			Name: "add_under_not_in",
			Src:  "a + b not in c",
			Want: "(a + b) not in c",
		},
		{
			Name: "neg_over_mul",
			Src:  "-1 * 2",
			Want: "(-1) * 2",
		},
		{
			Name: "neg_under_index",
			Src:  "-x[i]",
			Want: "-(x[i])",
		},
		{
			Name: "neg_neg",
			Src:  "- -a",
			Want: "-(-a)",
		},
		{
			Name: "sub_neg",
			Src:  "a - -b",
			Want: "a - (-b)",
		},
	})

	// --- keywords ignore case
	tests_utils.RunExprTree(t, []tests_utils.TreeCase{
		{
			Name: "not_in",
			Src:  "a not in b",
			Want: "a not in b",
		},
		{
			Name: "mixed_not_in",
			Src:  "a Not In b",
			Want: "a not in b",
		},
		{
			Name: "upper_and",
			Src:  "a AND b",
			Want: "a and b",
		},
		{
			Name: "title_true",
			Src:  "True",
			Want: "true",
		},
		{
			Name: "upper_null",
			Src:  "NULL",
			Want: "null",
		},
	})

	// --- postfix and operands
	tests_utils.RunExprTree(t, []tests_utils.TreeCase{
		{
			Name: "index_member_call",
			Src:  "x[i].f(42)",
			Want: "x[i].f(42)",
		},
		{
			Name: "member_call",
			Src:  "x.f()",
			Want: "x.f()",
		},
		{
			Name: "member_on_binary",
			Src:  "(a + b).c",
			Want: "(a + b).c",
		},
		{
			Name: "index_on_neg",
			Src:  "(-a)[0]",
			Want: "(-a)[0]",
		},
		{
			Name: "call_args_grouped",
			Src:  "f(a + b * c, -d)",
			Want: "f(a + (b * c), -d)",
		},
		{
			Name: "call_split_lines",
			Src:  "f(\n\n1,\n\n)",
			Want: "f(1)",
		},
		{
			Name: "group_dropped",
			Src:  "(4)",
			Want: "4",
		},
		{
			Name: "array_empty",
			Src:  "[]",
			Want: "[]",
		},
		{
			Name: "array_trailing_comma",
			Src:  "[1,]",
			Want: "[1]",
		},
		{
			Name: "array_nested",
			Src:  "[1, [2, 3]]",
			Want: "[1, [2, 3]]",
		},
		{
			Name: "array_elem_grouped",
			Src:  "[a + b * c]",
			Want: "[a + (b * c)]",
		},
		{
			Name: "object_empty",
			Src:  "{}",
			Want: "{}",
		},
		{
			Name: "object_trailing_comma",
			Src:  `{"a": 1,}`,
			Want: "{a: 1}",
		},
		{
			Name: "object_value_grouped",
			Src:  `{"b-c": 1 - 2 - 3}`,
			Want: `{"b-c": (1 - 2) - 3}`,
		},
		{
			Name: "string_concat",
			Src:  `"foo" + "bar"`,
			Want: `"foo" + "bar"`,
		},
		{
			Name: "template",
			Src:  `"hi ${user.name}"`,
			Want: `"hi ${user.name}"`,
		},
	})

	// --- errors, position stripped
	tests_utils.RunExprTree(t, []tests_utils.TreeCase{
		{
			Name: "cmp_chain",
			Src:  "a < b < c",
			Want: "comparisons do not chain",
		},
		{
			Name: "in_not_in_chain",
			Src:  "a in b not in c",
			Want: "in does not chain",
		},
		{
			Name: "slice",
			Src:  "a[1:2]",
			Want: "slices are not supported in v1",
		},
		{
			Name: "dangling_operator",
			Src:  "1 +",
			Want: "expected an expression, got end of file",
		},
		{
			Name: "bang_is_not_not_in",
			Src:  "a ! in b",
			Want: `unexpected "!" after expression`,
		},
		{
			Name: "keyword_member",
			Src:  "a.if",
			Want: `expected a member name after ., got "if"`,
		},
		{
			Name: "shorthand_entry",
			Src:  "{a}",
			Want: `expected :, got "}"`,
		},
		{
			Name: "tuple",
			Src:  "(1, 2)",
			Want: `expected ), got ","`,
		},
		{
			Name: "missing_comma",
			Src:  "f(a b)",
			Want: "expected ), got identifier b",
		},
	})
}

func TestStmtParseTrees(t *testing.T) {
	// --- statements
	tests_utils.RunTree(t, []tests_utils.TreeCase{
		{
			Name: "let",
			Src:  "let a = 1",
			Want: "let a = 1",
		},
		{
			Name: "upper_let",
			Src:  "LET a = 1",
			Want: "let a = 1",
		},
		{
			Name: "let_grouped_value",
			Src:  "let x = a * b + c",
			Want: "let x = (a * b) + c",
		},
		{
			Name: "var_bare",
			Src:  "var b",
			Want: "var b",
		},
		{
			Name: "var_init",
			Src:  "var b = [1]",
			Want: "var b = [1]",
		},
		{
			Name: "assign",
			Src:  "x = n + 1",
			Want: "x = n + 1",
		},
		{
			Name: "assign_nested_target",
			Src:  "a.b[0] = 1",
			Want: "a.b[0] = 1",
		},
		{
			Name: "call",
			Src:  "f(1, 2)",
			Want: "f(1, 2)",
		},
		{
			Name: "return_bare",
			Src:  "return",
			Want: "return",
		},
		{
			Name: "exit_value",
			Src:  "exit 1",
			Want: "exit 1",
		},
		{
			Name: "return_grouped",
			Src:  "return a or b and c",
			Want: "return a or (b and c)",
		},
		{
			Name: "template",
			Src:  `let s = "hi ${name}"`,
			Want: `let s = "hi ${name}"`,
		},
		{
			Name: "object",
			Src:  `let o = {a: 1, "b-c": [1, 2]}`,
			Want: `let o = {a: 1, "b-c": [1, 2]}`,
		},
	})

	// --- blocks
	tests_utils.RunTree(t, []tests_utils.TreeCase{
		{
			Name: "if_else",
			Src:  "if a {\n    b()\n} else {\n}\n",
			Want: "if a {\n    b()\n} else {\n}",
		},
		{
			Name: "else_if",
			Src:  "if a {\n} else if b {\n}\n",
			Want: "if a {\n} else if b {\n}",
		},
		{
			Name: "if_header_grouped",
			Src:  "if a and b or c {\n}\n",
			Want: "if (a and b) or c {\n}",
		},
		{
			Name: "if_header_object",
			Src:  "if ({a: 1}) {\n}\n",
			Want: "if ({a: 1}) {\n}",
		},
		{
			Name: "block_stmts",
			Src:  "if a {\n  let b = 1\n\t\tvar c\n}\n",
			Want: "if a {\n    let b = 1\n    var c\n}",
		},
		{
			Name: "for_in",
			Src:  "for x in a {\n}\n",
			Want: "for x in a {\n}",
		},
		{
			Name: "for_var_pair",
			Src:  "for var i, v in xs {\n    continue\n}\n",
			Want: "for var i, v in xs {\n    continue\n}",
		},
		{
			Name: "for_range",
			Src:  "for i range [0..3] {\n    break\n}\n",
			Want: "for i range [0..3] {\n    break\n}",
		},
		{
			Name: "for_range_grouped_bounds",
			Src:  "for i range [a * 2 + 1..n - 1 - m] {\n}\n",
			Want: "for i range [(a * 2) + 1..(n - 1) - m] {\n}",
		},
	})

	// --- file, blank lines and comments
	tests_utils.RunTree(t, []tests_utils.TreeCase{
		{
			Name: "two_lines",
			Src:  "let a = 1\n\n// c\nvar b\n",
			Want: "let a = 1\nvar b",
		},
		{
			Name: "multi_line_value",
			Src:  "x = (1 +\n2)\n",
			Want: "x = 1 + 2",
		},
		{
			Name: "empty",
			Src:  "\n// only a comment\n",
			Want: "",
		},
	})

	// --- errors, position stripped
	tests_utils.RunTree(t, []tests_utils.TreeCase{
		{
			Name: "grouped_target",
			Src:  "(a) = 1",
			Want: "cannot assign to this expression",
		},
		{
			Name: "grouped_member_target",
			Src:  "(a).b = 1",
			Want: "cannot assign to this expression",
		},
		{
			Name: "unused_value",
			Src:  "x",
			Want: "unused value, only a call can stand alone",
		},
		{
			Name: "break_outside_loop",
			Src:  "break",
			Want: "break outside of a loop",
		},
		{
			Name: "upper_break_outside_loop",
			Src:  "BREAK",
			Want: "break outside of a loop",
		},
		{
			Name: "two_statements_one_line",
			Src:  "let a = 1 let b = 2",
			Want: `expected end of line, got "let"`,
		},
		{
			Name: "block_on_one_line",
			Src:  "if a {}",
			Want: `expected a new line after {, got "}"`,
		},
	})
}

func TestParseErrors(t *testing.T) {
	// --- golden chunks with expected errors
	tests_utils.RunChunks(t, "testdata/errors.ak")
}

func TestNestingLimit(t *testing.T) {
	// --- deep nesting is rejected instead of exhausting the stack
	tests_utils.RunNesting(t, []tests_utils.ParseOkCase{
		{
			Name: "parens_rejected_in_lexer",
			Src:  "let x = " + strings.Repeat("(", lexer.MaxDepth+1),
		},
		{
			Name: "parens_rejected_in_parser",
			Src:  "let x = " + strings.Repeat("(", 900) + "1" + strings.Repeat(")", 900),
		},
		{
			Name: "not_chain",
			Src:  "let x = " + strings.Repeat("not ", 5000) + "a",
		},
		{
			Name: "bang_chain",
			Src:  "let x = " + strings.Repeat("!", 5000) + "a",
		},
		{
			Name: "neg_chain",
			Src:  "let x = " + strings.Repeat("- ", 5000) + "a",
		},
		{
			Name: "nested_if",
			Src:  strings.Repeat("if a {\n", 600),
		},
		{
			Name: "nested_for",
			Src:  strings.Repeat("for x in a {\n", 600),
		},
	})

	// --- reasonable nesting still parses
	tests_utils.RunOK(t, []tests_utils.ParseOkCase{
		{
			Name: "parens_50",
			Src:  "let x = " + strings.Repeat("(", 50) + "1" + strings.Repeat(")", 50),
		},
		{
			Name: "arrays_50",
			Src:  "let x = " + strings.Repeat("[", 50) + strings.Repeat("]", 50),
		},
		{
			Name: "nested_if_50",
			Src:  strings.Repeat("if a {\n", 50) + strings.Repeat("}\n", 50),
		},
	})
}

func TestGrammar(t *testing.T) {
	// --- grammar.txt is closed: every production is defined once and used
	tests_utils.RunGrammarClosed(t, tests_utils.GrammarFile)

	// --- every production has a documented parse function
	tests_utils.RunGrammarCoverage(t, tests_utils.GrammarFile, "parser.go")
}

func TestParserCallGraphCycles(t *testing.T) {
	// --- every recursive cycle of parser methods passes through p.enter
	tests_utils.RunCallGraphCycles(t, "parser.go", "parser", "p", "enter")
}

package parser

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/tests_utils"
)

func TestParserFullScripts(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		{
			Name:     "simple canonical sample",
			Input:    Sample,
			Expected: SampleCanonical,
		},
		{
			Name:     "simple canonical sample with CRLF",
			Input:    SampleCRLF,
			Expected: SampleCRLFCanonical,
		},
		{
			Name:     "akha spec def file",
			Input:    "file://testdata/spec_def.ak",
			Expected: "file://testdata/spec_def.canonical.ak",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			validateParserOutput(t, testCase)
		})
	}
}

func TestParserCanonical(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- empty input
		{
			Name:     "empty_source",
			Input:    "",
			Expected: "",
		},
		{
			Name:     "comment_only",
			Input:    "// c",
			Expected: "",
		},
		{
			Name:     "blank_lines_only",
			Input:    "\n\n\n",
			Expected: "",
		},
		// --- layout is dropped
		{
			Name:     "comments_and_blank_lines",
			Input:    "// top\n\nlet a = 1 // x\n\n\n// end\n",
			Expected: "let a = 1",
		},
		{
			Name:     "indented_top_level",
			Input:    "    let a = 1\n\t\tf()",
			Expected: "let a = 1\nf()",
		},
		{
			Name:     "bom",
			Input:    "\xEF\xBB\xBFlet a = 1",
			Expected: "let a = 1",
		},
		{
			Name:     "comment_only_block",
			Input:    "if a {\n    // nothing\n}",
			Expected: "if a {\n}",
		},
		// --- keywords are case insensitive
		{
			Name:     "upper_case_keywords",
			Input:    "LET a = TRUE\nVar b = Null\nIF a {\n    b = 1\n} ELSE {\n    RETURN\n}",
			Expected: "let a = true\nvar b = null\nif a {\n    b = 1\n} else {\n    return\n}",
		},
		{
			Name:     "symbolic_logic_prints_words",
			Input:    "let x = a && b || !c",
			Expected: "let x = a and b or not c",
		},
		// --- precedence and grouping
		{
			Name:     "left_assoc_parens_dropped",
			Input:    "let x = (a - b) - c",
			Expected: "let x = a - b - c",
		},
		{
			Name:     "right_group_kept",
			Input:    "let x = a - (b - c)",
			Expected: "let x = a - (b - c)",
		},
		{
			Name:     "mixed_add_sub",
			Input:    "let x = a - b + c",
			Expected: "let x = a - b + c",
		},
		{
			Name:     "or_inside_and",
			Input:    "let x = (a or b) and c",
			Expected: "let x = (a or b) and c",
		},
		{
			Name:     "not_of_group",
			Input:    "let x = not (a and b)",
			Expected: "let x = not (a and b)",
		},
		{
			Name:     "grouped_not",
			Input:    "let x = (not a) and b",
			Expected: "let x = not a and b",
		},
		{
			Name:     "not_binds_below_in",
			Input:    "let x = not a in b",
			Expected: "let x = not a in b",
		},
		{
			Name:     "grouped_comparison",
			Input:    "let x = (a < b) == c",
			Expected: "let x = (a < b) == c",
		},
		{
			Name:     "grouped_in",
			Input:    "let x = (a in b) in c",
			Expected: "let x = (a in b) in c",
		},
		{
			Name:     "comparison_inside_in",
			Input:    "let x = a == b in c",
			Expected: "let x = a == b in c",
		},
		{
			Name:     "not_in",
			Input:    "let x = a not in b",
			Expected: "let x = a not in b",
		},
		{
			Name:     "double_minus",
			Input:    "let x = -(-a)",
			Expected: "let x = - -a",
		},
		{
			Name:     "minus_of_group",
			Input:    "let x = -(a + b)",
			Expected: "let x = -(a + b)",
		},
		{
			Name:     "minus_binds_tighter_than_mul",
			Input:    "let x = 1 + -2 * 3",
			Expected: "let x = 1 + -2 * 3",
		},
		{
			Name:     "member_of_unary",
			Input:    "let x = (-a).b",
			Expected: "let x = (-a).b",
		},
		{
			Name:     "index_of_group",
			Input:    "let x = (a + b)[0]",
			Expected: "let x = (a + b)[0]",
		},
		{
			Name:     "redundant_parens",
			Input:    "let x = ((a))",
			Expected: "let x = a",
		},
		// --- numbers are normalised
		{
			Name:     "numbers",
			Input:    "let a = 1.50\nlet b = 0.0\nlet c = 10.25\nlet d = 100",
			Expected: "let a = 1.5\nlet b = 0\nlet c = 10.25\nlet d = 100",
		},
		// --- strings and templates
		{
			Name:     "string_escapes",
			Input:    `let s = "a\tb\r\n"`,
			Expected: `let s = "a\tb\r\n"`,
		},
		{
			Name:     "string_unicode",
			Input:    `let s = "héllo ✓"`,
			Expected: `let s = "héllo ✓"`,
		},
		{
			Name:     "plain_dollars",
			Input:    `let s = "$5 and $"`,
			Expected: `let s = "$5 and $"`,
		},
		{
			Name:     "escaped_and_real_interp",
			Input:    `let s = "\${x} ${y}"`,
			Expected: `let s = "\${x} ${y}"`,
		},
		{
			Name:     "template_only_expr",
			Input:    `let s = "${a}"`,
			Expected: `let s = "${a}"`,
		},
		{
			Name:     "template_spaces_trimmed",
			Input:    `let s = "${ a.b }"`,
			Expected: `let s = "${a.b}"`,
		},
		{
			Name:     "template_string_index",
			Input:    `let s = "${a["k"]}"`,
			Expected: `let s = "${a["k"]}"`,
		},
		{
			Name:     "template_name_index",
			Input:    `let s = "${a[b][0]}"`,
			Expected: `let s = "${a[b][0]}"`,
		},
		{
			Name:     "template_quotes_around_expr",
			Input:    `let s = "say \"${a}\""`,
			Expected: `let s = "say \"${a}\""`,
		},
		// --- object keys
		{
			Name:     "string_keys_become_names",
			Input:    `let o = {"a": 1, "a b": 2, "if": 3, "1x": 4, _: 5}`,
			Expected: `let o = {a: 1, "a b": 2, "if": 3, "1x": 4, _: 5}`,
		},
		{
			Name:     "keys_stay_quoted",
			Input:    `let o = {"a\"b": 1, "If": 2, "": 3}`,
			Expected: `let o = {"a\"b": 1, "If": 2, "": 3}`,
		},
		// --- object literals in headers
		{
			Name:     "for_over_grouped_object",
			Input:    "for x in ({a: 1}) {\n}",
			Expected: "for x in ({a: 1}) {\n}",
		},
		{
			Name:     "if_member_of_grouped_object",
			Input:    "if ({a: 1}).a {\n}",
			Expected: "if ({a: 1}.a) {\n}",
		},
		{
			Name:     "object_after_header_start",
			Input:    "if a in {a: 1} {\n}",
			Expected: "if a in {a: 1} {\n}",
		},
		// --- multi line expressions
		{
			Name:     "multi_line_call",
			Input:    "f(\n    1,\n    2,\n)",
			Expected: "f(1, 2)",
		},
		{
			Name:     "multi_line_empty_array",
			Input:    "let x = [\n]",
			Expected: "let x = []",
		},
		{
			Name:     "multi_line_group",
			Input:    "let x = (\n    1\n)",
			Expected: "let x = 1",
		},
		{
			Name:     "not_in_across_lines",
			Input:    "let x = (a\n    not in b)",
			Expected: "let x = a not in b",
		},
		{
			Name:     "nested_multi_line_literals",
			Input:    "let o = {\n    a: [\n        1,\n    ],\n}",
			Expected: "let o = {a: [1]}",
		},
		// --- postfix chains
		{
			Name:     "postfix_chain",
			Input:    "let x = a.b(1)[0].c",
			Expected: "let x = a.b(1)[0].c",
		},
		{
			Name:     "call_of_call",
			Input:    "f()()",
			Expected: "f()()",
		},
		{
			Name:     "grouped_callee",
			Input:    "(f)(1)",
			Expected: "f(1)",
		},
		{
			Name:     "nested_calls",
			Input:    "f(g(h(1)), [])",
			Expected: "f(g(h(1)), [])",
		},
		// --- statements
		{
			Name:     "exit_and_return",
			Input:    "return f()\nexit 1\nexit",
			Expected: "return f()\nexit 1\nexit",
		},
		{
			Name:     "nested_targets",
			Input:    "a.b[0] = 1\na[i].c = 2",
			Expected: "a.b[0] = 1\na[i].c = 2",
		},
		{
			Name:     "discard_call",
			Input:    "let _ = f()",
			Expected: "let _ = f()",
		},
		{
			Name:     "discard_loop_value",
			Input:    "for _ in a {\n}",
			Expected: "for _ in a {\n}",
		},
		{
			Name:     "else_if_chain_empty_blocks",
			Input:    "if a {\n} else if b {\n} else {\n}",
			Expected: "if a {\n} else if b {\n} else {\n}",
		},
		{
			Name:     "nested_loops",
			Input:    "for i range [0..n + 1] {\nfor var k, v in o {\nif v {\nbreak\n}\n}\ncontinue\n}",
			Expected: "for i range [0..n + 1] {\n    for var k, v in o {\n        if v {\n            break\n        }\n    }\n    continue\n}",
		},
		{
			Name:     "range_with_float_bounds",
			Input:    "for i range [1.5..2] {\n}",
			Expected: "for i range [1.5..2] {\n}",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateParserOutput(t, c)
		})
	}
}

func TestParserLoopExit(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- return and exit leave the script from any loop depth
		{
			Name:     "exit_in_for_in",
			Input:    "for x in a {\nif x {\nexit x\n}\n}",
			Expected: "for x in a {\n    if x {\n        exit x\n    }\n}",
		},
		{
			Name:     "return_in_nested_range",
			Input:    "for i range [0..n] {\nfor var k, v in o {\nreturn v\n}\n}",
			Expected: "for i range [0..n] {\n    for var k, v in o {\n        return v\n    }\n}",
		},
		// --- break and continue stay loop only
		{
			Name:  "break_after_loop",
			Input: "for x in a {\n}\nbreak",
			Err:   errors.New("error[loop-control]: break outside of a loop"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateParserOutput(t, c)
		})
	}
}

func TestParserMemberAccess(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- dot is the object member operator
		{
			Name:     "member_chain_on_name",
			Input:    "let x = input.user.name",
			Expected: "let x = input.user.name",
		},
		// @todo: if value provides checking during parsing check - and trow an error id not allowed
		// 	- example is an object if it has a value in the script trow an error if member object exist
		{
			Name:     "member_on_array_literal",
			Input:    `let x = [{b: "1""}].b`,
			Expected: "let x = input.user.name",
		},
		{
			Name:  "member_on_string_literal",
			Input: `let x = "a".b`,
			Err:   errors.New("error[member-kind]: member access works only on objects"),
		},
		{
			Name:  "member_on_array_literal",
			Input: "let x = [1].b",
			Err:   errors.New("error[member-kind]: member access works only on objects"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateParserOutput(t, c)
		})
	}
}

func TestParserErrors(t *testing.T) {
	cases, err := extractErrors()
	if err != nil {
		t.Fatalf("Failed to load errors: %v", err)
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			validateParserOutput(t, testCase)
		})
	}
}

func TestParserErrorKinds(t *testing.T) {
	cases := []tests_utils.Case[string, error]{
		{
			Name:     "lex_error",
			Input:    "let a = 1;",
			Expected: lexer.ErrLex,
		},
		{
			Name:     "parse_error",
			Input:    "let a = ",
			Expected: lexer.ErrParse,
		},
		{
			Name:     "lex_error_in_template",
			Input:    `let s = "${}"`,
			Expected: lexer.ErrLex,
		},
		{
			Name:     "parse_error_in_template",
			Input:    `let s = "${f()}"`,
			Expected: lexer.ErrParse,
		},
		{
			Name:     "lex_error_inside_interp",
			Input:    `let s = "${a["\q"]}"`,
			Expected: lexer.ErrLex,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := New("", c.Input).Parse()
			if !errors.Is(err, c.Expected) {
				t.Errorf("got %v, want %v", err, c.Expected)
			}
		})
	}
}

func TestParserLimits(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- within limits
		{
			Name:  "moderate_parens",
			Input: "let x = " + strings.Repeat("(", 50) + "a" + strings.Repeat(")", 50),
		},
		{
			Name:  "moderate_nested_ifs",
			Input: nestedIfs(50),
		},
		{
			Name:  "long_flat_expression",
			Input: "let x = " + strings.Repeat("a + ", 2000) + "a",
		},
		{
			Name:  "many_sequential_blocks",
			Input: strings.Repeat("if a {\n}\n", 800),
		},
		// --- over the limit
		{
			Name:  "deep_parens",
			Input: "let x = " + strings.Repeat("(", 200) + "a" + strings.Repeat(")", 200),
			Err:   errors.New("error[nesting]: excessive nesting"),
		},
		{
			Name:  "deep_minus",
			Input: "let x = " + strings.Repeat("-", 2000) + "a",
			Err:   errors.New("error[nesting]: excessive nesting"),
		},
		{
			Name:  "deep_not",
			Input: "let x = " + strings.Repeat("not ", 1000) + "a",
			Err:   errors.New("error[nesting]: excessive nesting"),
		},
		{
			Name:  "deep_ifs",
			Input: nestedIfs(600),
			Err:   errors.New("error[nesting]: excessive nesting"),
		},
		{
			Name:  "number_out_of_range",
			Input: "let x = 1" + strings.Repeat("0", 400),
			Err:   errors.New("1:9: error[invalid-number]: invalid number 1000"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := New("", c.Input).Parse()
			if c.Err == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil || !strings.Contains(err.Error(), c.Err.Error()) {
				t.Fatalf("expected error: %v, got: %v", c.Err, err)
			}
		})
	}
}

func TestParserStatementLines(t *testing.T) {
	cases := []tests_utils.Case[string, []int]{
		{
			Name:     "blank_lines_between",
			Input:    "let a = 1\n\nif a {\n    f()\n}\nreturn",
			Expected: []int{1, 3, 4, 6},
		},
		{
			Name:     "multi_line_literal",
			Input:    "let x = [\n    1,\n]\nf()",
			Expected: []int{1, 4},
		},
		{
			Name:     "leading_comment",
			Input:    "// c\n\n    var v",
			Expected: []int{3},
		},
		{
			Name:     "crlf_and_bom",
			Input:    "\xEF\xBB\xBFlet a = 1\r\nlet b = 2",
			Expected: []int{1, 2},
		},
		{
			Name:     "nested_blocks",
			Input:    "if a {\n    f()\n} else if b {\n\n    g()\n} else {\n    h()\n}\nfor x in a {\n    break\n}",
			Expected: []int{1, 2, 3, 5, 7, 9, 10},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			script, err := New("", c.Input).Parse()
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			got := stmtLines(script.Stmts)
			if !slices.Equal(got, c.Expected) {
				t.Errorf("got %v, want %v", got, c.Expected)
			}
		})
	}
}

func TestParserFile(t *testing.T) {
	t.Run("missing_file", func(t *testing.T) {
		_, err := New("testdata/missing.ak", "").Parse()
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("expected a not exist error, got %v", err)
		}
	})

	t.Run("error_has_file_prefix", func(t *testing.T) {
		_, err := New("testdata/bad.ak", "").Parse()
		want := "testdata/bad.ak:2:9: error[leading-zero]"
		if err == nil || !strings.HasPrefix(err.Error(), want) {
			t.Errorf("expected prefix %q, got %v", want, err)
		}
	})
}

func TestParserParseTwice(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		{
			Name:  "valid_source",
			Input: "for x in a {\n    break\n}",
		},
		{
			Name:  "invalid_source",
			Input: "let a = ",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			p := New("", c.Input)
			first, firstErr := p.Parse()
			second, secondErr := p.Parse()
			if fmt.Sprint(firstErr) != fmt.Sprint(secondErr) {
				t.Fatalf("errors differ: %v and %v", firstErr, secondErr)
			}

			if firstErr == nil && first.String() != second.String() {
				t.Errorf("output differs:\n%s\n%s", first, second)
			}
		})
	}
}

func extractErrors() ([]tests_utils.Case[string, string], error) {
	var cases []tests_utils.Case[string, string]

	text, err := os.ReadFile("testdata/errors.ak")
	if err != nil {
		return nil, err
	}

	header := `let a = 1
var b
b = a + 1`

	count := 1
	for part := range strings.SplitSeq(string(text), "---") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		caseErr, ok := extractErrFromCase(part)
		if !ok {
			return nil, fmt.Errorf("case %d has no // ### marker: %q", count, part)
		}

		cases = append(cases, tests_utils.Case[string, string]{
			Name:  fmt.Sprintf("error_%d", count),
			Input: header + "\n" + part,
			Err:   caseErr,
		})
		count++
	}

	return cases, nil
}

func extractErrFromCase(text string) (error, bool) {
	errText := ""
	for textLine := range strings.SplitSeq(text, "\n") {
		if _, after, ok := strings.Cut(textLine, "// ###"); ok {
			errText = strings.TrimSpace(after)
		}
	}

	if errText == "" {
		return nil, false
	}

	return errors.New(errText), true
}

func validateParserOutput(t *testing.T, c tests_utils.Case[string, string]) {
	t.Helper()
	var stringParser Parser
	if filePath, ok := strings.CutPrefix(c.Input, "file://"); ok {
		stringParser = New(filePath, "")
	} else {
		stringParser = New("", c.Input)
	}

	script, err := stringParser.Parse()
	if c.Err != nil {
		if err == nil {
			t.Fatalf("expected error: %v, got none", c.Err)
		}

		if _, ok := errors.AsType[*lexer.Error](err); !ok {
			t.Fatalf("expected a *lexer.Error, got %T: %v", err, err)
		}

		if !strings.Contains(err.Error(), c.Err.Error()) {
			t.Fatalf("expected error: %v, got: %v", c.Err, err)
		}

		return
	}

	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var expected = c.Expected
	var expectedContent []byte
	if filePath, ok := strings.CutPrefix(expected, "file://"); ok {
		expectedContent, err = os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read expected file: %v", err)
		}

		expected = string(expectedContent)
	}

	got := script.String()
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}

	again, err := New("", got).Parse()
	if err != nil {
		t.Fatalf("reparse canonical output: %v", err)
	}

	if again.String() != got {
		t.Errorf("canonical output is not stable:\nfirst: %s\nsecond: %s", got, again)
	}
}

func nestedIfs(n int) string {
	return strings.Repeat("if a {\n", n) + strings.Repeat("}\n", n)
}

func stmtLines(stmts []ast.Stmt) []int {
	var lines []int
	for _, s := range stmts {
		lines = append(lines, s.Position().Line)
		switch s := s.(type) {
		case *ast.IfStmt:
			lines = append(lines, stmtLines(s.Then.Stmts)...)
			switch e := s.Else.(type) {
			case *ast.IfStmt:
				lines = append(lines, stmtLines([]ast.Stmt{e})...)
			case *ast.Block:
				lines = append(lines, stmtLines(e.Stmts)...)
			}
		case *ast.ForInStmt:
			lines = append(lines, stmtLines(s.Body.Stmts)...)
		case *ast.ForRangeStmt:
			lines = append(lines, stmtLines(s.Body.Stmts)...)
		}
	}

	return lines
}

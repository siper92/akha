package lexer

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestLexCanTokenize(t *testing.T) {
	cases := []tests_utils.Case[string, []Kind]{
		{Name: "empty", Input: "", Expected: []Kind{EOF}},
		{
			Name:  "all_keywords",
			Input: strings.Join(LexKeywords, " "),
			Expected: []Kind{
				Let, Var, If, Else, For, In, Range,
				Break, Continue, Return, Exit,
				And, Or, Not,
				True, False, Null,
				EOF,
			},
		},
		{
			Name:  "arithmetic",
			Input: "+ - * / % == != < <= > >= =",
			Expected: []Kind{
				Plus, Minus, Star, Slash, Percent,
				Eq, NotEq, Lt, LtEq, Gt, GtEq,
				Assign, EOF,
			},
		},
		{
			Name:  "delimiters",
			Input: "( ) [ ] { } , : . ..",
			Expected: []Kind{
				LParen, RParen, LBracket, RBracket,
				LBrace, RBrace, Comma, Colon,
				Dot, DotDot, EOF,
			},
		},
		{
			Name:  "identifiers",
			Input: "foo bar baz",
			Expected: []Kind{
				Ident, Ident, Ident, EOF,
			},
		},
		{
			Name:  "numbers",
			Input: "123 45.67 0.89",
			Expected: []Kind{
				Number, Number, Number, EOF,
			},
		},
		{
			Name:  "strings",
			Input: `"hello" "world"`,
			Expected: []Kind{
				String, String, EOF,
			},
		},
		{
			Name:  "simple_object",
			Input: `{"key": "value", "num": 42}`,
			Expected: []Kind{
				LBrace, String, Colon, String, Comma,
				String, Colon, Number, RBrace,
				EOF,
			},
		},
		//  tokens with mixed content
		{
			Name:  "mixed",
			Input: `let x = 42 if x > 10 { return "big" } else { return "small" }`,
			Expected: []Kind{
				Let, Ident, Assign, Number,
				If, Ident, Gt, Number, LBrace,
				Return, String, RBrace,
				Else, LBrace,
				Return, String, RBrace,
				EOF,
			},
		},
		// special characters
		{
			Name:  "newlines",
			Input: "let x = 42\n\n\n",
			Expected: []Kind{
				Let, Ident, Assign, Number,
				Newline, Newline, Newline,
				EOF,
			},
		},
		{
			Name:  "whitespace + mixed newlines",
			Input: "let a = 1\n\nif a {\n\r\n    let b = \"x\"\n}\r\n",
			Expected: []Kind{
				Let, Ident, Assign, Number,
				Newline, Newline,
				If, Ident, LBrace,
				Newline, Newline,
				Let, Ident, Assign, String,
				Newline,
				RBrace,
				Newline,
				EOF,
			},
		},
		// comments
		{
			Name:  "comments", // comments are ignored by the lexer
			Input: "let x = 42  // this is a comment\nx = x + 1",
			Expected: []Kind{
				Let, Ident, Assign, Number,
				Newline,
				Ident, Assign, Ident, Plus, Number,
				EOF,
			},
		},
		// template strings
		{
			Name:  "template_string",
			Input: `"Hello, ${name}!"`,
			Expected: []Kind{
				Template, EOF,
			},
		},
		// --- keywords
		{
			Name:  "keywords_case_insensitive",
			Input: "LET Var iF ELSE True NULL",
			Expected: []Kind{
				Let, Var, If, Else, True, Null, EOF,
			},
		},
		{
			Name:  "reserved_words",
			Input: "fn try catch FN",
			Expected: []Kind{
				Reserved, Reserved, Reserved, Reserved, EOF,
			},
		},
		{
			Name:  "keyword_prefixes_are_identifiers",
			Input: "_ _a a1 letx iff notin fnx",
			Expected: []Kind{
				Ident, Ident, Ident, Ident, Ident, Ident, Ident, EOF,
			},
		},
		{
			Name:  "not_in_is_two_tokens",
			Input: "a not in b",
			Expected: []Kind{
				Ident, Not, In, Ident, EOF,
			},
		},
		// --- operators
		{
			Name:  "symbolic_logic",
			Input: "&& || !",
			Expected: []Kind{
				And, Or, Not, EOF,
			},
		},
		{
			Name:  "compact_operators",
			Input: "a==b!=c<=d>=e",
			Expected: []Kind{
				Ident, Eq, Ident, NotEq, Ident, LtEq, Ident, GtEq, Ident, EOF,
			},
		},
		{
			Name:  "operator_pairs_split",
			Input: "!== => <>",
			Expected: []Kind{
				NotEq, Assign, Assign, Gt, Lt, Gt, EOF,
			},
		},
		{
			Name:  "member_chain",
			Input: "a.b[0](c)",
			Expected: []Kind{
				Ident, Dot, Ident, LBracket, Number, RBracket,
				LParen, Ident, RParen, EOF,
			},
		},
		// --- numbers and dots
		{
			Name:  "range_of_integers",
			Input: "[0..10]",
			Expected: []Kind{
				LBracket, Number, DotDot, Number, RBracket, EOF,
			},
		},
		{
			Name:  "range_of_floats",
			Input: "1.5..2",
			Expected: []Kind{
				Number, DotDot, Number, EOF,
			},
		},
		{
			Name:  "range_of_names",
			Input: "a.b..c",
			Expected: []Kind{
				Ident, Dot, Ident, DotDot, Ident, EOF,
			},
		},
		{
			Name:  "number_trailing_dot",
			Input: "1.",
			Expected: []Kind{
				Number, Dot, EOF,
			},
		},
		{
			Name:  "number_leading_dot",
			Input: ".5",
			Expected: []Kind{
				Dot, Number, EOF,
			},
		},
		{
			Name:  "triple_dot",
			Input: "...",
			Expected: []Kind{
				DotDot, Dot, EOF,
			},
		},
		{
			Name:  "zero_values",
			Input: "0 0.0 0.5",
			Expected: []Kind{
				Number, Number, Number, EOF,
			},
		},
		// --- comments and slashes
		{
			Name:  "slash_is_not_a_comment",
			Input: "a / b a/b",
			Expected: []Kind{
				Ident, Slash, Ident, Ident, Slash, Ident, EOF,
			},
		},
		{
			Name:  "comment_after_number",
			Input: "1// c",
			Expected: []Kind{
				Number, EOF,
			},
		},
		{
			Name:  "comment_only",
			Input: "// just a comment",
			Expected: []Kind{
				EOF,
			},
		},
		{
			Name:  "comment_lines_keep_newlines",
			Input: "// a\n// b\n",
			Expected: []Kind{
				Newline, Newline, EOF,
			},
		},
		{
			Name:  "comment_inside_string",
			Input: `"http://x"`,
			Expected: []Kind{
				String, EOF,
			},
		},
		// --- strings
		{
			Name:  "empty_string",
			Input: `""`,
			Expected: []Kind{
				String, EOF,
			},
		},
		{
			Name:  "escaped_dollar_is_string",
			Input: `"\${x}"`,
			Expected: []Kind{
				String, EOF,
			},
		},
		{
			Name:  "plain_dollar_is_string",
			Input: `"$5 $"`,
			Expected: []Kind{
				String, EOF,
			},
		},
		{
			Name:  "template_with_string_index",
			Input: `"${a["k"]}"`,
			Expected: []Kind{
				Template, EOF,
			},
		},
		{
			Name:  "template_with_braces_inside",
			Input: `"${ {} }"`,
			Expected: []Kind{
				Template, EOF,
			},
		},
		// --- whitespace
		{
			Name:  "tabs_and_lone_cr",
			Input: "\ta\t\r=\r1",
			Expected: []Kind{
				Ident, Assign, Number, EOF,
			},
		},
		{
			Name:  "crlf_is_one_newline",
			Input: "a\r\nb",
			Expected: []Kind{
				Ident, Newline, Ident, EOF,
			},
		},
		{
			Name:  "bom_is_skipped",
			Input: "\xEF\xBB\xBFlet",
			Expected: []Kind{
				Let, EOF,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateTokenizerOutput(t, c)
		})
	}
}

func TestLexTokenLiterals(t *testing.T) {
	cases := []tests_utils.Case[string, []Token]{
		// --- source text is kept
		{
			Name:  "keyword_case_is_kept",
			Input: "LET",
			Expected: []Token{
				{Kind: Let, Lit: "LET"},
				{Kind: EOF},
			},
		},
		{
			Name:  "symbolic_aliases",
			Input: "&& || !",
			Expected: []Token{
				{Kind: And, Lit: "&&"},
				{Kind: Or, Lit: "||"},
				{Kind: Not, Lit: "!"},
				{Kind: EOF},
			},
		},
		{
			Name:  "number_raw_text",
			Input: "3.50 10",
			Expected: []Token{
				{Kind: Number, Lit: "3.50"},
				{Kind: Number, Lit: "10"},
				{Kind: EOF},
			},
		},
		{
			Name:  "newline_literal",
			Input: "\n",
			Expected: []Token{
				{Kind: Newline, Lit: "\n"},
				{Kind: EOF},
			},
		},
		// --- strings are unescaped
		{
			Name:  "string_escapes",
			Input: `"a\tb\"c\\d\$e\n\r"`,
			Expected: []Token{
				{Kind: String, Lit: "a\tb\"c\\d$e\n\r"},
				{Kind: EOF},
			},
		},
		{
			Name:  "string_unicode",
			Input: `"héllo ✓"`,
			Expected: []Token{
				{Kind: String, Lit: "héllo ✓"},
				{Kind: EOF},
			},
		},
		{
			Name:  "template_lit_is_raw_source",
			Input: `"a ${b}"`,
			Expected: []Token{
				{Kind: Template, Lit: `"a ${b}"`},
				{Kind: EOF},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			toks, err := Tokenize(c.Input)
			if err != nil {
				t.Fatalf("tokenize %q: %v", c.Input, err)
			}

			if len(toks) != len(c.Expected) {
				t.Fatalf("token count mismatch for %q: got %d, want %d", c.Input, len(toks), len(c.Expected))
			}

			for i, tok := range toks {
				want := c.Expected[i]
				if tok.Kind != want.Kind || tok.Lit != want.Lit {
					t.Errorf("token %d: got %v %q, want %v %q", i, tok.Kind, tok.Lit, want.Kind, want.Lit)
				}
			}
		})
	}
}

func TestLexPositions(t *testing.T) {
	cases := []tests_utils.Case[string, []Pos]{
		// --- line and column
		{
			Name:  "two_lines",
			Input: "let x = 1\n  y",
			Expected: []Pos{
				{Line: 1, Col: 1},
				{Line: 1, Col: 5},
				{Line: 1, Col: 7},
				{Line: 1, Col: 9},
				{Line: 1, Col: 10},
				{Line: 2, Col: 3},
				{Line: 2, Col: 4},
			},
		},
		{
			Name:  "columns_count_runes",
			Input: `"é" x`,
			Expected: []Pos{
				{Line: 1, Col: 1},
				{Line: 1, Col: 5},
				{Line: 1, Col: 6},
			},
		},
		{
			Name:  "tab_is_one_column",
			Input: "\ta",
			Expected: []Pos{
				{Line: 1, Col: 2},
				{Line: 1, Col: 3},
			},
		},
		{
			Name:  "comment_is_skipped",
			Input: "a // c\nb",
			Expected: []Pos{
				{Line: 1, Col: 1},
				{Line: 1, Col: 7},
				{Line: 2, Col: 1},
				{Line: 2, Col: 2},
			},
		},
		// --- normalised input
		{
			Name:  "crlf",
			Input: "\ta\r\nb",
			Expected: []Pos{
				{Line: 1, Col: 2},
				{Line: 1, Col: 3},
				{Line: 2, Col: 1},
				{Line: 2, Col: 2},
			},
		},
		{
			Name:  "bom",
			Input: "\xEF\xBB\xBFlet",
			Expected: []Pos{
				{Line: 1, Col: 1},
				{Line: 1, Col: 4},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			toks, err := Tokenize(c.Input)
			if err != nil {
				t.Fatalf("tokenize %q: %v", c.Input, err)
			}

			got := make([]Pos, len(toks))
			for i, tok := range toks {
				got[i] = tok.Pos
			}

			if !slices.Equal(got, c.Expected) {
				t.Errorf("positions mismatch for %q:\n got: %v\nwant: %v", c.Input, got, c.Expected)
			}
		})
	}
}

func TestLexNewAtOffsetsPositions(t *testing.T) {
	toks, err := Collect(NewAt("a b", Pos{Line: 3, Col: 7}))
	if err != nil {
		t.Fatalf("collect: %v", err)
	}

	want := []Pos{{Line: 3, Col: 7}, {Line: 3, Col: 9}, {Line: 3, Col: 10}}
	for i, tok := range toks {
		if tok.Pos != want[i] {
			t.Errorf("token %d: got %v, want %v", i, tok.Pos, want[i])
		}
	}
}

func TestLexTemplateParts(t *testing.T) {
	cases := []tests_utils.Case[string, []Part]{
		// --- text and expressions
		{
			Name:  "text_expr_text",
			Input: `"a ${b} c"`,
			Expected: []Part{
				{Text: "a "},
				{Expr: "b", IsExpr: true, Pos: Pos{Line: 1, Col: 6}},
				{Text: " c"},
			},
		},
		{
			Name:  "adjacent_exprs",
			Input: `"${a}${b}"`,
			Expected: []Part{
				{Expr: "a", IsExpr: true, Pos: Pos{Line: 1, Col: 4}},
				{Expr: "b", IsExpr: true, Pos: Pos{Line: 1, Col: 8}},
			},
		},
		{
			Name:  "expr_keeps_spaces",
			Input: `"${ a.b }"`,
			Expected: []Part{
				{Expr: " a.b ", IsExpr: true, Pos: Pos{Line: 1, Col: 4}},
			},
		},
		{
			Name:  "expr_with_string_index",
			Input: `"${m["}"]}"`,
			Expected: []Part{
				{Expr: `m["}"]`, IsExpr: true, Pos: Pos{Line: 1, Col: 4}},
			},
		},
		// --- escapes around expressions
		{
			Name:  "escape_after_expr",
			Input: `"${a}\n"`,
			Expected: []Part{
				{Expr: "a", IsExpr: true, Pos: Pos{Line: 1, Col: 4}},
				{Text: "\n"},
			},
		},
		{
			Name:  "escaped_dollar_before_expr",
			Input: `"\${x} ${y}"`,
			Expected: []Part{
				{Text: "${x} "},
				{Expr: "y", IsExpr: true, Pos: Pos{Line: 1, Col: 10}},
			},
		},
		{
			Name:  "dollar_before_expr",
			Input: `"$${a}"`,
			Expected: []Part{
				{Text: "$"},
				{Expr: "a", IsExpr: true, Pos: Pos{Line: 1, Col: 5}},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			toks, err := Tokenize(c.Input)
			if err != nil {
				t.Fatalf("tokenize %q: %v", c.Input, err)
			}

			if toks[0].Kind != Template {
				t.Fatalf("expected a template, got %v", toks[0].Kind)
			}

			if !reflect.DeepEqual(toks[0].Parts, c.Expected) {
				t.Errorf("parts mismatch for %q:\n got: %+v\nwant: %+v", c.Input, toks[0].Parts, c.Expected)
			}
		})
	}
}

func TestLexErrors(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- punctuation
		{
			Name:  "semicolon",
			Input: "let a = 1;",
			Err:   errors.New("1:10: error[semicolon]: \";\" is not allowed\nhint: one statement per line"),
		},
		{
			Name:  "single_quote",
			Input: "'a'",
			Err:   errors.New("1:1: error[single-quote]: single quotes are not allowed"),
		},
		{
			Name:  "ampersand",
			Input: "a & b",
			Err:   errors.New(`1:3: error[unexpected-char]: unexpected character "&"`),
		},
		{
			Name:  "pipe",
			Input: "a | b",
			Err:   errors.New("1:3: error[unexpected-char]: unexpected character \"|\"\nhint: use or"),
		},
		{
			Name:  "at_sign",
			Input: "@",
			Err:   errors.New("1:1: error[unexpected-char]: unexpected character '@'"),
		},
		{
			Name:  "hash_comment",
			Input: "# note",
			Err:   errors.New("1:1: error[unexpected-char]: unexpected character '#'"),
		},
		{
			Name:  "unicode_identifier",
			Input: "let é = 1",
			Err:   errors.New("1:5: error[unexpected-char]: unexpected character 'é'"),
		},
		{
			Name:  "bom_in_the_middle",
			Input: "a\xEF\xBB\xBF",
			Err:   errors.New("1:2: error[unexpected-char]"),
		},
		// --- compound assignment
		{
			Name:  "plus_assign",
			Input: "a += 1",
			Err:   errors.New("1:3: error[compound-assign]: compound assignment += is not supported\nhint: use x = x + y"),
		},
		{
			Name:  "minus_assign",
			Input: "a -= 1",
			Err:   errors.New("error[compound-assign]: compound assignment -= is not supported"),
		},
		{
			Name:  "star_assign",
			Input: "a *= 1",
			Err:   errors.New("error[compound-assign]: compound assignment *= is not supported"),
		},
		{
			Name:  "slash_assign",
			Input: "a /= 1",
			Err:   errors.New("error[compound-assign]: compound assignment /= is not supported"),
		},
		{
			Name:  "percent_assign",
			Input: "a %= 1",
			Err:   errors.New("error[compound-assign]: compound assignment %= is not supported"),
		},
		// --- numbers
		{
			Name:  "leading_zero",
			Input: "let n = 007",
			Err:   errors.New("1:9: error[leading-zero]: numbers cannot have leading zeros\nhint: write 7 instead of 007"),
		},
		{
			Name:  "double_zero",
			Input: "00",
			Err:   errors.New("1:1: error[leading-zero]"),
		},
		{
			Name:  "leading_zero_float",
			Input: "01.5",
			Err:   errors.New("1:1: error[leading-zero]"),
		},
		{
			Name:  "number_with_letter",
			Input: "3x",
			Err:   errors.New("1:1: error[invalid-number]: invalid number literal"),
		},
		{
			Name:  "exponent",
			Input: "1e5",
			Err:   errors.New("1:1: error[invalid-number]"),
		},
		{
			Name:  "hex",
			Input: "0x1F",
			Err:   errors.New("1:1: error[invalid-number]"),
		},
		{
			Name:  "underscore_separator",
			Input: "1_000",
			Err:   errors.New("1:1: error[invalid-number]"),
		},
		{
			Name:  "float_with_suffix",
			Input: "1.5f",
			Err:   errors.New("1:1: error[invalid-number]"),
		},
		// --- strings
		{
			Name:  "unknown_escape",
			Input: `"a\qb"`,
			Err:   errors.New(`1:3: error[invalid-escape]: unknown escape \q`),
		},
		{
			Name:  "escaped_brace",
			Input: `"\{"`,
			Err:   errors.New(`1:2: error[invalid-escape]: unknown escape \{`),
		},
		{
			Name:  "unterminated",
			Input: `"abc`,
			Err:   errors.New("1:1: error[unterminated-string]: unterminated string"),
		},
		{
			Name:  "unterminated_after_backslash",
			Input: `"abc\`,
			Err:   errors.New("1:1: error[unterminated-string]"),
		},
		{
			Name:  "unterminated_after_interp",
			Input: `"${a} x`,
			Err:   errors.New("1:1: error[unterminated-string]"),
		},
		{
			Name:  "raw_newline",
			Input: "\"ab\ncd\"",
			Err:   errors.New("1:4: error[newline-in-string]: string cannot contain a raw newline"),
		},
		{
			Name:  "escaped_newline",
			Input: "\"ab\\\ncd\"",
			Err:   errors.New("1:5: error[newline-in-string]"),
		},
		{
			Name:  "crlf_in_string",
			Input: "\"ab\r\ncd\"",
			Err:   errors.New("1:4: error[newline-in-string]"),
		},
		// --- interpolation
		{
			Name:  "empty_interp",
			Input: `"${}"`,
			Err:   errors.New("1:2: error[empty-interp]: empty ${ } in string"),
		},
		{
			Name:  "blank_interp",
			Input: `"x ${  }"`,
			Err:   errors.New("1:4: error[empty-interp]"),
		},
		{
			Name:  "unterminated_interp",
			Input: `"${a`,
			Err:   errors.New("1:2: error[unterminated-interp]: missing } to close ${"),
		},
		{
			Name:  "unterminated_interp_in_string",
			Input: `"${a["k`,
			Err:   errors.New("1:2: error[unterminated-interp]"),
		},
		{
			Name:  "unterminated_interp_after_escape",
			Input: `"${a["k\`,
			Err:   errors.New("1:2: error[unterminated-interp]"),
		},
		{
			Name:  "newline_in_interp",
			Input: "\"${a\n}\"",
			Err:   errors.New("1:5: error[newline-in-string]"),
		},
		{
			Name:  "newline_in_interp_string",
			Input: "\"${a[\"k\n\"]}\"",
			Err:   errors.New("1:8: error[newline-in-string]"),
		},
		// --- encoding
		{
			Name:  "invalid_utf8",
			Input: "\xff",
			Err:   errors.New("1:1: error[invalid-utf8]: invalid UTF-8 in source"),
		},
		{
			Name:  "invalid_utf8_later_line",
			Input: "let a = 1\nlet b = \xff",
			Err:   errors.New("2:9: error[invalid-utf8]"),
		},
		{
			Name:  "invalid_utf8_in_string",
			Input: "\"a\xffb\"",
			Err:   errors.New("1:3: error[invalid-utf8]"),
		},
		// --- nesting
		{
			Name:  "nesting_over_limit",
			Input: strings.Repeat("(", MaxDepth+1),
			Err:   fmt.Errorf("1:%d: error[nesting]: excessive nesting", MaxDepth+1),
		},
		{
			Name:  "closers_do_not_go_below_zero",
			Input: strings.Repeat(")", 5) + strings.Repeat("(", MaxDepth+1),
			Err:   fmt.Errorf("1:%d: error[nesting]", MaxDepth+6),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateLexError(t, c)
		})
	}
}

func TestLexNestingWithinLimit(t *testing.T) {
	cases := []tests_utils.Case[string, int]{
		{
			Name:     "exactly_max_depth",
			Input:    strings.Repeat("(", MaxDepth),
			Expected: MaxDepth + 1,
		},
		{
			Name:     "mixed_brackets_at_max_depth",
			Input:    strings.Repeat("[", MaxDepth/2) + strings.Repeat("{", MaxDepth/2),
			Expected: MaxDepth + 1,
		},
		{
			Name:     "closed_pairs_reset_depth",
			Input:    strings.Repeat("()", MaxDepth*2),
			Expected: MaxDepth*4 + 1,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			toks, err := Tokenize(c.Input)
			if err != nil {
				t.Fatalf("tokenize: %v", err)
			}

			if len(toks) != c.Expected {
				t.Errorf("token count: got %d, want %d", len(toks), c.Expected)
			}
		})
	}
}

func TestLexErrorIsSticky(t *testing.T) {
	l := New("@ a")
	_, first := l.Next()
	_, second := l.Next()
	if first == nil {
		t.Fatalf("expected an error")
	}

	if first != second {
		t.Errorf("expected the same error, got %v and %v", first, second)
	}
}

func validateTokenizerOutput(t *testing.T, c tests_utils.Case[string, []Kind]) {
	t.Helper()
	toks, err := Tokenize(c.Input)
	if err != nil {
		t.Fatalf("tokenize %q: %v", c.Input, err)
	}

	got := make([]Kind, len(toks))
	for i, tok := range toks {
		got[i] = tok.Kind
	}

	if !slices.Equal(got, c.Expected) {
		t.Errorf("token kinds mismatch for %q:\n got: %v\nwant: %v", c.Input, got, c.Expected)
	}
}

func validateLexError(t *testing.T, c tests_utils.Case[string, struct{}]) {
	t.Helper()
	_, err := Tokenize(c.Input)
	if err == nil {
		t.Fatalf("tokenize %q: expected error %q, got none", c.Input, c.Err)
	}

	if !errors.Is(err, ErrLex) {
		t.Errorf("tokenize %q: %v is not ErrLex", c.Input, err)
	}

	if !strings.Contains(err.Error(), c.Err.Error()) {
		t.Errorf("tokenize %q:\n got: %v\nwant: %v", c.Input, err, c.Err)
	}
}

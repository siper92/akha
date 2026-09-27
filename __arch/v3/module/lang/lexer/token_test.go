package lexer

import (
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestLookup(t *testing.T) {
	cases := []tests_utils.Case[string, Kind]{
		// --- keywords
		{
			Name:     "lower_case_keyword",
			Input:    "let",
			Expected: Let,
		},
		{
			Name:     "upper_case_keyword",
			Input:    "LET",
			Expected: Let,
		},
		{
			Name:     "mixed_case_keyword",
			Input:    "Null",
			Expected: Null,
		},
		// --- reserved
		{
			Name:     "reserved",
			Input:    "fn",
			Expected: Reserved,
		},
		{
			Name:     "reserved_mixed_case",
			Input:    "Catch",
			Expected: Reserved,
		},
		// --- identifiers
		{
			Name:     "plain_name",
			Input:    "name",
			Expected: Ident,
		},
		{
			Name:     "underscore",
			Input:    "_",
			Expected: Ident,
		},
		{
			Name:     "keyword_prefix",
			Input:    "letx",
			Expected: Ident,
		},
		{
			Name:     "not_in_is_not_a_keyword",
			Input:    "not in",
			Expected: Ident,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := Lookup(c.Input); got != c.Expected {
				t.Errorf("Lookup(%q): got %v, want %v", c.Input, got, c.Expected)
			}
		})
	}
}

func TestLexKeywordsMatchKinds(t *testing.T) {
	for _, kw := range LexKeywords {
		t.Run(kw, func(t *testing.T) {
			k := Lookup(kw)
			if !k.IsKeyword() {
				t.Fatalf("%q is not a keyword kind: %v", kw, k)
			}

			if k.String() != kw {
				t.Errorf("kind name: got %q, want %q", k.String(), kw)
			}
		})
	}
}

func TestKindString(t *testing.T) {
	cases := []tests_utils.Case[Kind, string]{
		{
			Name:     "eof",
			Input:    EOF,
			Expected: "EOF",
		},
		{
			Name:     "operator",
			Input:    Plus,
			Expected: "+",
		},
		{
			Name:     "not_in",
			Input:    NotIn,
			Expected: "not in",
		},
		{
			Name:     "unknown_kind",
			Input:    Kind(999),
			Expected: "Kind(999)",
		},
		{
			Name:     "negative_kind",
			Input:    Kind(-1),
			Expected: "Kind(-1)",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := c.Input.String(); got != c.Expected {
				t.Errorf("got %q, want %q", got, c.Expected)
			}
		})
	}
}

func TestKindIsKeyword(t *testing.T) {
	cases := []tests_utils.Case[Kind, bool]{
		{
			Name:     "first_keyword",
			Input:    Let,
			Expected: true,
		},
		{
			Name:     "last_keyword",
			Input:    Null,
			Expected: true,
		},
		{
			Name:     "reserved",
			Input:    Reserved,
			Expected: false,
		},
		{
			Name:     "ident",
			Input:    Ident,
			Expected: false,
		},
		{
			Name:     "not_in",
			Input:    NotIn,
			Expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := c.Input.IsKeyword(); got != c.Expected {
				t.Errorf("got %v, want %v", got, c.Expected)
			}
		})
	}
}

func TestTokenDescribe(t *testing.T) {
	cases := []tests_utils.Case[Token, string]{
		// --- fixed descriptions
		{
			Name:     "eof",
			Input:    Token{Kind: EOF},
			Expected: "end of file",
		},
		{
			Name:     "newline",
			Input:    Token{Kind: Newline, Lit: "\n"},
			Expected: "end of line",
		},
		// --- kind and literal
		{
			Name:     "ident",
			Input:    Token{Kind: Ident, Lit: "x"},
			Expected: "identifier x",
		},
		{
			Name:     "number",
			Input:    Token{Kind: Number, Lit: "1.5"},
			Expected: "number 1.5",
		},
		{
			Name:     "reserved",
			Input:    Token{Kind: Reserved, Lit: "fn"},
			Expected: "reserved fn",
		},
		{
			Name:     "string_is_quoted",
			Input:    Token{Kind: String, Lit: `a"b`},
			Expected: `string "a\"b"`,
		},
		// --- quoted literal
		{
			Name:     "keyword_keeps_case",
			Input:    Token{Kind: Let, Lit: "LET"},
			Expected: `"LET"`,
		},
		{
			Name:     "symbolic_and",
			Input:    Token{Kind: And, Lit: "&&"},
			Expected: `"&&"`,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := c.Input.Describe(); got != c.Expected {
				t.Errorf("got %q, want %q", got, c.Expected)
			}
		})
	}
}

func TestPosIsValid(t *testing.T) {
	cases := []tests_utils.Case[Pos, bool]{
		{
			Name:     "zero",
			Input:    Pos{},
			Expected: false,
		},
		{
			Name:     "column_only",
			Input:    Pos{Col: 3},
			Expected: false,
		},
		{
			Name:     "first_line",
			Input:    Pos{Line: 1, Col: 1},
			Expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := c.Input.IsValid(); got != c.Expected {
				t.Errorf("got %v, want %v", got, c.Expected)
			}
		})
	}
}

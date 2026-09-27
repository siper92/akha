package token_test

import (
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/token"
)

func TestKeywords(t *testing.T) {
	// --- keywords_are_single_tokens
	var cases []tests_utils.LexSameCase
	for _, kw := range tests_utils.LexKeywords {
		cases = append(cases, tests_utils.LexSameCase{
			Name: "keyword_" + kw,
			Src:  kw,
			Ref:  "x",
		})
	}
	tests_utils.RunLexCount(t, cases)
}

func TestReserved(t *testing.T) {
	// --- reserved_words_are_single_tokens
	var cases []tests_utils.LexSameCase
	for _, rw := range tests_utils.LexReserved {
		cases = append(cases, tests_utils.LexSameCase{
			Name: "reserved_" + rw,
			Src:  rw,
			Ref:  "x",
		})
	}
	tests_utils.RunLexCount(t, cases)
}

func TestIdentifiers(t *testing.T) {
	// --- keywords_case_insensitive
	tests_utils.RunLexCount(t, []tests_utils.LexSameCase{
		{
			Name: "title_true_is_one_token",
			Src:  "True",
			Ref:  "x",
		},
		{
			Name: "upper_let_is_one_token",
			Src:  "LET",
			Ref:  "x",
		},
		{
			Name: "title_null_is_one_token",
			Src:  "Null",
			Ref:  "x",
		},
		{
			Name: "mixed_else_if_is_two_tokens",
			Src:  "eLsE If",
			Ref:  "a b",
		},
		{
			Name: "upper_not_in_is_two_tokens",
			Src:  "NOT IN",
			Ref:  "a b",
		},
		{
			Name: "title_reserved_fn_is_one_token",
			Src:  "Fn",
			Ref:  "x",
		},
	})

	// --- identifiers_case_sensitive
	tests_utils.RunLexCount(t, []tests_utils.LexSameCase{
		{
			Name: "upper_ident",
			Src:  "Name",
			Ref:  "x",
		},
		{
			Name: "upper_keyword_prefix",
			Src:  "LETTER",
			Ref:  "x",
		},
	})

	// --- identifier_shapes
	tests_utils.RunLexCount(t, []tests_utils.LexSameCase{
		{
			Name: "underscore_alone",
			Src:  "_",
			Ref:  "x",
		},
		{
			Name: "underscore_prefix",
			Src:  "_a1",
			Ref:  "x",
		},
		{
			Name: "mixed_case_digits_underscores",
			Src:  "A_b_9",
			Ref:  "x",
		},
		{
			Name: "camel_case",
			Src:  "byKey",
			Ref:  "x",
		},
		{
			Name: "trailing_digits",
			Src:  "arr3",
			Ref:  "x",
		},
	})

	// --- keyword_prefix_is_one_ident
	tests_utils.RunLexCount(t, []tests_utils.LexSameCase{
		{
			Name: "let_prefix",
			Src:  "letter",
			Ref:  "x",
		},
		{
			Name: "if_prefix",
			Src:  "iffy",
			Ref:  "x",
		},
		{
			Name: "for_prefix",
			Src:  "format",
			Ref:  "x",
		},
		{
			Name: "in_prefix",
			Src:  "input",
			Ref:  "x",
		},
		{
			Name: "not_in_joined",
			Src:  "notin",
			Ref:  "x",
		},
		{
			Name: "null_prefix",
			Src:  "nullable",
			Ref:  "x",
		},
		{
			Name: "true_prefix",
			Src:  "trueish",
			Ref:  "x",
		},
		{
			Name: "range_suffix_underscore",
			Src:  "range_",
			Ref:  "x",
		},
		{
			Name: "fn_prefix",
			Src:  "fnx",
			Ref:  "x",
		},
		{
			Name: "let_digit_suffix",
			Src:  "let1",
			Ref:  "x",
		},
	})

	// --- keywords_split_on_space
	tests_utils.RunLexCount(t, []tests_utils.LexSameCase{
		{
			Name: "not_in",
			Src:  "not in",
			Ref:  "a b",
		},
		{
			Name: "else_if",
			Src:  "else if",
			Ref:  "a b",
		},
		{
			Name: "for_var",
			Src:  "for var",
			Ref:  "a b",
		},
	})
}

func TestPos(t *testing.T) {
	// --- pos_is_comparable
	var a, b token.Pos
	if a != b {
		t.Fatalf("zero positions differ\n got: %v\nwant: %v", a, b)
	}

	// --- pos_string
	if got := (token.Pos{Line: 3, Col: 7}).String(); got != "3:7" {
		t.Fatalf("pos string\n got: %s\nwant: 3:7", got)
	}

	// --- pos_valid
	if a.IsValid() || !(token.Pos{Line: 1, Col: 1}).IsValid() {
		t.Fatalf("zero pos must be invalid, 1:1 valid")
	}
}

func TestLookup(t *testing.T) {
	// --- keywords and reserved words ignore case
	cases := []struct {
		name string
		src  string
		want token.Kind
	}{
		{name: "lower_keyword", src: "let", want: token.Let},
		{name: "upper_keyword", src: "LET", want: token.Let},
		{name: "mixed_not", src: "nOt", want: token.Not},
		{name: "title_true", src: "True", want: token.True},
		{name: "reserved", src: "fn", want: token.Reserved},
		{name: "upper_reserved", src: "CATCH", want: token.Reserved},
		{name: "ident", src: "Name", want: token.Ident},
		{name: "keyword_prefix", src: "LETTER", want: token.Ident},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := token.Lookup(c.src); got != c.want {
				t.Fatalf("lookup %q\n got: %s\nwant: %s", c.src, got, c.want)
			}
		})
	}
}

func TestDescribe(t *testing.T) {
	// --- tokens in error messages
	cases := []struct {
		name string
		tok  token.Token
		want string
	}{
		{name: "eof", tok: token.Token{Kind: token.EOF}, want: "end of file"},
		{name: "newline", tok: token.Token{Kind: token.Newline, Lit: "\n"}, want: "end of line"},
		{name: "ident", tok: token.Token{Kind: token.Ident, Lit: "a"}, want: "identifier a"},
		{name: "number", tok: token.Token{Kind: token.Number, Lit: "3.2"}, want: "number 3.2"},
		{name: "reserved", tok: token.Token{Kind: token.Reserved, Lit: "fn"}, want: "reserved fn"},
		{name: "string", tok: token.Token{Kind: token.String, Lit: "a\"b"}, want: `string "a\"b"`},
		{name: "keyword_keeps_case", tok: token.Token{Kind: token.Let, Lit: "LET"}, want: `"LET"`},
		{name: "punct", tok: token.Token{Kind: token.RBrace, Lit: "}"}, want: `"}"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.tok.Describe(); got != c.want {
				t.Fatalf("describe\n got: %s\nwant: %s", got, c.want)
			}
		})
	}
}

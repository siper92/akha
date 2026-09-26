package token_test

import (
	"testing"

	"github.com/siper92/akha/lang/tests"
	"github.com/siper92/akha/lang/token"
)

func TestKeywords(t *testing.T) {
	// --- keywords_are_single_tokens
	var cases []tests.LexSameCase
	for _, kw := range tests.LexKeywords {
		cases = append(cases, tests.LexSameCase{
			Name: "keyword_" + kw,
			Src:  kw,
			Ref:  "x",
		})
	}
	tests.RunLexCount(t, cases)
}

func TestReserved(t *testing.T) {
	// --- reserved_words_are_single_tokens
	var cases []tests.LexSameCase
	for _, rw := range tests.LexReserved {
		cases = append(cases, tests.LexSameCase{
			Name: "reserved_" + rw,
			Src:  rw,
			Ref:  "x",
		})
	}
	tests.RunLexCount(t, cases)
}

func TestIdentifiers(t *testing.T) {
	// --- keywords_case_insensitive
	tests.RunLexCount(t, []tests.LexSameCase{
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
	tests.RunLexCount(t, []tests.LexSameCase{
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
	tests.RunLexCount(t, []tests.LexSameCase{
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
	tests.RunLexCount(t, []tests.LexSameCase{
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
	tests.RunLexCount(t, []tests.LexSameCase{
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
}

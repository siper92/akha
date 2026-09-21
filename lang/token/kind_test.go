package token_test

import (
	"fmt"
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/token"
)

func TestKindString(t *testing.T) {
	cases := []tu.Case[token.Kind, string]{
		// --- known kinds
		{
			Name:     "illegal",
			Input:    token.ILLEGAL,
			Expected: "illegal",
		},
		{
			Name:     "eof",
			Input:    token.EOF,
			Expected: "eof",
		},
		{
			Name:     "newline",
			Input:    token.NEWLINE,
			Expected: "newline",
		},
		{
			Name:     "comment",
			Input:    token.COMMENT,
			Expected: "comment",
		},
		{
			Name:     "ident",
			Input:    token.IDENT,
			Expected: "ident",
		},
		{
			Name:     "string",
			Input:    token.STRING,
			Expected: "string",
		},
		{
			Name:     "int",
			Input:    token.INT,
			Expected: "int",
		},
		{
			Name:     "dot",
			Input:    token.DOT,
			Expected: "dot",
		},
		{
			Name:     "lparen",
			Input:    token.LPAREN,
			Expected: "lparen",
		},
		{
			Name:     "rparen",
			Input:    token.RPAREN,
			Expected: "rparen",
		},
		{
			Name:     "comma",
			Input:    token.COMMA,
			Expected: "comma",
		},
		{
			Name:     "assign",
			Input:    token.ASSIGN,
			Expected: "assign",
		},
		{
			Name:     "ellipsis",
			Input:    token.ELLIPSIS,
			Expected: "ellipsis",
		},
		// --- unknown kinds
		{
			Name:     "one_past_last_is_fallback",
			Input:    token.Kind(13),
			Expected: "kind(13)",
		},
		{
			Name:     "negative_is_fallback",
			Input:    token.Kind(-1),
			Expected: "kind(-1)",
		},
		{
			Name:     "large_is_fallback",
			Input:    token.Kind(100),
			Expected: "kind(100)",
		},
	}
	tu.Run(tu.New(t), cases, func(k token.Kind) (string, error) { return k.String(), nil }, nil)
}

func TestKindFormat(t *testing.T) {
	cases := []tu.Case[token.Kind, string]{
		// --- stringer via fmt
		{
			Name:     "percent_v_uses_stringer",
			Input:    token.IDENT,
			Expected: "ident",
		},
		{
			Name:     "percent_v_fallback",
			Input:    token.Kind(42),
			Expected: "kind(42)",
		},
	}
	tu.Run(tu.New(t), cases, func(k token.Kind) (string, error) { return fmt.Sprintf("%v", k), nil }, nil)
}

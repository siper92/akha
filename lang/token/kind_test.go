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
		// --- punctuation and operators
		{
			Name:     "lbrace",
			Input:    token.LBRACE,
			Expected: "lbrace",
		},
		{
			Name:     "rbrace",
			Input:    token.RBRACE,
			Expected: "rbrace",
		},
		{
			Name:     "plus",
			Input:    token.PLUS,
			Expected: "plus",
		},
		{
			Name:     "minus",
			Input:    token.MINUS,
			Expected: "minus",
		},
		{
			Name:     "star",
			Input:    token.STAR,
			Expected: "star",
		},
		{
			Name:     "slash",
			Input:    token.SLASH,
			Expected: "slash",
		},
		{
			Name:     "percent",
			Input:    token.PERCENT,
			Expected: "percent",
		},
		{
			Name:     "eq",
			Input:    token.EQ,
			Expected: "eq",
		},
		{
			Name:     "neq",
			Input:    token.NEQ,
			Expected: "neq",
		},
		{
			Name:     "lt",
			Input:    token.LT,
			Expected: "lt",
		},
		{
			Name:     "lte",
			Input:    token.LTE,
			Expected: "lte",
		},
		{
			Name:     "gt",
			Input:    token.GT,
			Expected: "gt",
		},
		{
			Name:     "gte",
			Input:    token.GTE,
			Expected: "gte",
		},
		// --- keywords
		{
			Name:     "let",
			Input:    token.LET,
			Expected: "let",
		},
		{
			Name:     "if",
			Input:    token.IF,
			Expected: "if",
		},
		{
			Name:     "else",
			Input:    token.ELSE,
			Expected: "else",
		},
		{
			Name:     "for",
			Input:    token.FOR,
			Expected: "for",
		},
		{
			Name:     "in",
			Input:    token.IN,
			Expected: "in",
		},
		{
			Name:     "while",
			Input:    token.WHILE,
			Expected: "while",
		},
		{
			Name:     "break",
			Input:    token.BREAK,
			Expected: "break",
		},
		{
			Name:     "continue",
			Input:    token.CONTINUE,
			Expected: "continue",
		},
		{
			Name:     "and",
			Input:    token.AND,
			Expected: "and",
		},
		{
			Name:     "or",
			Input:    token.OR,
			Expected: "or",
		},
		{
			Name:     "not",
			Input:    token.NOT,
			Expected: "not",
		},
		{
			Name:     "true",
			Input:    token.TRUE,
			Expected: "true",
		},
		{
			Name:     "false",
			Input:    token.FALSE,
			Expected: "false",
		},
		// --- unknown kinds
		{
			Name:     "one_past_last_is_fallback",
			Input:    token.Kind(39),
			Expected: "kind(39)",
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

func TestLookup(t *testing.T) {
	cases := []tu.Case[string, token.Kind]{
		// --- keywords
		{
			Name:     "let",
			Input:    "let",
			Expected: token.LET,
		},
		{
			Name:     "if",
			Input:    "if",
			Expected: token.IF,
		},
		{
			Name:     "else",
			Input:    "else",
			Expected: token.ELSE,
		},
		{
			Name:     "for",
			Input:    "for",
			Expected: token.FOR,
		},
		{
			Name:     "in",
			Input:    "in",
			Expected: token.IN,
		},
		{
			Name:     "while",
			Input:    "while",
			Expected: token.WHILE,
		},
		{
			Name:     "break",
			Input:    "break",
			Expected: token.BREAK,
		},
		{
			Name:     "continue",
			Input:    "continue",
			Expected: token.CONTINUE,
		},
		{
			Name:     "and",
			Input:    "and",
			Expected: token.AND,
		},
		{
			Name:     "or",
			Input:    "or",
			Expected: token.OR,
		},
		{
			Name:     "not",
			Input:    "not",
			Expected: token.NOT,
		},
		{
			Name:     "true",
			Input:    "true",
			Expected: token.TRUE,
		},
		{
			Name:     "false",
			Input:    "false",
			Expected: token.FALSE,
		},
		// --- not keywords
		{
			Name:     "plain_ident",
			Input:    "name",
			Expected: token.IDENT,
		},
		{
			Name:     "case_sensitive",
			Input:    "Let",
			Expected: token.IDENT,
		},
		{
			Name:     "keyword_prefix",
			Input:    "letter",
			Expected: token.IDENT,
		},
		{
			Name:     "empty",
			Input:    "",
			Expected: token.IDENT,
		},
	}
	tu.Run(tu.New(t), cases, func(s string) (token.Kind, error) { return token.Lookup(s), nil }, nil)
}

package parser_test

import (
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/parser"
)

func TestErrorFormat(t *testing.T) {
	cases := []tu.Case[parser.Error, string]{
		// --- line and col prefix
		{
			Name:     "line_col_prefix",
			Input:    perr(3, 4, "expected '(', got newline"),
			Expected: "3:4: expected '(', got newline",
		},
		{
			Name:     "zero_pos",
			Input:    perr(0, 0, "x"),
			Expected: "0:0: x",
		},
	}
	tu.Run(tu.New(t), cases, func(e parser.Error) (string, error) {
		return e.Error(), nil
	}, nil)
}

func TestErrorsFormat(t *testing.T) {
	cases := []tu.Case[parser.Errors, string]{
		// --- empty
		{
			Name:     "nil_errors",
			Input:    nil,
			Expected: "",
		},
		{
			Name:     "empty_errors",
			Input:    parser.Errors{},
			Expected: "",
		},
		// --- joined by newline
		{
			Name:     "one_error",
			Input:    perrs(perr(1, 1, "a")),
			Expected: "1:1: a",
		},
		{
			Name:     "two_errors",
			Input:    perrs(perr(1, 1, "a"), perr(2, 5, "b")),
			Expected: "1:1: a\n2:5: b",
		},
	}
	tu.Run(tu.New(t), cases, func(es parser.Errors) (string, error) {
		return es.Error(), nil
	}, nil)
}

func TestParseErrorString(t *testing.T) {
	cases := []tu.Case[string, string]{
		// --- error returned by Parse
		{
			Name:     "clean_has_no_error",
			Input:    `Ak.Log("a")`,
			Expected: "",
		},
		{
			Name:     "rule_and_syntax_errors_joined",
			Input:    "Ak.Log(yes)\nAk",
			Expected: "1:8: unexpected identifier\n2:3: expected '.', got eof",
		},
	}
	tu.Run(tu.New(t), cases, func(src string) (string, error) {
		_, err := parser.New(lexer.New(src)).Parse()
		if err == nil {
			return "", nil
		}
		return err.Error(), nil
	}, nil)
}

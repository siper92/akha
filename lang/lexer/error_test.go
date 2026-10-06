package lexer

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestErrorString(t *testing.T) {
	cases := []tests_utils.Case[*Error, string]{
		{
			Name: "file_and_hint",
			Input: &Error{
				Kind: ErrParse,
				File: "a.ak",
				Pos:  Pos{Line: 2, Col: 3},
				Code: "x",
				Msg:  "m",
				Hint: "h",
			},
			Expected: "a.ak:2:3: error[x]: m\nhint: h",
		},
		{
			Name:     "no_file_no_hint",
			Input:    NewLexError(ErrLex, Pos{Line: 1, Col: 1}, "y", "m", ""),
			Expected: "1:1: error[y]: m",
		},
		{
			Name:     "zero_position",
			Input:    &Error{Code: "z", Msg: "m"},
			Expected: "0:0: error[z]: m",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := c.Input.Error(); got != c.Expected {
				t.Errorf("got %q, want %q", got, c.Expected)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	err := error(NewLexError(ErrParse, Pos{Line: 1, Col: 1}, "x", "m", ""))
	if !errors.Is(err, ErrParse) {
		t.Errorf("expected ErrParse")
	}

	if errors.Is(err, ErrLex) {
		t.Errorf("did not expect ErrLex")
	}
}

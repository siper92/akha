package parser

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestParserArgs(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- positional then named arguments
		{
			Name:     "positional_then_named",
			Input:    "f(a == b, 1, c=2, d=x + 1)",
			Expected: "f(a == b, 1, c=2, d=x + 1)",
		},
		{
			Name:     "multi_line_named_trailing_comma",
			Input:    "f(\n    1,\n    b=2,\n)",
			Expected: "f(1, b=2)",
		},
		// --- named argument errors
		{
			Name:  "positional_after_named",
			Input: "f(a=1, 2)",
			Err:   errors.New("1:8: error[kwarg-order]: positional argument after a named argument"),
		},
		{
			Name:  "duplicate_named",
			Input: "f(a=1, a=2)",
			Err:   errors.New("1:8: error[duplicate-kwarg]: duplicate named argument a"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateParserOutput(t, c)
		})
	}
}

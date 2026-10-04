package parser

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestParserModuleArgs(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- module call arguments keep their order
		{
			Name:     "positional_across_lines",
			Input:    "fs.write(\n    \"a.txt\",\n    \"b\",\n)",
			Expected: `fs.write("a.txt", "b")`,
		},
		{
			Name:     "named_keep_order",
			Input:    `ak.setup(debug="x", log="y")`,
			Expected: `ak.setup(debug="x", log="y")`,
		},
		// --- bound arguments are checked against the signature
		{
			Name:  "positional_on_named_only",
			Input: `ak.setup("x")`,
			Err:   errors.New("1:9: error[arity]: wrong number of arguments for ak.setup, want 0, got 1"),
		},
		{
			Name:  "named_arg_kind",
			Input: "ak.setup(log=1)",
			Err:   errors.New("1:9: error[arg-kind]: argument log of ak.setup must be a string"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateParserOutput(t, c)
		})
	}
}

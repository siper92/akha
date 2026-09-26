package parser_test

import (
	"testing"

	"github.com/siper92/akha/internal/tu"
)

func TestParsingLoops(t *testing.T) {
	cases := []tu.Case[string, result]{
		{
			Name:  "while_loop",
			Input: "while true {\n  Ak.Log(x)\n }",
			Expected: result{Script: script(
				while(
					ident("x", 1, 7),
					block(1, 2, callStmt("Ak", "Log", 2, 3, args(ident("x", 2, 10)), nil)),
					1, 2,
				))},
		},
		{
			Name:     "for_loop",
			Input:    "for i range 10 {\n  Ak.Log(x) \n}",
			Expected: result{Script: script()},
		},
	}

	tu.Run(tu.New(t), cases, parseWith(), nil)
}

package parser_test

import (
	"testing"

	"github.com/siper92/akha/internal/tu"
)

func TestParsingLoops(t *testing.T) {
	cases := []tu.Case[string, result]{
		{
			Name: "while_loop",
			Input: `while true {
    Ak.Log(x)
}`,
			Expected: result{Script: script(
				while(
					boolean(true, 1, 7),
					block(2, 5, callStmt("Ak", "Log", 2, 3, args(ident("x", 2, 10)), nil)),
					1, 1,
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

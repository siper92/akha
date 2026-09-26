package parser_test

import (
	"testing"

	"github.com/siper92/akha/internal/tu"
)

func TestParsingConditions(t *testing.T) {
	cases := []tu.Case[string, result]{
		{
			Name:  "if_condition",
			Input: "if true {\n  Ak.Log(x)\n }",
			Expected: result{Script: script(
				if_exp(
					boolean(true, 1, 4),
					block(1, 2, callStmt("Ak", "Log", 2, 3, args(ident("x", 2, 10)), nil)),
					nil,
					1, 2,
				))},
		},
	}

	tu.Run(tu.New(t), cases, parseWith(), nil)
}

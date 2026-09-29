package eval_test

import (
	"errors"
	"os"
	"testing"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline"
)

func TestEvalModulo(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- % on numbers, sign follows the dividend
		{
			Name:     "modulo_integral",
			Input:    "return 10 % 3",
			Expected: "1",
		},
		{
			Name:     "modulo_fraction_and_negative",
			Input:    "return [7.5 % 2, -7 % 3]",
			Expected: "[1.5,-1]",
		},
		{
			Name:  "modulo_by_runtime_zero",
			Input: "var zero = 0\nreturn 1 % zero",
			Err:   errors.New("2: runtime error: modulo by zero"),
		},
		{
			Name:  "modulo_kind_mismatch",
			Input: "return \"a\" % 2",
			Err:   errors.New("1: runtime error: cannot modulo string and number"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateRun(t, c, nil)
		})
	}
}

func TestEvalState(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- scopes, value semantics and loop flow
		{
			Name:     "shadow_restored_after_block",
			Input:    "let i = 0\nfor var i in [1, 2] {\n    i = 5\n}\nreturn i",
			Expected: "0",
		},
		{
			Name:     "copy_on_store",
			Input:    "var a = [1, [2]]\nvar b = a\nb[1][0] = 9\nreturn [a, b]",
			Expected: "[[1,[2]],[1,[9]]]",
		},
		{
			Name:     "return_leaves_every_loop",
			Input:    "for i range [0..3] {\n    for k, v in ({a: 1, b: 2}) {\n        if v == 2 {\n            return \"${k}${i}\"\n        }\n    }\n}",
			Expected: `"b0"`,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateRun(t, c, nil)
		})
	}
}

func TestEvalErrors(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- runtime errors pass the check and stop the eval
		{
			Name:  "index_out_of_range",
			Input: "let a = [1]\nreturn a[1]",
			Err:   errors.New("2: runtime error: index 1 out of range for length 1"),
		},
		{
			Name:  "member_on_null",
			Input: "var o = null\nreturn o.name",
			Err:   errors.New("2: runtime error: access on null"),
		},
		{
			Name:  "iterate_number",
			Input: "var n = 5\nfor x in n {\n}",
			Err:   errors.New("2: runtime error: cannot iterate number"),
		},
		{
			Name:  "max_iterations",
			Input: "for i range [0..2000] {\n}",
			Err:   errors.New("1: runtime error: max iterations 1000 exceeded"),
		},
		// --- operators and targets fail on runtime kinds hidden by var
		{
			Name:  "division_by_runtime_zero",
			Input: "var zero = 0\nreturn 1 / zero",
			Err:   errors.New("2: runtime error: division by zero"),
		},
		{
			Name:  "compare_kind_mismatch",
			Input: "var a = 1\nreturn a < \"b\"",
			Err:   errors.New("2: runtime error: cannot compare number and string"),
		},
		{
			Name:  "array_index_kind",
			Input: "var a = [1]\nreturn a[\"x\"]",
			Err:   errors.New("2: runtime error: array index must be a number, got string"),
		},
		{
			Name:  "set_member_on_number",
			Input: "var n = 1\nn.x = 2",
			Err:   errors.New("2: runtime error: cannot set member x on number"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateRun(t, c, nil)
		})
	}
}

func TestEvalSpecDef(t *testing.T) {
	src, err := os.ReadFile("../parser/testdata/spec_def.ak")
	if err != nil {
		t.Fatal(err)
	}

	input, err := eval.NewCodec().FromGo(map[string]any{"items": []any{1.0}})
	if err != nil {
		t.Fatal(err)
	}

	c := tests_utils.Case[string, string]{
		Input:    string(src),
		Expected: `{"total":9,"greeting":"hello ann, first is 1"}`,
	}
	pipeline.ValidateRun(t, c, input)
}

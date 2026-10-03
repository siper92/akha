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

func TestEvalLoops(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- scopes, value semantics and loop flow
		{
			Name:     "loops_to_200",
			Input:    "var sum = 0\nfor i range [0..200] {\nsum = sum + i\n}\nreturn sum",
			Expected: "19900",
		},
		{
			//@TODO: increase max inout to more than - 1000 and test the max iterations error
			//original: 2000
			Name:     "loops_to_500",
			Input:    "var sum = 0\nfor i range [0..500] {\nsum = sum + 1\n}\nreturn sum",
			Expected: "500",
		},
		{
			Name:     "loops_till_break",
			Input:    "var sum = 0\nfor i range [0..200] {\nif i == 100 {\nbreak\n}\nsum = sum + 1\n}\nreturn sum",
			Expected: "100",
		},
		{
			//@TODO: error if the loop is infinite
			Name:     "loops_till_break",
			Input:    "var sum = 5\nfor i range [0..1000] {\n\n}\nreturn sum",
			Expected: "5",
			//Err:   errors.New("2: runtime error: infinite loop detected"),
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
			Name:     "copy_on_store",
			Input:    "var a = [1, [2,9]]\nvar b = a\nb[1][0] = 9\nreturn [a, b]",
			Expected: "[[1,[2,9]],[1,[9,9]]]",
		},
		{
			Name: "complex_range",
			Input: `
let pair2 = "testLeft2"
var pair1 = ""
var total = [1,2,3,4,5,100,200,300]
var totalRes
for n in total {
    if n > 200 {
		pair1 = "test" + pair2
        totalRes = "break3" + pair1
        break
    }
}
return totalRes`,
			Expected: `"break3testtestLeft2"`,
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

package lexer_test

import (
	"errors"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline"
)

func TestErrorCodesKwargs(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- named argument codes come from the parser
		{
			Name:  "kwarg_order",
			Input: `ak.setup(log="a", "b")`,
			Err:   errors.New("1:19: error[kwarg-order]: positional argument after a named argument"),
		},
		{
			Name:  "duplicate_kwarg",
			Input: `ak.setup(log="a", log="b")`,
			Err:   errors.New("1:19: error[duplicate-kwarg]: duplicate named argument log"),
		},
		{
			Name:  "unknown_kwarg",
			Input: `ak.setup(level="a")`,
			Err:   errors.New("1:9: error[unknown-kwarg]: unknown named argument level for ak.setup"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateParse(t, c)
		})
	}
}

func TestErrorCodesModuleParse(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- signatures of registered modules are checked while parsing
		{
			Name:  "unknown_func",
			Input: `ak.print("a")`,
			Err:   errors.New("1:9: error[unknown-func]: unknown function ak.print"),
		},
		{
			Name:  "arity",
			Input: "fs.read()",
			Err:   errors.New("1:8: error[arity]: wrong number of arguments for fs.read, want 1, got 0"),
		},
		{
			Name:  "arg_kind",
			Input: "ak.log(1)",
			Err:   errors.New("1:7: error[arg-kind]: argument msg of ak.log must be a string"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateParse(t, c)
		})
	}
}

func TestErrorCodesModuleCheck(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- module scope codes come from the checker
		{
			Name:  "unknown_module",
			Input: `net.get("a")`,
			Err:   errors.New("1: error[unknown-module]: unknown module net"),
		},
		{
			Name:  "not_imported",
			Input: `let s = fs.read("a")`,
			Err:   errors.New("1: error[not-imported]: module fs is not imported"),
		},
		{
			Name:  "import_arg",
			Input: "let sf = \"fs\"\nak.import(sf)",
			Err:   errors.New("2: error[import-arg]: import takes a module name as a string literal"),
		},
		{
			Name:  "import_scope",
			Input: "if input {\n    ak.import(\"fs\")\n}",
			Err:   errors.New("2: error[import-scope]: imports are allowed only in the root block"),
		},
		{
			Name:  "import_dup",
			Input: "ak.import(\"fs\")\nak.import(\"fs\")",
			Err:   errors.New("2: error[import-dup]: module fs is already imported"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateCheck(t, c)
		})
	}
}

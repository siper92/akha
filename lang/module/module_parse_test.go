package module_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module/std"
	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline_test"
)

func TestModuleParse(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- module calls keep their canonical form
		{
			Name:     "import_then_call",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"b\")",
			Expected: "ak.import(\"fs\")\nfs.write(\"a.txt\", \"b\")",
		},
		{
			Name:     "named_args_across_lines",
			Input:    "ak.setup(\n    log=\"info\",\n)",
			Expected: "ak.setup(log=\"info\")",
		},
		// --- registered signatures are checked while parsing
		{
			Name:  "missing_positional_arg",
			Input: "fs.write(\"a.txt\")",
			Err:   errors.New("1:9: error[arity]: wrong number of arguments for fs.write, want 2, got 1"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateParse(t, c)
		})
	}
}

func TestModuleImport(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- a module is callable only after ak.import
		{
			Name:  "call_without_import",
			Input: "let s = fs.read(\"a.txt\")",
			Err:   errors.New("1: error[not-imported]: module fs is not imported"),
		},
		{
			Name:  "import_then_call",
			Input: "ak.import(\"fs\")\nlet s = fs.read(\"a.txt\")",
		},
		{
			Name:  "core_is_always_imported",
			Input: "ak.log(\"hi\")",
		},
		{
			Name:  "call_on_declared_name",
			Input: "let o = {}\no.get()",
			Err:   errors.New("2: error[no-callable]: o is not a module"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateCheck(t, c)
		})
	}
}

func TestModuleRun(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- fs works inside a sandboxed root
		{
			Name:     "write_then_read",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"hi\")\nreturn fs.read(\"a.txt\")",
			Expected: `"hi"`,
		},
		{
			Name:  "read_outside_root",
			Input: "ak.import(\"fs\")\nreturn fs.read(\"../a.txt\")",
			Err:   errors.New("2: runtime error: fs.read:"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			reg, err := std.New(t.TempDir(), slog.New(slog.DiscardHandler))
			if err != nil {
				t.Fatalf("modules: %v", err)
			}

			pipeline_test.ValidateRunAsTest(t, c, eval.Options{Modules: reg}, nil)
		})
	}
}

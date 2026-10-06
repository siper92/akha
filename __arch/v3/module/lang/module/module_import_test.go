package module_test

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/siper92/akha/lang/module/std"
	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline"
)

func TestModuleImportRules(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- import targets must be known and imported once
		{
			Name:  "unknown_module",
			Input: "ak.import(\"net\")",
			Err:   errors.New("1: error[unknown-module]: unknown module net"),
		},
		{
			Name:  "duplicate_import",
			Input: "ak.import(\"fs\")\nak.import(\"fs\")",
			Err:   errors.New("2: error[import-dup]: module fs is already imported"),
		},
		// --- imports live in the root block and take a literal
		{
			Name:  "import_in_nested_block",
			Input: "if true {\n    ak.import(\"fs\")\n}",
			Err:   errors.New("2: error[import-scope]: imports are allowed only in the root block"),
		},
		{
			Name:  "import_from_variable",
			Input: "let m = \"fs\"\nak.import(m)",
			Err:   errors.New("2: error[import-arg]: import takes a module name as a string literal"),
		},
		// --- an import is visible in nested blocks
		{
			Name:  "imported_module_in_nested_block",
			Input: "ak.import(\"fs\")\nif fs.exists(\"a.txt\") {\n    let s = fs.read(\"a.txt\")\n}",
		},
		{
			Name:  "module_name_as_variable",
			Input: "let fs = 1",
			Err:   errors.New("1: error[module-name]: fs is a module name"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateCheck(t, c)
		})
	}
}

func TestModuleSignatures(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- module signatures are checked while parsing
		{
			Name:  "unknown_function",
			Input: "fs.remove(\"a.txt\")",
			Err:   errors.New("1:10: error[unknown-func]: unknown function fs.remove"),
		},
		{
			Name:  "wrong_arg_kind",
			Input: "fs.read(1)",
			Err:   errors.New("1:8: error[arg-kind]: argument path of fs.read must be a string"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateParse(t, c)
		})
	}
}

func TestModuleImportRun(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- imported modules run against the sandboxed root
		{
			Name:     "exists_after_write",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"hi\")\nreturn fs.exists(\"a.txt\")",
			Expected: "true",
		},
		{
			Name:     "missing_file",
			Input:    "ak.import(\"fs\")\nreturn fs.exists(\"b.txt\")",
			Expected: "false",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			reg, err := std.New(t.TempDir(), slog.New(slog.DiscardHandler))
			if err != nil {
				t.Fatalf("modules: %v", err)
			}

			pipeline.ValidateRunModules(t, c, reg)
		})
	}
}

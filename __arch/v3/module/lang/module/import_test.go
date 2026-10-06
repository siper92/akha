package module_test

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module/std"
	"github.com/siper92/akha/lang/runner"
	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline"
)

func TestImportCheck(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- order of import and use
		{
			Name:  "call_before_import",
			Input: "let s = fs.read(\"a.txt\")\nak.import(\"fs\")",
			Err:   errors.New("1: error[not-imported]: module fs is not imported"),
		},
		{
			Name:  "import_after_other_statements",
			Input: "let a = 1\nak.import(\"fs\")\nlet b = fs.exists(\"a.txt\")",
		},
		{
			Name:  "unknown_after_valid_import",
			Input: "ak.import(\"fs\")\nak.import(\"http\")",
			Err:   errors.New("2: error[unknown-module]: unknown module http"),
		},
		// --- imports are rejected in every nested block
		{
			Name:  "import_in_loop_body",
			Input: "for i range [0..2] {\n    ak.import(\"fs\")\n}",
			Err:   errors.New("2: error[import-scope]: imports are allowed only in the root block"),
		},
		{
			Name:  "import_in_else_block",
			Input: "if false {\n} else {\n    ak.import(\"fs\")\n}",
			Err:   errors.New("3: error[import-scope]: imports are allowed only in the root block"),
		},
		// --- the module name must be a plain string literal
		{
			Name:  "import_from_template",
			Input: "let x = \"fs\"\nak.import(\"${x}\")",
			Err:   errors.New("2: error[import-arg]: import takes a module name as a string literal"),
		},
		// --- module names stay reserved in every scope
		{
			Name:  "shadow_module_in_block",
			Input: "ak.import(\"fs\")\nif true {\n    let fs = 1\n}",
			Err:   errors.New("3: error[module-name]: fs is a module name"),
		},
		{
			Name:  "loop_variable_named_ak",
			Input: "for ak in [1] {\n}",
			Err:   errors.New("1: error[module-name]: ak is a module name"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateCheck(t, c)
		})
	}
}

func TestImportValid(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- a root import is enough
		{
			Name:     "import_only",
			Input:    "ak.import(\"fs\")\nreturn 1",
			Expected: "1",
		},
		{
			Name:     "core_calls_before_import",
			Input:    "ak.log(\"start\")\nak.import(\"fs\")\nfs.write(\"a.txt\", \"b\")\nreturn fs.read(\"a.txt\")",
			Expected: `"b"`,
		},
		{
			Name:     "setup_after_import",
			Input:    "ak.import(\"fs\")\nak.setup(log=\"info\")\nreturn fs.exists(\"a.txt\")",
			Expected: "false",
		},
		// --- imported modules are usable in nested blocks
		{
			Name:     "imported_in_if_cond",
			Input:    "ak.import(\"fs\")\nif fs.exists(\"a.txt\") {\n    return \"yes\"\n}\nreturn \"no\"",
			Expected: `"no"`,
		},
		{
			Name:     "imported_in_else_if",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"x\")\nif false {\n    return \"first\"\n} else if fs.exists(\"a.txt\") {\n    return \"second\"\n}\nreturn \"none\"",
			Expected: `"second"`,
		},
		{
			Name:     "imported_in_for_in",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"x\")\nvar n = 0\nfor f in [\"a.txt\", \"b.txt\"] {\n    if fs.exists(f) {\n        n = n + 1\n    }\n}\nreturn n",
			Expected: "1",
		},
		{
			Name:     "imported_in_nested_loops",
			Input:    "ak.import(\"fs\")\nfor i range [0..2] {\n    for j range [0..2] {\n        fs.write(\"a.txt\", \"${i}${j}\")\n    }\n}\nreturn fs.read(\"a.txt\")",
			Expected: `"00011011"`,
		},
		// --- module args can come from script values
		{
			Name:     "let_arg",
			Input:    "ak.import(\"fs\")\nlet p = \"a.txt\"\nreturn fs.exists(p)",
			Expected: "false",
		},
		{
			Name:     "var_arg",
			Input:    "ak.import(\"fs\")\nvar p = \"a.txt\"\np = \"b.txt\"\nfs.write(p, \"x\")\nreturn fs.exists(\"b.txt\")",
			Expected: "true",
		},
		{
			Name:     "object_member_arg",
			Input:    "ak.import(\"fs\")\nlet cfg = {path: \"a.txt\", body: \"hi\"}\nfs.write(cfg.path, cfg.body)\nreturn fs.read(cfg.path)",
			Expected: `"hi"`,
		},
		{
			Name:     "call_result_as_arg",
			Input:    "ak.import(\"fs\")\nak.log(fs.exists(\"a.txt\"))\nreturn fs.exists(\"a.txt\")",
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

func TestImportOutput(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- module results are visible in the log output
		{
			Name:     "log_read_content",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"hello\")\nak.log(fs.read(\"a.txt\"))\nreturn 1",
			Expected: "level=INFO msg=hello",
		},
		{
			Name:     "log_exists_result",
			Input:    "ak.import(\"fs\")\nak.debug(fs.exists(\"a.txt\"))\nreturn 1",
			Expected: "level=DEBUG msg=false",
		},
		{
			Name:     "log_in_loop",
			Input:    "ak.import(\"fs\")\nfor i range [0..3] {\n    fs.write(\"a.txt\", \"line ${i}\")\n \n}\nak.log(fs.read(\"a.txt\"))\nreturn 1",
			Expected: "level=INFO msg=\"line 0line 1line 2\"",
		},
	}

	for _, c := range cases {
		var buf bytes.Buffer
		t.Run(c.Name, func(t *testing.T) {
			reg, err := std.New(
				t.TempDir(),
				slog.New(slog.NewTextHandler(
					&buf, &slog.HandlerOptions{Level: slog.LevelDebug},
				)),
			)
			if err != nil {
				t.Fatalf("modules: %v", err)
			}

			out, err := runner.New(eval.Options{Modules: reg}).Run(t.Context(), c.Input, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := out.String(); got != "1" {
				t.Errorf("expected return: 1, got: %s", got)
			}
			if got := buf.String(); !strings.Contains(got, c.Expected) {
				t.Errorf("expected: %s, got: %s", c.Expected, got)
			}
		})
	}
}

func TestImportParse(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- import signature is checked while parsing
		{
			Name:  "import_number",
			Input: "ak.import(1)",
			Err:   errors.New("1:10: error[arg-kind]: argument name of ak.import must be a string"),
		},
		{
			Name:  "import_two_names",
			Input: "ak.import(\"fs\", \"ak\")",
			Err:   errors.New("1:10: error[arity]: wrong number of arguments for ak.import, want 1, got 2"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline.ValidateParse(t, c)
		})
	}
}

func TestImportRun(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- imported modules work inside nested blocks
		{
			Name:     "write_in_loop",
			Input:    "ak.import(\"fs\")\nfor i range [0..3] {\n    fs.write(\"a.txt\", \"${i}\")\n}\nreturn fs.read(\"a.txt\")",
			Expected: `"012"`,
		},
		{
			Name:     "core_and_imported_module",
			Input:    "ak.import(\"fs\")\nak.log(\"start\")\nreturn fs.exists(\"x.txt\")",
			Expected: "false",
		},
		// --- module results flow into script values
		{
			Name:     "read_variable_path",
			Input:    "ak.import(\"fs\")\nlet p = \"a.txt\"\nfs.write(p, \"x\")\nreturn fs.read(p)",
			Expected: `"x"`,
		},
		{
			Name:     "overwrite_file",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"a\")\nfs.write(\"a.txt\", \"b\")\nreturn fs.read(\"a.txt\")",
			Expected: `"ab"`,
		},
		{
			Name:     "exists_in_if",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"x\")\nif fs.exists(\"a.txt\") {\n    return \"yes\"\n}\nreturn \"no\"",
			Expected: `"yes"`,
		},
		{
			Name:     "concat_reads",
			Input:    "ak.import(\"fs\")\nfs.write(\"a.txt\", \"a\")\nfs.write(\"b.txt\", \"b\")\nreturn fs.read(\"a.txt\") + fs.read(\"b.txt\")",
			Expected: `"ab"`,
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

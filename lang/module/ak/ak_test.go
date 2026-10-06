package ak_test

import (
	"bytes"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
	"github.com/siper92/akha/lang/module/ak"
	"github.com/siper92/akha/lang/module/std"
	"github.com/siper92/akha/lang/runner"
	"github.com/siper92/akha/lang/tests_utils"
	"github.com/siper92/akha/lang/tests_utils/pipeline_test"
)

func TestAkFuncs(t *testing.T) {
	m, err := ak.New(slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("new: %v", err)
	}

	cases := []tests_utils.Case[string, bool]{
		// --- only import is a loader
		{
			Name:     "import_is_loader",
			Input:    ak.Import,
			Expected: true,
		},
		{
			Name:     "log_is_not_loader",
			Input:    ak.Log,
			Expected: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			f, ok := m.Func(c.Input)
			if !ok {
				t.Fatalf("missing func %s", c.Input)
			}
			if _, got := f.(module.ILoader); got != c.Expected {
				t.Errorf("expected: %v, got: %v", c.Expected, got)
			}
		})
	}

	var names []string
	for _, f := range m.Funcs() {
		names = append(names, f.Name())
	}
	if want := []string{ak.Import, ak.Setup, ak.Log, ak.Debug}; !slices.Equal(names, want) {
		t.Errorf("expected: %v, got: %v", want, names)
	}
}

func TestAkSignatures(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- ak signatures are checked while parsing
		{
			Name:  "setup_unknown_kwarg",
			Input: "ak.setup(level=\"info\")",
			Err:   errors.New("1:9: error[unknown-kwarg]: unknown named argument level for ak.setup"),
		},
		{
			Name:  "import_without_name",
			Input: "ak.import()",
			Err:   errors.New("1:10: error[arity]: wrong number of arguments for ak.import, want 1, got 0"),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateParse(t, c)
		})
	}
}

func TestAkImport(t *testing.T) {
	cases := []tests_utils.Case[string, struct{}]{
		// --- ak is always imported
		{
			Name:  "import_ak",
			Input: "ak.import(\"ak\")",
			Err:   errors.New("1: error[import-dup]: module ak is already imported"),
		},
		{
			Name:  "setup_without_import",
			Input: "ak.setup(log=\"info\", debug=\"x\")",
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			pipeline_test.ValidateCheck(t, c)
		})
	}
}

func TestAkLog(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		// --- log and debug write to the module logger
		{
			Name:     "log_info",
			Input:    "ak.log(\"hi\")",
			Expected: "level=INFO msg=hi",
		},
		{
			Name:     "log_debug",
			Input:    "ak.debug(\"hi\")",
			Expected: "level=DEBUG msg=hi",
		},
		// --- non string values are logged in their string form
		{
			Name:     "log_number",
			Input:    "ak.log(42)",
			Expected: "level=INFO msg=42",
		},
		{
			Name:     "log_bool",
			Input:    "ak.log(true)",
			Expected: "level=INFO msg=true",
		},
		{
			Name:     "log_variable",
			Input:    "let n = 1 + 2\nak.debug(n)",
			Expected: "level=DEBUG msg=3",
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

			if _, err := runner.New(eval.Options{Modules: reg}).Run(t.Context(), c.Input, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got := buf.String(); !strings.Contains(got, c.Expected) {
				t.Errorf("expected: %s, got: %s", c.Expected, got)
			}
		})
	}
}

package pipeline

import (
	"errors"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/check"
	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/parser"
	"github.com/siper92/akha/lang/runner"
	"github.com/siper92/akha/lang/tests_utils"
)

func ValidateParse(t *testing.T, c tests_utils.Case[string, string]) {
	t.Helper()

	script, err := parser.New("", c.Input).Parse()
	if c.Err != nil {
		if err == nil || !strings.Contains(err.Error(), c.Err.Error()) {
			t.Fatalf("expected error: %v, got: %v", c.Err, err)
		}

		return
	}

	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	if got := script.String(); got != c.Expected {
		t.Errorf("expected: %s, got: %s", c.Expected, got)
	}
}

func ValidateCheck(t *testing.T, c tests_utils.Case[string, struct{}]) {
	t.Helper()

	script, err := parser.New("", c.Input).Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	err = check.New("").Check(t.Context(), script)
	if c.Err == nil {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		return
	}

	if !errors.Is(err, check.ErrCheck) || !strings.Contains(err.Error(), c.Err.Error()) {
		t.Fatalf("expected error: %v, got: %v", c.Err, err)
	}
}

func ValidateRun(t *testing.T, c tests_utils.Case[string, string], input eval.Value) {
	t.Helper()

	out, err := runner.New(eval.Options{MaxStmts: 10_000, MaxIterations: 1_000}).Run(t.Context(), c.Input, input)
	if c.Err != nil {
		if err == nil || !strings.Contains(err.Error(), c.Err.Error()) {
			t.Fatalf("expected error: %v, got: %v", c.Err, err)
		}

		return
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := out.String(); got != c.Expected {
		t.Errorf("expected: %s, got: %s", c.Expected, got)
	}
}

package parser_test

import (
	"testing"

	"github.com/siper92/akha/lang/tests"
)

func TestScripts(t *testing.T) {
	// --- script: whole scripts print in canonical ebnf form
	tests.RunAst(t, []tests.AstCase{
		{
			Name: "sample",
			Src:  tests.Sample,
			Want: tests.SampleCanonical,
		},
		{
			Name: "sample_crlf",
			Src:  tests.SampleCRLF,
			Want: tests.SampleCRLFCanonical,
		},
		{
			Name: "canonical_is_fixed_point",
			Src:  tests.SampleCanonical,
			Want: tests.SampleCanonical,
		},
	})

	// --- spec by example
	tests.RunOK(t, []tests.ParseOkCase{
		{
			Name: "spec_def",
			Src:  tests.SpecDef,
		},
	})

	tests.RunSame(t, []tests.ParseSameCase{
		{
			Name: "spec_def_is_stable",
			Src:  tests.SpecDef,
			Same: tests.SpecDef,
		},
	})
}

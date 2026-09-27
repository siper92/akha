package parser_test

import (
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestScripts(t *testing.T) {
	// --- script: whole scripts print in canonical ebnf form
	tests_utils.RunAst(t, []tests_utils.AstCase{
		{
			Name: "sample",
			Src:  tests_utils.Sample,
			Want: tests_utils.SampleCanonical,
		},
		{
			Name: "sample_crlf",
			Src:  tests_utils.SampleCRLF,
			Want: tests_utils.SampleCRLFCanonical,
		},
		{
			Name: "canonical_is_fixed_point",
			Src:  tests_utils.SampleCanonical,
			Want: tests_utils.SampleCanonical,
		},
	})

	// --- spec by example
	tests_utils.RunOK(t, []tests_utils.ParseOkCase{
		{
			Name: "spec_def",
			Src:  tests_utils.SpecDef,
		},
	})

	tests_utils.RunSame(t, []tests_utils.ParseSameCase{
		{
			Name: "spec_def_is_stable",
			Src:  tests_utils.SpecDef,
			Same: tests_utils.SpecDef,
		},
	})
}

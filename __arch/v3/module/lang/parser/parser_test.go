package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestParserFullScripts(t *testing.T) {
	cases := []tests_utils.Case[string, string]{
		{
			Name:     "simple canonical sample",
			Input:    Sample,
			Expected: SampleCanonical,
		},
		{
			Name:     "simple canonical sample with CRLF",
			Input:    SampleCRLF,
			Expected: SampleCRLFCanonical,
		},
		{
			Name:     "akha spec def file",
			Input:    "file://testdata/spec_def.ak",
			Expected: "file://testdata/spec_def.canonical.ak",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			validateParserOutput(t, testCase)
		})
	}
}

func validateParserOutput(t *testing.T, c tests_utils.Case[string, string]) {
	t.Helper()
	var stringParser Parser
	if filePath, ok := strings.CutPrefix(c.Input, "file://"); ok {
		stringParser = New(filePath, "")
	} else {
		stringParser = New("", c.Input)
	}

	script, err := stringParser.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var expected = c.Expected
	var expectedContent []byte
	if filePath, ok := strings.CutPrefix(expected, "file://"); ok {
		expectedContent, err = os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read expected file: %v", err)
		}

		expected = string(expectedContent)
	}

	got := script.String()
	if got != expected {
		t.Errorf("Expected: %s, got: %s", expected, got)
	}
}

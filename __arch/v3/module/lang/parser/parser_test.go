package parser

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/lexer"
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

func TestParserErrors(t *testing.T) {
	cases, err := extractErrors()
	if err != nil {
		t.Fatalf("Failed to load errors: %v", err)
	}

	for _, testCase := range cases {
		t.Run(testCase.Name, func(t *testing.T) {
			validateParserOutput(t, testCase)
		})
	}
}

func extractErrors() ([]tests_utils.Case[string, string], error) {
	cases := []tests_utils.Case[string, string]{}

	text, err := os.ReadFile("testdata/errors.ak")
	if err != nil {
		return nil, err
	}

	header := `let a = 1
var b
b = a + 1`

	count := 1
	for part := range strings.SplitSeq(string(text), "---") {
		caseText := header + "\n" + strings.TrimSpace(part)
		cases = append(cases, tests_utils.Case[string, string]{
			Name:  fmt.Sprintf("error_%d", count),
			Input: caseText,
			Err:   extractErrFromCase(part),
		})
		count++
	}

	return cases, nil
}

func extractErrFromCase(text string) error {
	errText := ""
	for textLine := range strings.SplitSeq(text, "\n") {
		if _, after, ok := strings.Cut(textLine, "// ###"); ok {
			errText = strings.TrimSpace(after)
		}
	}

	return errors.New(errText)
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
		if _, ok := errors.AsType[*lexer.Error](err); ok {
			if c.Err != nil {
				if strings.Contains(err.Error(), c.Err.Error()) {
					return
				}

				t.Fatalf("expected error: %v, got: %v", c.Err, err)
			}
		}

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

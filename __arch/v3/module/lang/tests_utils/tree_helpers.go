package tests_utils

import (
	"errors"
	"strings"
	"testing"

	"github.com/siper92/akha/lang/diag"
)

const NestingMsg = "excessive nesting"

type TreeCase struct {
	Name string
	Src  string
	Want string
}

func StripPos(err error) string {
	var de *diag.Error
	if errors.As(err, &de) {
		return de.Msg
	}
	s := err.Error()
	if i := strings.Index(s, ": "); i >= 0 {
		return s[i+len(": "):]
	}

	return s
}

func TreeOf(n any, err error) string {
	if err != nil {
		return StripPos(err)
	}

	return Productions(n)
}

func RunTree(t *testing.T, cases []TreeCase) {
	t.Helper()
	runTree(t, cases, func(src string) (string, error) {
		s, err := Parse(ParseFile, src)
		return TreeOf(s, err), err
	})
}

func RunExprTree(t *testing.T, cases []TreeCase) {
	t.Helper()
	runTree(t, cases, func(src string) (string, error) {
		x, err := ParseExpr(src)
		return TreeOf(x, err), err
	})
}

func runTree(t *testing.T, cases []TreeCase, tree func(string) (string, error)) {
	t.Helper()
	defined := make(map[string]bool)
	for _, r := range ReadGrammar(t, GrammarFile) {
		defined[r.Name] = true
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, err := tree(c.Src)
			if got != c.Want {
				t.Fatalf("tree of %q\n got: %s\nwant: %s", c.Src, got, c.Want)
			}
			if err != nil {
				return
			}
			for _, name := range ProductionNames(c.Want) {
				if !defined[name] {
					t.Fatalf("production %s is not in %s", name, GrammarFile)
				}
			}
		})
	}
}

func RunNesting(t *testing.T, cases []ParseOkCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := Parse(ParseFile, c.Src)
			if err == nil {
				t.Fatalf("expected %q, got none", NestingMsg)
			}
			if de := AsDiag(t, err); de.Code != diag.CodeNesting || !strings.Contains(de.Msg, NestingMsg) {
				t.Fatalf("nesting\n got: %v\nwant: %s", err, NestingMsg)
			}
		})
	}
}

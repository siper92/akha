package tests

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/siper92/akha/lang/token"
)

const LexHintSemicolon = "one statement per line"

var LexKeywords = []string{
	"let", "var", "if", "else", "for", "in", "range",
	"break", "continue", "return", "exit",
	"and", "or", "not",
	"true", "false", "null",
}

var LexReserved = []string{"fn", "try", "catch"}

type LexOKCase struct {
	Name string
	Src  string
}

type LexSameCase struct {
	Name string
	Src  string
	Ref  string
}

type LexErrCase struct {
	Name string
	Src  string
	Line int
	Col  int
}

func LexTokenize(t *testing.T, src string) []token.Token {
	t.Helper()
	toks, err := Tokenize(src)
	if err != nil {
		t.Fatalf("tokenize %q: %v", src, err)
	}

	return toks
}

func LexErrPosRe(line, col int) *regexp.Regexp {
	c := `\d+`
	if col > 0 {
		c = strconv.Itoa(col)
	}

	return regexp.MustCompile(fmt.Sprintf(`(^|[^0-9])%d:%s([^0-9]|$)`, line, c))
}

func RunLexOK(t *testing.T, cases []LexOKCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			LexTokenize(t, c.Src)
		})
	}
}

func RunLexCount(t *testing.T, cases []LexSameCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, want := len(LexTokenize(t, c.Src)), len(LexTokenize(t, c.Ref))
			if got != want {
				t.Fatalf("tokens of %q\n got: %d\nwant: %d (same as %q)", c.Src, got, want, c.Ref)
			}
		})
	}
}

func RunLexSame(t *testing.T, cases []LexSameCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, want := LexTokenize(t, c.Src), LexTokenize(t, c.Ref)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("tokens of %q\n got: %v\nwant: %v", c.Src, got, want)
			}
		})
	}
}

func RunLexErr(t *testing.T, cases []LexErrCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := Tokenize(c.Src)
			if err == nil {
				t.Fatalf("expected lex error at %d:%d, got none", c.Line, c.Col)
			}

			de := AsDiag(t, err)
			if de.Code == "" {
				t.Fatalf("expected an error code, got none: %v", err)
			}

			if !LexErrPosRe(c.Line, c.Col).MatchString(err.Error()) {
				t.Fatalf("pos\n got: %v\nwant: %d:%d", err, c.Line, c.Col)
			}
		})
	}
}

func RunLexHint(t *testing.T, cases []LexOKCase, hint string) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			_, err := Tokenize(c.Src)
			if err == nil {
				t.Fatalf("expected lex error, got none")
			}

			AsDiag(t, err)
			if !regexp.MustCompile(regexp.QuoteMeta(hint)).MatchString(err.Error()) {
				t.Fatalf("hint\n got: %v\nwant: %s", err, hint)
			}
		})
	}
}

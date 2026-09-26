package tests

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

const ParseFile = "t.ak"

type ParseOkCase struct {
	Name string
	Src  string
}

type ParseSameCase struct {
	Name string
	Src  string
	Same string
}

type ParseErrCase struct {
	Name string
	Src  string
	Line int
	Msg  string
}

type PosCase struct {
	Name  string
	Src   string
	Lines []int
}

func Show(n any) string {
	return fmt.Sprint(n)
}

func MustParse(t *testing.T, src string) string {
	t.Helper()
	s, err := Parse(ParseFile, src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}

	return Show(s)
}

func MustParseExpr(t *testing.T, src string) string {
	t.Helper()
	e, err := ParseExpr(src)
	if err != nil {
		t.Fatalf("parse expr %q: %v", src, err)
	}

	return Show(e)
}

func MustCanonical(t *testing.T, src string, parse func(*testing.T, string) string) string {
	t.Helper()
	got := parse(t, src)
	if again := parse(t, got); again != got {
		t.Fatalf("not canonical ebnf\n got: %s\nagain: %s", got, again)
	}

	return got
}

func RunAst(t *testing.T, cases []AstCase) {
	t.Helper()
	runAst(t, cases, MustParse)
}

func RunExprAst(t *testing.T, cases []AstCase) {
	t.Helper()
	runAst(t, cases, MustParseExpr)
}

func runAst(t *testing.T, cases []AstCase, parse func(*testing.T, string) string) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if got := MustCanonical(t, c.Src, parse); got != c.Want {
				t.Fatalf("ast\n got: %s\nwant: %s", got, c.Want)
			}
		})
	}
}

func RunOK(t *testing.T, cases []ParseOkCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			MustCanonical(t, c.Src, MustParse)
		})
	}
}

func RunExprOK(t *testing.T, cases []ParseOkCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			MustCanonical(t, c.Src, MustParseExpr)
		})
	}
}

func RunSame(t *testing.T, cases []ParseSameCase) {
	t.Helper()
	runCompare(t, cases, MustParse, true)
}

func RunSameExpr(t *testing.T, cases []ParseSameCase) {
	t.Helper()
	runCompare(t, cases, MustParseExpr, true)
}

func RunDiff(t *testing.T, cases []ParseSameCase) {
	t.Helper()
	runCompare(t, cases, MustParse, false)
}

func RunDiffExpr(t *testing.T, cases []ParseSameCase) {
	t.Helper()
	runCompare(t, cases, MustParseExpr, false)
}

func runCompare(t *testing.T, cases []ParseSameCase, parse func(*testing.T, string) string, same bool) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got, other := MustCanonical(t, c.Src, parse), MustCanonical(t, c.Same, parse)
			if same && got != other {
				t.Fatalf("not same\n got: %s\nwant: %s", got, other)
			}
			if !same && got == other {
				t.Fatalf("expected different ast, both: %s", got)
			}
		})
	}
}

func RunParseErr(t *testing.T, cases []ParseErrCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			s, err := Parse(ParseFile, c.Src)
			if err == nil {
				t.Fatalf("expected error for %q, got ast %s", c.Src, Show(s))
			}
			de := AsDiag(t, err)
			msg := err.Error()
			if de.Code == "" {
				t.Fatalf("empty code: %s", msg)
			}
			if !strings.Contains(msg, "error["+de.Code+"]") {
				t.Fatalf("format\n got: %s\nwant: error[%s]", msg, de.Code)
			}
			if c.Line > 0 {
				prefix := fmt.Sprintf("%s:%d:", ParseFile, c.Line)
				if !strings.HasPrefix(msg, prefix) {
					t.Fatalf("line\n got: %s\nwant prefix: %s", msg, prefix)
				}
			}
			if c.Msg != "" && !strings.Contains(msg, c.Msg) {
				t.Fatalf("msg\n got: %s\nwant: %s", msg, c.Msg)
			}
		})
	}
}

func ParseNode(t *testing.T, src string) any {
	t.Helper()
	s, err := Parse(ParseFile, src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	return s
}

func derefValue(v reflect.Value) reflect.Value {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}

	return v
}

func PosLine(t *testing.T, pos reflect.Value) int {
	t.Helper()
	switch pos.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int(pos.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int(pos.Uint())
	case reflect.Struct:
		for _, name := range []string{"Col", "Column"} {
			if pos.FieldByName(name).IsValid() {
				t.Fatalf("ast Pos %s has a %s field, want line only", pos.Type(), name)
			}
		}
		l := pos.FieldByName("Line")
		if !l.IsValid() {
			t.Fatalf("ast Pos %s has no Line field", pos.Type())
		}

		return PosLine(t, l)
	}
	t.Fatalf("unsupported Pos kind %s", pos.Kind())

	return 0
}

func WalkPos(t *testing.T, v reflect.Value, fn func(line int)) {
	t.Helper()
	v = derefValue(v)
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			WalkPos(t, v.Index(i), fn)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if v.Type().Field(i).Name == "Pos" {
				fn(PosLine(t, v.Field(i)))
				continue
			}
			WalkPos(t, v.Field(i), fn)
		}
	}
}

func TopLines(t *testing.T, s any) []int {
	t.Helper()
	v := derefValue(reflect.ValueOf(s))
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		if f.Kind() != reflect.Slice {
			continue
		}
		lines := make([]int, 0, f.Len())
		for j := 0; j < f.Len(); j++ {
			n := derefValue(f.Index(j))
			p := n.FieldByName("Pos")
			if !p.IsValid() {
				t.Fatalf("stmt %s has no Pos field", n.Type())
			}
			lines = append(lines, PosLine(t, p))
		}

		return lines
	}
	t.Fatalf("script %s has no statement slice", v.Type())

	return nil
}

func RunPos(t *testing.T, cases []PosCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			got := TopLines(t, ParseNode(t, c.Src))
			if fmt.Sprint(got) != fmt.Sprint(c.Lines) {
				t.Fatalf("lines\n got: %v\nwant: %v", got, c.Lines)
			}
		})
	}
}

func RunPosLineOnly(t *testing.T, cases []PosCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			last := strings.Count(c.Src, "\n") + 1
			n := 0
			WalkPos(t, reflect.ValueOf(ParseNode(t, c.Src)), func(line int) {
				n++
				if line < 1 || line > last {
					t.Fatalf("line %d out of range 1..%d", line, last)
				}
			})
			if n == 0 {
				t.Fatalf("no Pos fields found")
			}
		})
	}
}

package parser_test

import (
	"os"
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
)

const specPath = "../lexer/testdata/spec.ak"

func specScript() *ast.Script {
	return script(
		call("Ak", "Allow", 2, 1, args(spread("FS", 2, 10)), nil),
		call("Ak", "Setup", 3, 1, nil, kwargs(
			kw("log", str("info.log", 4, 9), 4, 5),
			kw("debug", str("test.log", 5, 11), 5, 5),
		)),
		call("Ak", "Log", 7, 1, args(str("hello", 7, 8)), nil),
		call("Ak", "Debug", 8, 1, args(str("dbg", 8, 10), num("1", 8, 17), str("two", 8, 20)), nil),
		call("FS", "WriteFile", 10, 1, args(str("file.txt", 10, 14), str("content", 10, 26)), nil),
		call("FS", "ReadFile", 11, 1, args(str("file.txt", 11, 13)), nil),
		call("FS", "UpdateFile", 12, 1, args(str("file.txt", 12, 15), str("new content", 12, 27)), nil),
		call("FS", "ListFiles", 14, 1, args(str("path/", 14, 14)), nil),
		call("Ak", "Exit", 15, 1, args(str("done", 15, 9)), kwargs(kw("code", num("0", 15, 22), 15, 17))),
	)
}

func TestGoldenSpec(t *testing.T) {
	src, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}
	cases := []tu.Case[[]lexer.Option, result]{
		// --- spec.ak parses without errors in both lexer modes
		{
			Name:     "comments_dropped",
			Input:    nil,
			Expected: result{Script: specScript()},
		},
		{
			Name:     "comments_kept",
			Input:    []lexer.Option{lexer.WithComments()},
			Expected: result{Script: specScript()},
		},
	}
	tu.Run(tu.New(t), cases, func(opts []lexer.Option) (result, error) {
		return parseWith(opts...)(string(src))
	}, nil)
}

func TestGoldenSpecPerCall(t *testing.T) {
	src, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read %s: %v", specPath, err)
	}
	got, err := parseWith()(string(src))
	if err != nil {
		t.Fatalf("parse %s: %v", specPath, err)
	}
	want := specScript()
	if len(got.Script.Calls) != len(want.Calls) {
		t.Fatalf("want %d calls, got %d", len(want.Calls), len(got.Script.Calls))
	}
	cases := make([]tu.Case[int, *ast.Call], 0, len(want.Calls))
	for i, c := range want.Calls {
		cases = append(cases, tu.Case[int, *ast.Call]{
			Name:     c.Target.Module + "." + c.Target.Name,
			Input:    i,
			Expected: c,
		})
	}
	tu.Run(tu.New(t), cases, func(i int) (*ast.Call, error) {
		return got.Script.Calls[i], nil
	}, nil)
}

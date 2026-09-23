package parser_test

import (
	"os"
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/token"
)

const (
	specPath = "../lexer/testdata/spec.ak"
	varsPath = "../lexer/testdata/vars.ak"
)

func specScript() *ast.Script {
	return script(
		callStmt("Ak", "Allow", 2, 1, args(spread("FS", 2, 10)), nil),
		callStmt("Ak", "Setup", 3, 1, nil, kwargs(
			kw("log", str("info.log", 4, 9), 4, 5),
			kw("debug", str("test.log", 5, 11), 5, 5),
		)),
		callStmt("Ak", "Log", 7, 1, args(str("hello", 7, 8)), nil),
		callStmt("Ak", "Debug", 8, 1, args(str("dbg", 8, 10), num("1", 8, 17), str("two", 8, 20)), nil),
		callStmt("FS", "WriteFile", 10, 1, args(str("file.txt", 10, 14), str("content", 10, 26)), nil),
		callStmt("FS", "ReadFile", 11, 1, args(str("file.txt", 11, 13)), nil),
		callStmt("FS", "UpdateFile", 12, 1, args(str("file.txt", 12, 15), str("new content", 12, 27)), nil),
		callStmt("FS", "ListFiles", 14, 1, args(str("path/", 14, 14)), nil),
		callStmt("Ak", "Exit", 15, 1, args(str("done", 15, 9)), kwargs(kw("code", num("0", 15, 22), 15, 17))),
	)
}

func varsScript() *ast.Script {
	return script(
		callStmt("Ak", "Allow", 2, 1, args(spread("FS", 2, 10)), nil),
		let("name", str("world", 3, 12), 3, 1),
		assign("name", str("akha", 4, 8), 4, 1),
		let("content", call("FS", "ReadFile", 5, 15, args(str("f.txt", 5, 27)), nil), 5, 1),
		let("n", binary(token.PLUS,
			num("1", 6, 9),
			binary(token.STAR, num("2", 6, 13), num("3", 6, 17), 6, 15),
			6, 11), 6, 1),
		let("s", binary(token.PLUS, str("a", 7, 9), str("b", 7, 15), 7, 13), 7, 1),
		let("neg", binary(token.PERCENT,
			binary(token.SLASH,
				unary(token.MINUS, binary(token.MINUS, ident("n", 8, 13), num("10", 8, 17), 8, 15), 8, 11),
				num("2", 8, 23),
				8, 21),
			num("3", 8, 27),
			8, 25), 8, 1),
		let("ok", binary(token.OR,
			binary(token.AND,
				binary(token.GT, ident("n", 9, 10), num("3", 9, 14), 9, 12),
				binary(token.NEQ, ident("name", 9, 20), str("", 9, 28), 9, 25),
				9, 16),
			unary(token.NOT, binary(token.EQ, ident("s", 9, 39), str("ab", 9, 44), 9, 41), 9, 34),
			9, 31), 9, 1),
		let("le", binary(token.LTE, ident("n", 10, 10), num("7", 10, 15), 10, 12), 10, 1),
		let("ge", binary(token.GTE, ident("n", 11, 10), num("7", 11, 15), 11, 12), 11, 1),
		let("lt", binary(token.LT, ident("n", 12, 10), num("7", 12, 14), 12, 12), 12, 1),
		callStmt("Ak", "Log", 13, 1, args(ident("s", 13, 8)), nil),
		callStmt("Ak", "Exit", 14, 1, args(str("done", 14, 9)), kwargs(kw("code", ident("n", 14, 22), 14, 17))),
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

func TestGoldenVars(t *testing.T) {
	src, err := os.ReadFile(varsPath)
	if err != nil {
		t.Fatalf("read %s: %v", varsPath, err)
	}
	cases := []tu.Case[[]lexer.Option, result]{
		// --- vars.ak parses without errors in both lexer modes
		{
			Name:     "comments_dropped",
			Input:    nil,
			Expected: result{Script: varsScript()},
		},
		{
			Name:     "comments_kept",
			Input:    []lexer.Option{lexer.WithComments()},
			Expected: result{Script: varsScript()},
		},
	}

	tu.Run(tu.New(t), cases, func(opts []lexer.Option) (result, error) {
		return parseWith(opts...)(string(src))
	}, nil)
}

func TestGoldenPerStmt(t *testing.T) {
	files := map[string]func() *ast.Script{
		specPath: specScript,
		varsPath: varsScript,
	}
	for path, want := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		got, err := parseWith()(string(src))
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		wantStmts := want().Stmts
		if len(got.Script.Stmts) != len(wantStmts) {
			t.Fatalf("%s: want %d statements, got %d", path, len(wantStmts), len(got.Script.Stmts))
		}
		cases := make([]tu.Case[int, ast.Stmt], 0, len(wantStmts))
		for i, st := range wantStmts {
			cases = append(cases, tu.Case[int, ast.Stmt]{
				Name:     stmtName(st),
				Input:    i,
				Expected: st,
			})
		}
		tu.Run(tu.New(t), cases, func(i int) (ast.Stmt, error) {
			return got.Script.Stmts[i], nil
		}, nil)
	}
}

func stmtName(st ast.Stmt) string {
	switch s := st.(type) {
	case *ast.CallStmt:
		return s.Call.Target.Module + "." + s.Call.Target.Name
	case *ast.Let:
		return "let_" + s.Name
	case *ast.Assign:
		return "assign_" + s.Name
	}
	return "stmt"
}

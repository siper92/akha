package parser_test

import (
	"errors"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/parser"
	"github.com/siper92/akha/lang/token"
)

type result struct {
	Script *ast.Script
	Errs   parser.Errors
}

func parseWith(opts ...lexer.Option) func(string) (result, error) {
	return func(src string) (result, error) {
		s, err := parser.New(lexer.New(src, opts...)).Parse()
		if err == nil {
			return result{Script: s}, nil
		}
		var errs parser.Errors
		if !errors.As(err, &errs) {
			return result{}, err
		}
		return result{Script: s, Errs: errs}, nil
	}
}

func pos(l, c int) token.Pos {
	return token.Pos{Line: l, Col: c}
}

func str(v string, l, c int) ast.Expr {
	return &ast.Literal{Kind: token.STRING, Value: v, P: pos(l, c)}
}

func num(v string, l, c int) ast.Expr {
	return &ast.Literal{Kind: token.INT, Value: v, P: pos(l, c)}
}

func ident(v string, l, c int) ast.Expr {
	return &ast.Literal{Kind: token.IDENT, Value: v, P: pos(l, c)}
}

func spread(m string, l, c int) ast.Expr {
	return &ast.Spread{Module: m, P: pos(l, c)}
}

func kw(name string, v ast.Expr, l, c int) ast.Kwarg {
	return ast.Kwarg{Name: name, Value: v, P: pos(l, c)}
}

func args(e ...ast.Expr) []ast.Expr {
	return e
}

func kwargs(k ...ast.Kwarg) []ast.Kwarg {
	return k
}

func call(mod, name string, l, c int, a []ast.Expr, k []ast.Kwarg) *ast.Call {
	return &ast.Call{
		Target: ast.Selector{Module: mod, Name: name, P: pos(l, c)},
		Args:   a,
		Kwargs: k,
		P:      pos(l, c),
	}
}

func script(calls ...*ast.Call) *ast.Script {
	return &ast.Script{Calls: calls}
}

func perr(l, c int, msg string) parser.Error {
	return parser.Error{Pos: pos(l, c), Msg: msg}
}

func perrs(e ...parser.Error) parser.Errors {
	return e
}

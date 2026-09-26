package tests

import (
	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/parser"
	"github.com/siper92/akha/lang/token"
)

func Tokenize(src string) ([]token.Token, error) {
	return lexer.Tokenize(src)
}

func Parse(file, src string) (*ast.Script, error) {
	return parser.New(file, src).Parse()
}

func ParseExpr(src string) (ast.Expr, error) {
	return parser.New("", src).ParseExpr()
}

package parser

import (
	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/token"
)

type Parser interface {
	Parse() (*ast.Script, error)
}

type Error struct {
	Pos token.Pos
	Msg string
}

type Errors []Error

var (
	_ error = (*Error)(nil)
	_ error = (Errors)(nil)
)

func (e *Error) Error() string { return e.Msg }

func (es Errors) Error() string {
	if len(es) == 0 {
		return ""
	}
	return es[0].Msg
}

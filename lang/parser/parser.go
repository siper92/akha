package parser

import (
	"fmt"
	"strings"

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

func (e *Error) Error() string {
	return fmt.Sprintf("%d:%d: %s", e.Pos.Line, e.Pos.Col, e.Msg)
}

func (es Errors) Error() string {
	msgs := make([]string, len(es))
	for i := range es {
		msgs[i] = es[i].Error()
	}
	return strings.Join(msgs, "\n")
}

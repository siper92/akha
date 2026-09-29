package check

import (
	"errors"
	"fmt"
	"strings"

	"github.com/siper92/akha/lang/ast"
)

var ErrCheck = errors.New("check error")

var _ error = (*Error)(nil)

type Error struct {
	File string
	Line int
	Code string
	Msg  string
	Hint string
}

func newError(pos ast.Pos, code, hint, format string, args ...any) *Error {
	return &Error{Line: pos.Line, Code: code, Msg: fmt.Sprintf(format, args...), Hint: hint}
}

func (e *Error) Error() string {
	var b strings.Builder
	if e.File != "" {
		b.WriteString(e.File)
		b.WriteByte(':')
	}

	fmt.Fprintf(&b, "%d: error[%s]: %s", e.Line, e.Code, e.Msg)
	if e.Hint != "" {
		b.WriteString("\nhint: ")
		b.WriteString(e.Hint)
	}

	return b.String()
}

func (e *Error) Unwrap() error {
	return ErrCheck
}

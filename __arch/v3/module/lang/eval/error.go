package eval

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrRuntime      = errors.New("runtime error")
	errKindMismatch = errors.New("kind mismatch")
)

const (
	CodeKindMismatch = "kind-mismatch"
	CodeOrdering     = "ordering"
	CodeDivZero      = "div-zero"
	CodeIndexRange   = "index-range"
	CodeIndexKind    = "index-kind"
	CodeNonIntegral  = "non-integral"
	CodeNullAccess   = "null-access"
	CodeMemberAccess = "member-access"
	CodeNotIterable  = "not-iterable"
	CodeNumberRange  = "number-range"
	CodeImmutable    = "immutable"
	CodeUndeclared   = "undeclared"
	CodeRedeclared   = "redeclared"
	CodeNoCallable   = "no-callable"
	CodeLimit        = "limit"
	CodeCanceled     = "canceled"
	CodeInternal     = "internal"
)

var _ error = (*RuntimeError)(nil)

type RuntimeError struct {
	File string
	Line int
	Code string
	Msg  string
}

func newError(code, format string, args ...any) *RuntimeError {
	return &RuntimeError{Code: code, Msg: fmt.Sprintf(format, args...)}
}

func (e *RuntimeError) Error() string {
	var b strings.Builder
	if e.File != "" {
		b.WriteString(e.File)
		b.WriteByte(':')
	}
	fmt.Fprintf(&b, "%d: runtime error: %s", e.Line, e.Msg)
	return b.String()
}

func (e *RuntimeError) Unwrap() error {
	return ErrRuntime
}

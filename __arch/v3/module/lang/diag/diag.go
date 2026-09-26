package diag

import (
	"errors"
	"fmt"
	"strings"

	"github.com/siper92/akha/lang/token"
)

var (
	ErrLex   = errors.New("lex error")
	ErrParse = errors.New("parse error")
)

const (
	CodeInvalidUTF8        = "invalid-utf8"
	CodeSemicolon          = "semicolon"
	CodeSingleQuote        = "single-quote"
	CodeLeadingZero        = "leading-zero"
	CodeInvalidNumber      = "invalid-number"
	CodeInvalidEscape      = "invalid-escape"
	CodeUnterminatedString = "unterminated-string"
	CodeNewlineInString    = "newline-in-string"
	CodeEmptyInterp        = "empty-interp"
	CodeUnterminatedInterp = "unterminated-interp"
	CodeCompoundAssign     = "compound-assign"
	CodeUnexpectedChar     = "unexpected-char"

	CodeUnexpectedToken   = "unexpected-token"
	CodeExpectedExpr      = "expected-expr"
	CodeExpectedName      = "expected-name"
	CodeKeywordName       = "keyword-name"
	CodeReserved          = "reserved"
	CodeExpectedNewline   = "expected-newline"
	CodeBlockOpen         = "block-open"
	CodeBlockNewline      = "block-newline"
	CodeUnclosedBlock     = "unclosed-block"
	CodeElsePlacement     = "else-placement"
	CodeLetInit           = "let-init"
	CodeDiscardVar        = "discard-var"
	CodeForLet            = "for-let"
	CodeRangeVars         = "range-vars"
	CodeRangeSyntax       = "range-syntax"
	CodeChainedComparison = "chained-comparison"
	CodeChainedIn         = "chained-in"
	CodeUnusedValue       = "unused-value"
	CodeAssignTarget      = "assign-target"
	CodeDuplicateKey      = "duplicate-key"
	CodeObjectKey         = "object-key"
	CodeHeaderObject      = "header-object"
	CodeSlice             = "slice"
	CodeInterpExpr        = "interp-expr"
	CodeLoopControl       = "loop-control"
)

type Error struct {
	Kind error
	File string
	Pos  token.Pos
	Code string
	Msg  string
	Hint string
}

func New(kind error, pos token.Pos, code, msg, hint string) *Error {
	return &Error{Kind: kind, Pos: pos, Code: code, Msg: msg, Hint: hint}
}

func (e *Error) Error() string {
	var b strings.Builder
	if e.File != "" {
		b.WriteString(e.File)
		b.WriteByte(':')
	}
	fmt.Fprintf(&b, "%d:%d: error[%s]: %s", e.Pos.Line, e.Pos.Col, e.Code, e.Msg)
	if e.Hint != "" {
		b.WriteString("\nhint: ")
		b.WriteString(e.Hint)
	}
	return b.String()
}

func (e *Error) Unwrap() error {
	return e.Kind
}

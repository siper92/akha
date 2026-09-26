package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/siper92/akha/lang/diag"
	"github.com/siper92/akha/lang/token"
)

type Lexer interface {
	Next() (token.Token, error)
}

var _ Lexer = (*lexer)(nil)

type lexer struct {
	src  []rune
	off  int
	line int
	col  int
	err  error
}

var single = map[rune]token.Kind{
	'+': token.Plus,
	'-': token.Minus,
	'*': token.Star,
	'/': token.Slash,
	'%': token.Percent,
	'(': token.LParen,
	')': token.RParen,
	'[': token.LBracket,
	']': token.RBracket,
	'{': token.LBrace,
	'}': token.RBrace,
	',': token.Comma,
	':': token.Colon,
}

var escapes = map[rune]rune{
	'n':  '\n',
	't':  '\t',
	'r':  '\r',
	'"':  '"',
	'\\': '\\',
	'$':  '$',
}

const escapeHint = `valid escapes: \n \t \r \" \\ \$`

func New(src string) Lexer {
	src = strings.TrimPrefix(src, "﻿")
	src = strings.ReplaceAll(src, "\r\n", "\n")
	return newLexer(src, token.Pos{Line: 1, Col: 1})
}

func NewAt(src string, pos token.Pos) Lexer {
	return newLexer(src, pos)
}

func Tokenize(src string) ([]token.Token, error) {
	return Collect(New(src))
}

func Collect(l Lexer) ([]token.Token, error) {
	var toks []token.Token
	for {
		t, err := l.Next()
		if err != nil {
			return nil, err
		}
		toks = append(toks, t)
		if t.Kind == token.EOF {
			return toks, nil
		}
	}
}

func newLexer(src string, pos token.Pos) *lexer {
	l := &lexer{line: pos.Line, col: pos.Col}
	if bad, ok := invalidUTF8(src, pos); ok {
		l.err = diag.New(diag.ErrLex, bad, diag.CodeInvalidUTF8, "invalid UTF-8 in source", "")
		return l
	}
	l.src = []rune(src)
	return l
}

func invalidUTF8(src string, pos token.Pos) (token.Pos, bool) {
	for i := 0; i < len(src); {
		r, size := utf8.DecodeRuneInString(src[i:])
		if r == utf8.RuneError && size == 1 {
			return pos, true
		}
		if r == '\n' {
			pos.Line++
			pos.Col = 1
		} else {
			pos.Col++
		}
		i += size
	}
	return pos, false
}

func (l *lexer) Next() (token.Token, error) {
	if l.err != nil {
		return token.Token{}, l.err
	}
	t, err := l.scan()
	if err != nil {
		l.err = err
	}
	return t, err
}

func (l *lexer) scan() (token.Token, error) {
	l.skipSpace()
	pos := l.pos()
	if l.off >= len(l.src) {
		return l.tok(token.EOF, "", pos)
	}
	r := l.src[l.off]
	switch {
	case r == '\n':
		l.advance()
		return l.tok(token.Newline, "\n", pos)
	case isLetter(r):
		return l.ident(pos)
	case isDigit(r):
		return l.number(pos)
	case r == '"':
		return l.str(pos)
	}
	return l.operator(pos)
}

func (l *lexer) skipSpace() {
	for l.off < len(l.src) {
		switch r := l.src[l.off]; {
		case r == ' ' || r == '\t' || r == '\r':
			l.advance()
		case r == '/' && l.peek(1) == '/':
			for l.off < len(l.src) && l.src[l.off] != '\n' {
				l.advance()
			}
		default:
			return
		}
	}
}

func (l *lexer) ident(pos token.Pos) (token.Token, error) {
	start := l.off
	for l.off < len(l.src) && (isLetter(l.src[l.off]) || isDigit(l.src[l.off])) {
		l.advance()
	}
	lit := string(l.src[start:l.off])
	return l.tok(token.Lookup(lit), lit, pos)
}

func (l *lexer) number(pos token.Pos) (token.Token, error) {
	start := l.off
	if l.src[l.off] == '0' && isDigit(l.peek(1)) {
		return l.fail(pos, diag.CodeLeadingZero, "write 7 instead of 007", "numbers cannot have leading zeros")
	}
	l.digits()
	if l.peek(0) == '.' && isDigit(l.peek(1)) {
		l.advance()
		l.digits()
	}
	if isLetter(l.peek(0)) {
		return l.fail(pos, diag.CodeInvalidNumber, "", "invalid number literal")
	}
	return l.tok(token.Number, string(l.src[start:l.off]), pos)
}

func (l *lexer) digits() {
	for l.off < len(l.src) && isDigit(l.src[l.off]) {
		l.advance()
	}
}

func (l *lexer) str(pos token.Pos) (token.Token, error) {
	start := l.off
	l.advance()
	var text strings.Builder
	var parts []token.Part
	for l.off < len(l.src) {
		at := l.pos()
		switch r := l.src[l.off]; r {
		case '"':
			l.advance()
			if parts == nil {
				return l.tok(token.String, text.String(), pos)
			}
			if text.Len() > 0 {
				parts = append(parts, token.Part{Text: text.String()})
			}
			return token.Token{Kind: token.Template, Lit: string(l.src[start:l.off]), Pos: pos, Parts: parts}, nil
		case '\n':
			return l.fail(at, diag.CodeNewlineInString, `use \n for a line break`, "string cannot contain a raw newline")
		case '\\':
			l.advance()
			if l.off >= len(l.src) {
				return l.fail(pos, diag.CodeUnterminatedString, "", "unterminated string")
			}
			e, ok := escapes[l.src[l.off]]
			if !ok {
				return l.fail(at, diag.CodeInvalidEscape, escapeHint, fmt.Sprintf("unknown escape \\%c", l.src[l.off]))
			}
			l.advance()
			text.WriteRune(e)
		case '$':
			if l.peek(1) != '{' {
				text.WriteRune(l.advance())
				continue
			}
			if text.Len() > 0 {
				parts = append(parts, token.Part{Text: text.String()})
				text.Reset()
			}
			part, err := l.interp(at)
			if err != nil {
				return token.Token{}, err
			}
			parts = append(parts, part)
		default:
			text.WriteRune(l.advance())
		}
	}
	return l.fail(pos, diag.CodeUnterminatedString, "", "unterminated string")
}

func (l *lexer) interp(at token.Pos) (token.Part, error) {
	l.advance()
	l.advance()
	pos := l.pos()
	start := l.off
	depth := 0
	for l.off < len(l.src) {
		switch l.src[l.off] {
		case '\n':
			return token.Part{}, l.errAt(l.pos(), diag.CodeNewlineInString, `use \n for a line break`, "string cannot contain a raw newline")
		case '"':
			if err := l.skipString(at); err != nil {
				return token.Part{}, err
			}
			continue
		case '{':
			depth++
		case '}':
			if depth > 0 {
				depth--
				break
			}
			expr := string(l.src[start:l.off])
			l.advance()
			if strings.TrimSpace(expr) == "" {
				return token.Part{}, l.errAt(at, diag.CodeEmptyInterp, "", "empty ${ } in string")
			}
			return token.Part{Expr: expr, IsExpr: true, Pos: pos}, nil
		}
		l.advance()
	}
	return token.Part{}, l.errAt(at, diag.CodeUnterminatedInterp, "", "missing } to close ${")
}

func (l *lexer) skipString(at token.Pos) error {
	l.advance()
	for l.off < len(l.src) {
		switch l.src[l.off] {
		case '\n':
			return l.errAt(l.pos(), diag.CodeNewlineInString, `use \n for a line break`, "string cannot contain a raw newline")
		case '"':
			l.advance()
			return nil
		case '\\':
			l.advance()
			if l.off >= len(l.src) {
				return l.errAt(at, diag.CodeUnterminatedInterp, "", "missing } to close ${")
			}
		}
		l.advance()
	}
	return l.errAt(at, diag.CodeUnterminatedInterp, "", "missing } to close ${")
}

func (l *lexer) operator(pos token.Pos) (token.Token, error) {
	r := l.advance()
	next := l.peek(0)
	switch r {
	case '+', '-', '*', '/', '%':
		if next == '=' {
			return l.fail(pos, diag.CodeCompoundAssign, fmt.Sprintf("use x = x %c y", r), fmt.Sprintf("compound assignment %c= is not supported", r))
		}
	case '=':
		if next == '=' {
			l.advance()
			return l.tok(token.Eq, "==", pos)
		}
		return l.tok(token.Assign, "=", pos)
	case '!':
		if next == '=' {
			l.advance()
			return l.tok(token.NotEq, "!=", pos)
		}
		return l.tok(token.Not, "!", pos)
	case '<':
		if next == '=' {
			l.advance()
			return l.tok(token.LtEq, "<=", pos)
		}
		return l.tok(token.Lt, "<", pos)
	case '>':
		if next == '=' {
			l.advance()
			return l.tok(token.GtEq, ">=", pos)
		}
		return l.tok(token.Gt, ">", pos)
	case '&':
		if next == '&' {
			l.advance()
			return l.tok(token.And, "&&", pos)
		}
		return l.fail(pos, diag.CodeUnexpectedChar, "use and", `unexpected character "&"`)
	case '|':
		if next == '|' {
			l.advance()
			return l.tok(token.Or, "||", pos)
		}
		return l.fail(pos, diag.CodeUnexpectedChar, "use or", `unexpected character "|"`)
	case '.':
		if next == '.' {
			l.advance()
			return l.tok(token.DotDot, "..", pos)
		}
		return l.tok(token.Dot, ".", pos)
	case ';':
		return l.fail(pos, diag.CodeSemicolon, "one statement per line", `";" is not allowed`)
	case '\'':
		return l.fail(pos, diag.CodeSingleQuote, `use double quotes "..."`, "single quotes are not allowed")
	}
	if k, ok := single[r]; ok {
		return l.tok(k, string(r), pos)
	}
	return l.fail(pos, diag.CodeUnexpectedChar, "", fmt.Sprintf("unexpected character %q", r))
}

func (l *lexer) tok(kind token.Kind, lit string, pos token.Pos) (token.Token, error) {
	return token.Token{Kind: kind, Lit: lit, Pos: pos}, nil
}

func (l *lexer) fail(pos token.Pos, code, hint, msg string) (token.Token, error) {
	return token.Token{}, l.errAt(pos, code, hint, msg)
}

func (l *lexer) errAt(pos token.Pos, code, hint, msg string) error {
	return diag.New(diag.ErrLex, pos, code, msg, hint)
}

func (l *lexer) pos() token.Pos {
	return token.Pos{Line: l.line, Col: l.col}
}

func (l *lexer) peek(n int) rune {
	if l.off+n < len(l.src) {
		return l.src[l.off+n]
	}
	return 0
}

func (l *lexer) advance() rune {
	r := l.src[l.off]
	l.off++
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func isLetter(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r == '_'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

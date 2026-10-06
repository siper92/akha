package lexer

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Lexer interface {
	Next() (Token, error)
}

var _ Lexer = (*lexer)(nil)

const MaxDepth = 1000

type lexer struct {
	src   []rune
	off   int
	line  int
	col   int
	depth int
	err   error
}

var single = map[rune]Kind{
	'+': Plus,
	'-': Minus,
	'*': Star,
	'/': Slash,
	'%': Percent,
	'(': LParen,
	')': RParen,
	'[': LBracket,
	']': RBracket,
	'{': LBrace,
	'}': RBrace,
	',': Comma,
	':': Colon,
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

var LexKeywords = []string{
	"let", "var", "if", "else", "for", "in", "range",
	"break", "continue", "return", "exit",
	"and", "or", "not",
	"true", "false", "null",
}

func New(src string) Lexer {
	src = strings.TrimPrefix(src, "\xEF\xBB\xBF") // remove BOM
	src = strings.ReplaceAll(src, "\r\n", "\n")
	return newLexer(src, Pos{Line: 1, Col: 1})
}

func NewAt(src string, pos Pos) Lexer {
	return newLexer(src, pos)
}

func Tokenize(src string) ([]Token, error) {
	return Collect(New(src))
}

func Collect(l Lexer) ([]Token, error) {
	var toks []Token
	for {
		t, err := l.Next()
		if err != nil {
			return nil, err
		}

		toks = append(toks, t)
		if t.Kind == EOF {
			return toks, nil
		}
	}
}

func newLexer(src string, pos Pos) *lexer {
	l := &lexer{line: pos.Line, col: pos.Col}
	if bad, ok := invalidUTF8(src, pos); ok {
		l.err = NewLexError(ErrLex, bad, CodeInvalidUTF8, "invalid UTF-8 in source", "")
		return l
	}
	l.src = []rune(src)
	return l
}

func invalidUTF8(src string, pos Pos) (Pos, bool) {
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

func (l *lexer) Next() (Token, error) {
	if l.err != nil {
		return Token{}, l.err
	}

	t, err := l.scan()
	if err != nil {
		l.err = err
	}

	return t, err
}

func (l *lexer) scan() (Token, error) {
	l.skipSpace()
	pos := l.pos()
	if l.off >= len(l.src) {
		return l.tok(EOF, "", pos)
	}
	r := l.src[l.off]
	switch {
	case r == '\n':
		l.advance()
		return l.tok(Newline, "\n", pos)
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

func (l *lexer) ident(pos Pos) (Token, error) {
	start := l.off
	for l.off < len(l.src) && (isLetter(l.src[l.off]) || isDigit(l.src[l.off])) {
		l.advance()
	}

	lit := string(l.src[start:l.off])

	return l.tok(Lookup(lit), lit, pos)
}

func (l *lexer) number(pos Pos) (Token, error) {
	start := l.off
	if l.src[l.off] == '0' && isDigit(l.peek(1)) {
		return l.fail(pos, CodeLeadingZero, "write 7 instead of 007", "numbers cannot have leading zeros")
	}

	l.digits()
	if l.peek(0) == '.' && isDigit(l.peek(1)) {
		l.advance()
		l.digits()
	}

	if isLetter(l.peek(0)) {
		return l.fail(pos, CodeInvalidNumber, "", "invalid number literal")
	}

	return l.tok(Number, string(l.src[start:l.off]), pos)
}

func (l *lexer) digits() {
	for l.off < len(l.src) && isDigit(l.src[l.off]) {
		l.advance()
	}
}

func (l *lexer) str(pos Pos) (Token, error) {
	start := l.off
	l.advance()
	var text strings.Builder
	var parts []Part
	for l.off < len(l.src) {
		at := l.pos()
		switch r := l.src[l.off]; r {
		case '"':
			l.advance()
			if parts == nil {
				return l.tok(String, text.String(), pos)
			}
			if text.Len() > 0 {
				parts = append(parts, Part{Text: text.String()})
			}
			return Token{Kind: Template, Lit: string(l.src[start:l.off]), Pos: pos, Parts: parts}, nil
		case '\n':
			return l.fail(at, CodeNewlineInString, `use \n for a line break`, "string cannot contain a raw newline")
		case '\\':
			l.advance()
			if l.off >= len(l.src) {
				return l.fail(pos, CodeUnterminatedString, "", "unterminated string")
			}
			if l.src[l.off] == '\n' {
				return l.fail(l.pos(), CodeNewlineInString, `use \n for a line break`, "string cannot contain a raw newline")
			}
			e, ok := escapes[l.src[l.off]]
			if !ok {
				return l.fail(at, CodeInvalidEscape, escapeHint, fmt.Sprintf("unknown escape \\%c", l.src[l.off]))
			}
			l.advance()
			text.WriteRune(e)
		case '$':
			if l.peek(1) != '{' {
				text.WriteRune(l.advance())
				continue
			}
			if text.Len() > 0 {
				parts = append(parts, Part{Text: text.String()})
				text.Reset()
			}
			part, err := l.interp(at)
			if err != nil {
				return Token{}, err
			}
			parts = append(parts, part)
		default:
			text.WriteRune(l.advance())
		}
	}
	return l.fail(pos, CodeUnterminatedString, "", "unterminated string")
}

func (l *lexer) interp(at Pos) (Part, error) {
	l.advance()
	l.advance()
	pos := l.pos()
	start := l.off
	depth := 0
	for l.off < len(l.src) {
		switch l.src[l.off] {
		case '\n':
			return Part{}, l.errAt(l.pos(), CodeNewlineInString, `use \n for a line break`, "string cannot contain a raw newline")
		case '"':
			if err := l.skipString(at); err != nil {
				return Part{}, err
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
				return Part{}, l.errAt(at, CodeEmptyInterp, "", "empty ${ } in string")
			}
			return Part{Expr: expr, IsExpr: true, Pos: pos}, nil
		}
		l.advance()
	}
	return Part{}, l.errAt(at, CodeUnterminatedInterp, "", "missing } to close ${")
}

func (l *lexer) skipString(at Pos) error {
	l.advance()
	for l.off < len(l.src) {
		switch l.src[l.off] {
		case '\n':
			return l.errAt(l.pos(), CodeNewlineInString, `use \n for a line break`, "string cannot contain a raw newline")
		case '"':
			l.advance()
			return nil
		case '\\':
			l.advance()
			if l.off >= len(l.src) {
				return l.errAt(at, CodeUnterminatedInterp, "", "missing } to close ${")
			}
		}
		l.advance()
	}
	return l.errAt(at, CodeUnterminatedInterp, "", "missing } to close ${")
}

func (l *lexer) operator(pos Pos) (Token, error) {
	r := l.advance()
	next := l.peek(0)
	switch r {
	case '+', '-', '*', '/', '%':
		if next == '=' {
			return l.fail(pos, CodeCompoundAssign, fmt.Sprintf("use x = x %c y", r), fmt.Sprintf("compound assignment %c= is not supported", r))
		}
	case '=':
		if next == '=' {
			l.advance()
			return l.tok(Eq, "==", pos)
		}
		return l.tok(Assign, "=", pos)
	case '!':
		if next == '=' {
			l.advance()
			return l.tok(NotEq, "!=", pos)
		}
		return l.tok(Not, "!", pos)
	case '<':
		if next == '=' {
			l.advance()
			return l.tok(LtEq, "<=", pos)
		}
		return l.tok(Lt, "<", pos)
	case '>':
		if next == '=' {
			l.advance()
			return l.tok(GtEq, ">=", pos)
		}
		return l.tok(Gt, ">", pos)
	case '&':
		if next == '&' {
			l.advance()
			return l.tok(And, "&&", pos)
		}
		return l.fail(pos, CodeUnexpectedChar, "", `unexpected character "&"`)
	case '|':
		if next == '|' {
			l.advance()
			return l.tok(Or, "||", pos)
		}
		return l.fail(pos, CodeUnexpectedChar, "use or", `unexpected character "|"`)
	case '.':
		if next == '.' {
			l.advance()
			return l.tok(DotDot, "..", pos)
		}
		return l.tok(Dot, ".", pos)
	case ';':
		return l.fail(pos, CodeSemicolon, "one statement per line", `";" is not allowed`)
	case '\'':
		return l.fail(pos, CodeSingleQuote, `use double quotes "..."`, "single quotes are not allowed")
	}
	if k, ok := single[r]; ok {
		if err := l.nest(k, pos); err != nil {
			return Token{}, err
		}
		return l.tok(k, string(r), pos)
	}
	return l.fail(pos, CodeUnexpectedChar, "", fmt.Sprintf("unexpected character %q", r))
}

func (l *lexer) nest(k Kind, pos Pos) error {
	switch k {
	case LParen, LBracket, LBrace:
		l.depth++
		if l.depth > MaxDepth {
			return l.errAt(pos, CodeNesting, "", "excessive nesting")
		}
	case RParen, RBracket, RBrace:
		if l.depth > 0 {
			l.depth--
		}
	}
	return nil
}

func (l *lexer) tok(kind Kind, lit string, pos Pos) (Token, error) {
	return Token{Kind: kind, Lit: lit, Pos: pos}, nil
}

func (l *lexer) fail(pos Pos, code, hint, msg string) (Token, error) {
	return Token{}, l.errAt(pos, code, hint, msg)
}

func (l *lexer) errAt(pos Pos, code, hint, msg string) error {
	return NewLexError(ErrLex, pos, code, msg, hint)
}

func (l *lexer) pos() Pos {
	return Pos{Line: l.line, Col: l.col}
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

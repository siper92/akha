package lexer

import (
	"iter"
	"strings"

	"github.com/siper92/akha/lang/token"
)

type lexer struct {
	src  []rune
	pos  int
	line int
	col  int
	opts options
}

var _ Lexer = (*lexer)(nil)

func New(src string, opts ...Option) Lexer {
	lx := &lexer{src: []rune(src), line: 1, col: 1}
	for _, o := range opts {
		o(&lx.opts)
	}
	return lx
}

func (l *lexer) Next() token.Token {
	for {
		l.skipSpace()
		if l.pos >= len(l.src) {
			return token.Token{Kind: token.EOF, Pos: l.here()}
		}
		r := l.src[l.pos]
		start := l.here()
		switch {
		case r == '\n':
			return l.emit(token.NEWLINE, 1, start)
		case r == '/' && l.peek(1) == '/':
			text := l.comment()
			if l.opts.comments {
				return token.Token{Kind: token.COMMENT, Lit: text, Pos: start}
			}
			continue
		case isIdentStart(r):
			lit := l.ident()
			return token.Token{Kind: token.Lookup(lit), Lit: lit, Pos: start}
		case isDigit(r):
			return token.Token{Kind: token.INT, Lit: l.digits(), Pos: start}
		case r == '"':
			return l.str(start)
		case r == '.':
			if l.peek(1) == '.' && l.peek(2) == '.' {
				return l.emit(token.ELLIPSIS, 3, start)
			}
			return l.emit(token.DOT, 1, start)
		case r == '(':
			return l.emit(token.LPAREN, 1, start)
		case r == ')':
			return l.emit(token.RPAREN, 1, start)
		case r == '{':
			return l.emit(token.LBRACE, 1, start)
		case r == '}':
			return l.emit(token.RBRACE, 1, start)
		case r == ',':
			return l.emit(token.COMMA, 1, start)
		case r == '+':
			return l.emit(token.PLUS, 1, start)
		case r == '-':
			return l.emit(token.MINUS, 1, start)
		case r == '*':
			return l.emit(token.STAR, 1, start)
		case r == '/':
			return l.emit(token.SLASH, 1, start)
		case r == '%':
			return l.emit(token.PERCENT, 1, start)
		case r == '=':
			if l.peek(1) == '=' {
				return l.emit(token.EQ, 2, start)
			}
			return l.emit(token.ASSIGN, 1, start)
		case r == '!' && l.peek(1) == '=':
			return l.emit(token.NEQ, 2, start)
		case r == '<':
			if l.peek(1) == '=' {
				return l.emit(token.LTE, 2, start)
			}
			return l.emit(token.LT, 1, start)
		case r == '>':
			if l.peek(1) == '=' {
				return l.emit(token.GTE, 2, start)
			}
			return l.emit(token.GT, 1, start)
		}
		return l.emit(token.ILLEGAL, 1, start)
	}
}

func (l *lexer) All() iter.Seq[token.Token] {
	return func(yield func(token.Token) bool) {
		for {
			t := l.Next()
			if !yield(t) || t.Kind == token.EOF {
				return
			}
		}
	}
}

func (l *lexer) here() token.Pos {
	return token.Pos{Line: l.line, Col: l.col}
}

func (l *lexer) peek(n int) rune {
	if l.pos+n >= len(l.src) {
		return 0
	}
	return l.src[l.pos+n]
}

func (l *lexer) advance() {
	if l.src[l.pos] == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	l.pos++
}

func (l *lexer) emit(kind token.Kind, n int, start token.Pos) token.Token {
	lit := string(l.src[l.pos : l.pos+n])
	for range n {
		l.advance()
	}
	return token.Token{Kind: kind, Lit: lit, Pos: start}
}

func (l *lexer) skipSpace() {
	for l.pos < len(l.src) {
		switch l.src[l.pos] {
		case ' ', '\t', '\r':
			l.advance()
		default:
			return
		}
	}
}

func (l *lexer) comment() string {
	l.advance()
	l.advance()
	begin := l.pos
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		l.advance()
	}
	return strings.TrimSuffix(string(l.src[begin:l.pos]), "\r")
}

func (l *lexer) ident() string {
	begin := l.pos
	for l.pos < len(l.src) && isIdentPart(l.src[l.pos]) {
		l.advance()
	}
	return string(l.src[begin:l.pos])
}

func (l *lexer) digits() string {
	begin := l.pos
	for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
		l.advance()
	}
	return string(l.src[begin:l.pos])
}

func (l *lexer) str(start token.Pos) token.Token {
	begin := l.pos
	l.advance()
	var sb strings.Builder
	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		r := l.src[l.pos]
		l.advance()
		switch r {
		case '"':
			return token.Token{Kind: token.STRING, Lit: sb.String(), Pos: start}
		case '\\':
			if l.pos >= len(l.src) || l.src[l.pos] == '\n' {
				continue
			}
			e := l.src[l.pos]
			l.advance()
			switch e {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(e)
			}
		default:
			sb.WriteRune(r)
		}
	}
	return token.Token{Kind: token.ILLEGAL, Lit: string(l.src[begin:l.pos]), Pos: start}
}

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || isDigit(r)
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

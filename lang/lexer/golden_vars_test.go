package lexer_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/token"
)

func TestGoldenVars(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "vars.ak"))
	if err != nil {
		t.Fatalf("read vars.ak: %v", err)
	}
	t.Run("default", func(t *testing.T) {
		got := slices.Collect(lexer.New(string(src)).All())
		assertTokens(t, varsTokens(false), got)
	})
	t.Run("with_comments", func(t *testing.T) {
		got := slices.Collect(lexer.New(string(src), lexer.WithComments()).All())
		assertTokens(t, varsTokens(true), got)
	})
}

func varsTokens(comments bool) []token.Token {
	var ts []token.Token
	add := func(k token.Kind, lit string, line, col int) {
		ts = append(ts, tok(k, lit, line, col))
	}
	cm := func(lit string, line, col int) {
		if comments {
			add(token.COMMENT, lit, line, col)
		}
	}

	cm(" variables", 1, 1)
	add(token.NEWLINE, "\n", 1, 13)

	add(token.IDENT, "Ak", 2, 1)
	add(token.DOT, ".", 2, 3)
	add(token.IDENT, "Allow", 2, 4)
	add(token.LPAREN, "(", 2, 9)
	add(token.IDENT, "FS", 2, 10)
	add(token.ELLIPSIS, "...", 2, 12)
	add(token.RPAREN, ")", 2, 15)
	add(token.NEWLINE, "\n", 2, 16)

	add(token.LET, "let", 3, 1)
	add(token.IDENT, "name", 3, 5)
	add(token.ASSIGN, "=", 3, 10)
	add(token.STRING, "world", 3, 12)
	add(token.NEWLINE, "\n", 3, 19)

	add(token.IDENT, "name", 4, 1)
	add(token.ASSIGN, "=", 4, 6)
	add(token.STRING, "akha", 4, 8)
	add(token.NEWLINE, "\n", 4, 14)

	add(token.LET, "let", 5, 1)
	add(token.IDENT, "content", 5, 5)
	add(token.ASSIGN, "=", 5, 13)
	add(token.IDENT, "FS", 5, 15)
	add(token.DOT, ".", 5, 17)
	add(token.IDENT, "ReadFile", 5, 18)
	add(token.LPAREN, "(", 5, 26)
	add(token.STRING, "f.txt", 5, 27)
	add(token.RPAREN, ")", 5, 34)
	add(token.NEWLINE, "\n", 5, 35)

	add(token.LET, "let", 6, 1)
	add(token.IDENT, "n", 6, 5)
	add(token.ASSIGN, "=", 6, 7)
	add(token.INT, "1", 6, 9)
	add(token.PLUS, "+", 6, 11)
	add(token.INT, "2", 6, 13)
	add(token.STAR, "*", 6, 15)
	add(token.INT, "3", 6, 17)
	add(token.NEWLINE, "\n", 6, 18)

	add(token.LET, "let", 7, 1)
	add(token.IDENT, "s", 7, 5)
	add(token.ASSIGN, "=", 7, 7)
	add(token.STRING, "a", 7, 9)
	add(token.PLUS, "+", 7, 13)
	add(token.STRING, "b", 7, 15)
	add(token.NEWLINE, "\n", 7, 18)

	add(token.LET, "let", 8, 1)
	add(token.IDENT, "neg", 8, 5)
	add(token.ASSIGN, "=", 8, 9)
	add(token.MINUS, "-", 8, 11)
	add(token.LPAREN, "(", 8, 12)
	add(token.IDENT, "n", 8, 13)
	add(token.MINUS, "-", 8, 15)
	add(token.INT, "10", 8, 17)
	add(token.RPAREN, ")", 8, 19)
	add(token.SLASH, "/", 8, 21)
	add(token.INT, "2", 8, 23)
	add(token.PERCENT, "%", 8, 25)
	add(token.INT, "3", 8, 27)
	add(token.NEWLINE, "\n", 8, 28)

	add(token.LET, "let", 9, 1)
	add(token.IDENT, "ok", 9, 5)
	add(token.ASSIGN, "=", 9, 8)
	add(token.IDENT, "n", 9, 10)
	add(token.GT, ">", 9, 12)
	add(token.INT, "3", 9, 14)
	add(token.AND, "and", 9, 16)
	add(token.IDENT, "name", 9, 20)
	add(token.NEQ, "!=", 9, 25)
	add(token.STRING, "", 9, 28)
	add(token.OR, "or", 9, 31)
	add(token.NOT, "not", 9, 34)
	add(token.LPAREN, "(", 9, 38)
	add(token.IDENT, "s", 9, 39)
	add(token.EQ, "==", 9, 41)
	add(token.STRING, "ab", 9, 44)
	add(token.RPAREN, ")", 9, 48)
	add(token.NEWLINE, "\n", 9, 49)

	add(token.LET, "let", 10, 1)
	add(token.IDENT, "le", 10, 5)
	add(token.ASSIGN, "=", 10, 8)
	add(token.IDENT, "n", 10, 10)
	add(token.LTE, "<=", 10, 12)
	add(token.INT, "7", 10, 15)
	add(token.NEWLINE, "\n", 10, 16)

	add(token.LET, "let", 11, 1)
	add(token.IDENT, "ge", 11, 5)
	add(token.ASSIGN, "=", 11, 8)
	add(token.IDENT, "n", 11, 10)
	add(token.GTE, ">=", 11, 12)
	add(token.INT, "7", 11, 15)
	add(token.NEWLINE, "\n", 11, 16)

	add(token.LET, "let", 12, 1)
	add(token.IDENT, "lt", 12, 5)
	add(token.ASSIGN, "=", 12, 8)
	add(token.IDENT, "n", 12, 10)
	add(token.LT, "<", 12, 12)
	add(token.INT, "7", 12, 14)
	add(token.NEWLINE, "\n", 12, 15)

	add(token.IDENT, "Ak", 13, 1)
	add(token.DOT, ".", 13, 3)
	add(token.IDENT, "Log", 13, 4)
	add(token.LPAREN, "(", 13, 7)
	add(token.IDENT, "s", 13, 8)
	add(token.RPAREN, ")", 13, 9)
	add(token.NEWLINE, "\n", 13, 10)

	add(token.IDENT, "Ak", 14, 1)
	add(token.DOT, ".", 14, 3)
	add(token.IDENT, "Exit", 14, 4)
	add(token.LPAREN, "(", 14, 8)
	add(token.STRING, "done", 14, 9)
	add(token.COMMA, ",", 14, 15)
	add(token.IDENT, "code", 14, 17)
	add(token.ASSIGN, "=", 14, 21)
	add(token.IDENT, "n", 14, 22)
	add(token.RPAREN, ")", 14, 23)
	add(token.NEWLINE, "\n", 14, 24)

	add(token.EOF, "", 15, 1)
	return ts
}

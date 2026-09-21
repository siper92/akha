package lexer_test

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	"github.com/siper92/akha/internal/tu"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/token"
)

type lexCase = tu.Case[string, []token.Token]

func tok(k token.Kind, lit string, line, col int) token.Token {
	return token.Token{Kind: k, Lit: lit, Pos: token.Pos{Line: line, Col: col}}
}

func eof(line, col int) token.Token {
	return tok(token.EOF, "", line, col)
}

func nl(line, col int) token.Token {
	return tok(token.NEWLINE, "\n", line, col)
}

func lexAll(opts ...lexer.Option) func(string) ([]token.Token, error) {
	return func(src string) ([]token.Token, error) {
		return slices.Collect(lexer.New(src, opts...).All()), nil
	}
}

func TestLexIdents(t *testing.T) {
	cases := []lexCase{
		// --- idents
		{
			Name:     "single_letter",
			Input:    "x",
			Expected: []token.Token{tok(token.IDENT, "x", 1, 1), eof(1, 2)},
		},
		{
			Name:     "underscore_digits_mixed_case",
			Input:    "_a1B",
			Expected: []token.Token{tok(token.IDENT, "_a1B", 1, 1), eof(1, 5)},
		},
		{
			Name:     "two_idents_space_separated",
			Input:    "Ak Allow",
			Expected: []token.Token{tok(token.IDENT, "Ak", 1, 1), tok(token.IDENT, "Allow", 1, 4), eof(1, 9)},
		},
		{
			Name:     "tab_counts_one_col",
			Input:    "\tx",
			Expected: []token.Token{tok(token.IDENT, "x", 1, 2), eof(1, 3)},
		},
		{
			Name:     "digit_cannot_start_ident",
			Input:    "a1 1a",
			Expected: []token.Token{
				tok(token.IDENT, "a1", 1, 1),
				tok(token.INT, "1", 1, 4),
				tok(token.IDENT, "a", 1, 5),
				eof(1, 6),
			},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(), nil)
}

func TestLexInts(t *testing.T) {
	cases := []lexCase{
		// --- ints
		{
			Name:     "zero",
			Input:    "0",
			Expected: []token.Token{tok(token.INT, "0", 1, 1), eof(1, 2)},
		},
		{
			Name:     "many_digits",
			Input:    "12345",
			Expected: []token.Token{tok(token.INT, "12345", 1, 1), eof(1, 6)},
		},
		{
			Name:     "leading_zeros_kept",
			Input:    "007",
			Expected: []token.Token{tok(token.INT, "007", 1, 1), eof(1, 4)},
		},
		{
			Name:     "minus_is_illegal_then_int",
			Input:    "-1",
			Expected: []token.Token{tok(token.ILLEGAL, "-", 1, 1), tok(token.INT, "1", 1, 2), eof(1, 3)},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(), nil)
}

func TestLexStrings(t *testing.T) {
	cases := []lexCase{
		// --- strings
		{
			Name:     "empty",
			Input:    `""`,
			Expected: []token.Token{tok(token.STRING, "", 1, 1), eof(1, 3)},
		},
		{
			Name:     "plain",
			Input:    `"hello"`,
			Expected: []token.Token{tok(token.STRING, "hello", 1, 1), eof(1, 8)},
		},
		{
			Name:     "spaces_inside_kept",
			Input:    `"a b"`,
			Expected: []token.Token{tok(token.STRING, "a b", 1, 1), eof(1, 6)},
		},
		{
			Name:     "unicode_inside_counts_one_col",
			Input:    `"é"`,
			Expected: []token.Token{tok(token.STRING, "é", 1, 1), eof(1, 4)},
		},
		{
			Name:     "two_strings",
			Input:    `"a" "b"`,
			Expected: []token.Token{tok(token.STRING, "a", 1, 1), tok(token.STRING, "b", 1, 5), eof(1, 8)},
		},
		{
			Name:     "cr_inside_kept_literally",
			Input:    "\"a\rb\"",
			Expected: []token.Token{tok(token.STRING, "a\rb", 1, 1), eof(1, 6)},
		},
		// --- escapes
		{
			Name:     "escape_quote",
			Input:    `"a\"b"`,
			Expected: []token.Token{tok(token.STRING, `a"b`, 1, 1), eof(1, 7)},
		},
		{
			Name:     "escape_backslash",
			Input:    `"a\\b"`,
			Expected: []token.Token{tok(token.STRING, `a\b`, 1, 1), eof(1, 7)},
		},
		{
			Name:     "escape_newline",
			Input:    `"a\nb"`,
			Expected: []token.Token{tok(token.STRING, "a\nb", 1, 1), eof(1, 7)},
		},
		{
			Name:     "escape_tab",
			Input:    `"a\tb"`,
			Expected: []token.Token{tok(token.STRING, "a\tb", 1, 1), eof(1, 7)},
		},
		{
			Name:     "unknown_escape_kept_literally",
			Input:    `"\x"`,
			Expected: []token.Token{tok(token.STRING, `\x`, 1, 1), eof(1, 5)},
		},
		{
			Name:     "all_escapes_together",
			Input:    `"a\"b\\c\n\t"`,
			Expected: []token.Token{tok(token.STRING, "a\"b\\c\n\t", 1, 1), eof(1, 15)},
		},
		// --- unterminated strings
		{
			Name:     "unterminated_at_eof",
			Input:    `"abc`,
			Expected: []token.Token{tok(token.ILLEGAL, `"abc`, 1, 1), eof(1, 5)},
		},
		{
			Name:     "unterminated_at_newline_then_continues",
			Input:    "\"abc\nX",
			Expected: []token.Token{
				tok(token.ILLEGAL, `"abc`, 1, 1),
				nl(1, 5),
				tok(token.IDENT, "X", 2, 1),
				eof(2, 2),
			},
		},
		{
			Name:     "unterminated_trailing_backslash_before_newline",
			Input:    "\"ab\\\n",
			Expected: []token.Token{tok(token.ILLEGAL, `"ab\`, 1, 1), nl(1, 5), eof(2, 1)},
		},
		{
			Name:     "unterminated_trailing_backslash_at_eof",
			Input:    `"ab\`,
			Expected: []token.Token{tok(token.ILLEGAL, `"ab\`, 1, 1), eof(1, 5)},
		},
		{
			Name:     "lone_quote",
			Input:    `"`,
			Expected: []token.Token{tok(token.ILLEGAL, `"`, 1, 1), eof(1, 2)},
		},
		{
			Name:     "escaped_quote_does_not_close",
			Input:    `"a\"`,
			Expected: []token.Token{tok(token.ILLEGAL, `"a\"`, 1, 1), eof(1, 5)},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(), nil)
}

func TestLexPunctuation(t *testing.T) {
	cases := []lexCase{
		// --- punctuation
		{
			Name:     "dot",
			Input:    ".",
			Expected: []token.Token{tok(token.DOT, ".", 1, 1), eof(1, 2)},
		},
		{
			Name:     "ellipsis",
			Input:    "...",
			Expected: []token.Token{tok(token.ELLIPSIS, "...", 1, 1), eof(1, 4)},
		},
		{
			Name:     "two_dots_are_two_dot_tokens",
			Input:    "..",
			Expected: []token.Token{tok(token.DOT, ".", 1, 1), tok(token.DOT, ".", 1, 2), eof(1, 3)},
		},
		{
			Name:     "four_dots_are_ellipsis_then_dot",
			Input:    "....",
			Expected: []token.Token{tok(token.ELLIPSIS, "...", 1, 1), tok(token.DOT, ".", 1, 4), eof(1, 5)},
		},
		{
			Name:     "parens",
			Input:    "()",
			Expected: []token.Token{tok(token.LPAREN, "(", 1, 1), tok(token.RPAREN, ")", 1, 2), eof(1, 3)},
		},
		{
			Name:     "comma",
			Input:    ",",
			Expected: []token.Token{tok(token.COMMA, ",", 1, 1), eof(1, 2)},
		},
		{
			Name:     "assign",
			Input:    "=",
			Expected: []token.Token{tok(token.ASSIGN, "=", 1, 1), eof(1, 2)},
		},
		{
			Name:     "kwarg_shape",
			Input:    "x=1",
			Expected: []token.Token{
				tok(token.IDENT, "x", 1, 1),
				tok(token.ASSIGN, "=", 1, 2),
				tok(token.INT, "1", 1, 3),
				eof(1, 4),
			},
		},
		{
			Name:     "comma_with_spaces",
			Input:    "a , b",
			Expected: []token.Token{
				tok(token.IDENT, "a", 1, 1),
				tok(token.COMMA, ",", 1, 3),
				tok(token.IDENT, "b", 1, 5),
				eof(1, 6),
			},
		},
		{
			Name:     "allow_with_spread",
			Input:    "Ak.Allow(FS...)",
			Expected: []token.Token{
				tok(token.IDENT, "Ak", 1, 1),
				tok(token.DOT, ".", 1, 3),
				tok(token.IDENT, "Allow", 1, 4),
				tok(token.LPAREN, "(", 1, 9),
				tok(token.IDENT, "FS", 1, 10),
				tok(token.ELLIPSIS, "...", 1, 12),
				tok(token.RPAREN, ")", 1, 15),
				eof(1, 16),
			},
		},
		{
			Name:     "allow_with_two_dots",
			Input:    "Ak.Allow(FS..)",
			Expected: []token.Token{
				tok(token.IDENT, "Ak", 1, 1),
				tok(token.DOT, ".", 1, 3),
				tok(token.IDENT, "Allow", 1, 4),
				tok(token.LPAREN, "(", 1, 9),
				tok(token.IDENT, "FS", 1, 10),
				tok(token.DOT, ".", 1, 12),
				tok(token.DOT, ".", 1, 13),
				tok(token.RPAREN, ")", 1, 14),
				eof(1, 15),
			},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(), nil)
}

func TestLexCommentsDropped(t *testing.T) {
	cases := []lexCase{
		// --- comments dropped by default
		{
			Name:     "only_comment_no_newline",
			Input:    "// only comment",
			Expected: []token.Token{eof(1, 16)},
		},
		{
			Name:     "empty_comment",
			Input:    "//",
			Expected: []token.Token{eof(1, 3)},
		},
		{
			Name:     "comment_after_code_newline_kept",
			Input:    "x // c\ny",
			Expected: []token.Token{
				tok(token.IDENT, "x", 1, 1),
				nl(1, 7),
				tok(token.IDENT, "y", 2, 1),
				eof(2, 2),
			},
		},
		{
			Name:     "comment_line_between_code",
			Input:    "a\n// c\nb",
			Expected: []token.Token{
				tok(token.IDENT, "a", 1, 1),
				nl(1, 2),
				nl(2, 5),
				tok(token.IDENT, "b", 3, 1),
				eof(3, 2),
			},
		},
		// --- lone slash
		{
			Name:     "lone_slash_is_illegal",
			Input:    "/",
			Expected: []token.Token{tok(token.ILLEGAL, "/", 1, 1), eof(1, 2)},
		},
		{
			Name:     "separated_slashes_are_two_illegal",
			Input:    "/ /",
			Expected: []token.Token{tok(token.ILLEGAL, "/", 1, 1), tok(token.ILLEGAL, "/", 1, 3), eof(1, 4)},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(), nil)
}

func TestLexCommentsKept(t *testing.T) {
	cases := []lexCase{
		// --- comments kept with WithComments
		{
			Name:     "only_comment_no_newline",
			Input:    "// only comment",
			Expected: []token.Token{tok(token.COMMENT, " only comment", 1, 1), eof(1, 16)},
		},
		{
			Name:     "empty_comment",
			Input:    "//",
			Expected: []token.Token{tok(token.COMMENT, "", 1, 1), eof(1, 3)},
		},
		{
			Name:     "comment_after_code",
			Input:    "x // c\ny",
			Expected: []token.Token{
				tok(token.IDENT, "x", 1, 1),
				tok(token.COMMENT, " c", 1, 3),
				nl(1, 7),
				tok(token.IDENT, "y", 2, 1),
				eof(2, 2),
			},
		},
		{
			Name:     "crlf_trailing_cr_trimmed",
			Input:    "// c\r\nx",
			Expected: []token.Token{
				tok(token.COMMENT, " c", 1, 1),
				nl(1, 6),
				tok(token.IDENT, "x", 2, 1),
				eof(2, 2),
			},
		},
		{
			Name:     "nested_slashes_part_of_text",
			Input:    "//a//b",
			Expected: []token.Token{tok(token.COMMENT, "a//b", 1, 1), eof(1, 7)},
		},
		{
			Name:     "quotes_inside_comment_not_lexed",
			Input:    `// "str"`,
			Expected: []token.Token{tok(token.COMMENT, ` "str"`, 1, 1), eof(1, 9)},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(lexer.WithComments()), nil)
}

func TestLexNewlines(t *testing.T) {
	cases := []lexCase{
		// --- newlines
		{
			Name:     "empty_source",
			Input:    "",
			Expected: []token.Token{eof(1, 1)},
		},
		{
			Name:     "single_newline",
			Input:    "\n",
			Expected: []token.Token{nl(1, 1), eof(2, 1)},
		},
		{
			Name:     "two_newlines",
			Input:    "\n\n",
			Expected: []token.Token{nl(1, 1), nl(2, 1), eof(3, 1)},
		},
		{
			Name:     "lf_between_idents",
			Input:    "a\nb",
			Expected: []token.Token{tok(token.IDENT, "a", 1, 1), nl(1, 2), tok(token.IDENT, "b", 2, 1), eof(2, 2)},
		},
		{
			Name:     "trailing_spaces_before_newline",
			Input:    "a \n  b",
			Expected: []token.Token{tok(token.IDENT, "a", 1, 1), nl(1, 3), tok(token.IDENT, "b", 2, 3), eof(2, 4)},
		},
		// --- carriage returns
		{
			Name:     "crlf_between_idents",
			Input:    "a\r\nb",
			Expected: []token.Token{tok(token.IDENT, "a", 1, 1), nl(1, 3), tok(token.IDENT, "b", 2, 1), eof(2, 2)},
		},
		{
			Name:     "lone_cr_skipped",
			Input:    "\r",
			Expected: []token.Token{eof(1, 2)},
		},
		{
			Name:     "two_crlf",
			Input:    "\r\n\r\n",
			Expected: []token.Token{nl(1, 2), nl(2, 2), eof(3, 1)},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(), nil)
}

func TestLexIllegal(t *testing.T) {
	cases := []lexCase{
		// --- illegal runes
		{
			Name:     "minus",
			Input:    "-",
			Expected: []token.Token{tok(token.ILLEGAL, "-", 1, 1), eof(1, 2)},
		},
		{
			Name:     "plus",
			Input:    "+",
			Expected: []token.Token{tok(token.ILLEGAL, "+", 1, 1), eof(1, 2)},
		},
		{
			Name:     "semicolon",
			Input:    ";",
			Expected: []token.Token{tok(token.ILLEGAL, ";", 1, 1), eof(1, 2)},
		},
		{
			Name:     "single_quotes",
			Input:    "'a'",
			Expected: []token.Token{
				tok(token.ILLEGAL, "'", 1, 1),
				tok(token.IDENT, "a", 1, 2),
				tok(token.ILLEGAL, "'", 1, 3),
				eof(1, 4),
			},
		},
		{
			Name:     "braces",
			Input:    "{}",
			Expected: []token.Token{tok(token.ILLEGAL, "{", 1, 1), tok(token.ILLEGAL, "}", 1, 2), eof(1, 3)},
		},
		{
			Name:     "illegal_between_idents",
			Input:    "a-b",
			Expected: []token.Token{
				tok(token.IDENT, "a", 1, 1),
				tok(token.ILLEGAL, "-", 1, 2),
				tok(token.IDENT, "b", 1, 3),
				eof(1, 4),
			},
		},
		// --- unicode and invalid utf8
		{
			Name:     "unicode_letter_col_counts_runes",
			Input:    "a é b",
			Expected: []token.Token{
				tok(token.IDENT, "a", 1, 1),
				tok(token.ILLEGAL, "é", 1, 3),
				tok(token.IDENT, "b", 1, 5),
				eof(1, 6),
			},
		},
		{
			Name:     "invalid_utf8_is_replacement_rune",
			Input:    "\xff",
			Expected: []token.Token{tok(token.ILLEGAL, "�", 1, 1), eof(1, 2)},
		},
	}
	tu.Run(tu.New(t), cases, lexAll(), nil)
}

func TestNextAfterEOF(t *testing.T) {
	cases := []lexCase{
		// --- eof repeats
		{
			Name:     "empty_source",
			Input:    "",
			Expected: []token.Token{eof(1, 1), eof(1, 1), eof(1, 1)},
		},
		{
			Name:     "after_ident",
			Input:    "x",
			Expected: []token.Token{eof(1, 2), eof(1, 2), eof(1, 2)},
		},
		{
			Name:     "after_newline",
			Input:    "x\n",
			Expected: []token.Token{eof(2, 1), eof(2, 1), eof(2, 1)},
		},
	}
	tu.Run(tu.New(t), cases, func(src string) ([]token.Token, error) {
		lx := lexer.New(src)
		for lx.Next().Kind != token.EOF {
		}
		return []token.Token{lx.Next(), lx.Next(), lx.Next()}, nil
	}, nil)
}

func TestAllTerminates(t *testing.T) {
	cases := []tu.Case[string, int]{
		// --- all yields every token then stops
		{
			Name:     "empty_source_yields_eof_only",
			Input:    "",
			Expected: 1,
		},
		{
			Name:     "ident_then_eof",
			Input:    "x",
			Expected: 2,
		},
		{
			Name:     "ident_newline_eof",
			Input:    "x\n",
			Expected: 3,
		},
		{
			Name:     "unterminated_string_still_ends",
			Input:    `"abc`,
			Expected: 2,
		},
	}
	tu.Run(tu.New(t), cases, func(src string) (int, error) {
		n := 0
		for range lexer.New(src).All() {
			n++
		}
		return n, nil
	}, nil)
}

func TestAllBreakEarly(t *testing.T) {
	cases := []lexCase{
		// --- breaking out of all stops it and next continues
		{
			Name:     "first_from_all_second_from_next",
			Input:    "a b",
			Expected: []token.Token{tok(token.IDENT, "a", 1, 1), tok(token.IDENT, "b", 1, 3)},
		},
		{
			Name:     "eof_only",
			Input:    "",
			Expected: []token.Token{eof(1, 1), eof(1, 1)},
		},
	}
	tu.Run(tu.New(t), cases, func(src string) ([]token.Token, error) {
		lx := lexer.New(src)
		var first token.Token
		for tk := range lx.All() {
			first = tk
			break
		}
		return []token.Token{first, lx.Next()}, nil
	}, nil)
}

func TestGoldenSpec(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "spec.ak"))
	if err != nil {
		t.Fatalf("read spec.ak: %v", err)
	}
	t.Run("default", func(t *testing.T) {
		got := slices.Collect(lexer.New(string(src)).All())
		assertTokens(t, specTokens(false), got)
	})
	t.Run("with_comments", func(t *testing.T) {
		got := slices.Collect(lexer.New(string(src), lexer.WithComments()).All())
		assertTokens(t, specTokens(true), got)
	})
}

func assertTokens(t *testing.T, want, got []token.Token) {
	t.Helper()
	if reflect.DeepEqual(want, got) {
		return
	}
	n := min(len(want), len(got))
	for i := 0; i < n; i++ {
		if want[i] != got[i] {
			t.Errorf("token %d: want %#v, got %#v", i, want[i], got[i])
			return
		}
	}
	t.Errorf("token count: want %d, got %d", len(want), len(got))
}

func specTokens(comments bool) []token.Token {
	var ts []token.Token
	add := func(k token.Kind, lit string, line, col int) {
		ts = append(ts, tok(k, lit, line, col))
	}
	cm := func(lit string, line, col int) {
		if comments {
			add(token.COMMENT, lit, line, col)
		}
	}

	cm(" utils - Ak always available", 1, 1)
	add(token.NEWLINE, "\n", 1, 31)

	add(token.IDENT, "Ak", 2, 1)
	add(token.DOT, ".", 2, 3)
	add(token.IDENT, "Allow", 2, 4)
	add(token.LPAREN, "(", 2, 9)
	add(token.IDENT, "FS", 2, 10)
	add(token.ELLIPSIS, "...", 2, 12)
	add(token.RPAREN, ")", 2, 15)
	add(token.NEWLINE, "\n", 2, 16)

	add(token.IDENT, "Ak", 3, 1)
	add(token.DOT, ".", 3, 3)
	add(token.IDENT, "Setup", 3, 4)
	add(token.LPAREN, "(", 3, 9)
	add(token.NEWLINE, "\n", 3, 10)

	add(token.IDENT, "log", 4, 5)
	add(token.ASSIGN, "=", 4, 8)
	add(token.STRING, "info.log", 4, 9)
	add(token.COMMA, ",", 4, 19)
	add(token.NEWLINE, "\n", 4, 20)

	add(token.IDENT, "debug", 5, 5)
	add(token.ASSIGN, "=", 5, 10)
	add(token.STRING, "test.log", 5, 11)
	add(token.COMMA, ",", 5, 21)
	cm(" trailing comma allowed", 5, 23)
	add(token.NEWLINE, "\n", 5, 48)

	add(token.RPAREN, ")", 6, 1)
	add(token.NEWLINE, "\n", 6, 2)

	add(token.IDENT, "Ak", 7, 1)
	add(token.DOT, ".", 7, 3)
	add(token.IDENT, "Log", 7, 4)
	add(token.LPAREN, "(", 7, 7)
	add(token.STRING, "hello", 7, 8)
	add(token.RPAREN, ")", 7, 15)
	add(token.NEWLINE, "\n", 7, 16)

	add(token.IDENT, "Ak", 8, 1)
	add(token.DOT, ".", 8, 3)
	add(token.IDENT, "Debug", 8, 4)
	add(token.LPAREN, "(", 8, 9)
	add(token.STRING, "dbg", 8, 10)
	add(token.COMMA, ",", 8, 15)
	add(token.INT, "1", 8, 17)
	add(token.COMMA, ",", 8, 18)
	add(token.STRING, "two", 8, 20)
	add(token.RPAREN, ")", 8, 25)
	add(token.NEWLINE, "\n", 8, 26)

	cm(" work with files", 9, 1)
	add(token.NEWLINE, "\n", 9, 19)

	add(token.IDENT, "FS", 10, 1)
	add(token.DOT, ".", 10, 3)
	add(token.IDENT, "WriteFile", 10, 4)
	add(token.LPAREN, "(", 10, 13)
	add(token.STRING, "file.txt", 10, 14)
	add(token.COMMA, ",", 10, 24)
	add(token.STRING, "content", 10, 26)
	add(token.RPAREN, ")", 10, 35)
	add(token.NEWLINE, "\n", 10, 36)

	add(token.IDENT, "FS", 11, 1)
	add(token.DOT, ".", 11, 3)
	add(token.IDENT, "ReadFile", 11, 4)
	add(token.LPAREN, "(", 11, 12)
	add(token.STRING, "file.txt", 11, 13)
	add(token.RPAREN, ")", 11, 23)
	add(token.NEWLINE, "\n", 11, 24)

	add(token.IDENT, "FS", 12, 1)
	add(token.DOT, ".", 12, 3)
	add(token.IDENT, "UpdateFile", 12, 4)
	add(token.LPAREN, "(", 12, 14)
	add(token.STRING, "file.txt", 12, 15)
	add(token.COMMA, ",", 12, 25)
	add(token.STRING, "new content", 12, 27)
	add(token.RPAREN, ")", 12, 40)
	add(token.NEWLINE, "\n", 12, 41)

	cm(" work with directories", 13, 1)
	add(token.NEWLINE, "\n", 13, 25)

	add(token.IDENT, "FS", 14, 1)
	add(token.DOT, ".", 14, 3)
	add(token.IDENT, "ListFiles", 14, 4)
	add(token.LPAREN, "(", 14, 13)
	add(token.STRING, "path/", 14, 14)
	add(token.RPAREN, ")", 14, 21)
	add(token.NEWLINE, "\n", 14, 22)

	add(token.IDENT, "Ak", 15, 1)
	add(token.DOT, ".", 15, 3)
	add(token.IDENT, "Exit", 15, 4)
	add(token.LPAREN, "(", 15, 8)
	add(token.STRING, "done", 15, 9)
	add(token.COMMA, ",", 15, 15)
	add(token.IDENT, "code", 15, 17)
	add(token.ASSIGN, "=", 15, 21)
	add(token.INT, "0", 15, 22)
	add(token.RPAREN, ")", 15, 23)
	add(token.NEWLINE, "\n", 15, 24)

	add(token.EOF, "", 16, 1)
	return ts
}

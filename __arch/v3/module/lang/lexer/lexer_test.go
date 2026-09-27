package lexer

import (
	"strings"
	"testing"

	"github.com/siper92/akha/lang/tests_utils"
)

func TestLexCanTokenize(t *testing.T) {
	cases := []tests_utils.Case[string, []Kind]{
		{Name: "empty", Input: "", Expected: []Kind{EOF}},
		{
			Name:  "all_keywords",
			Input: strings.Join(LexKeywords, " "),
			Expected: []Kind{
				Let, Var, If, Else, For, In, Range,
				Break, Continue, Return, Exit,
				And, Or, Not,
				True, False, Null,
				EOF,
			},
		},
		{
			Name:  "arithmetic",
			Input: "+ - * / % == != < <= > >= =",
			Expected: []Kind{
				Plus, Minus, Star, Slash, Percent,
				Eq, NotEq, Lt, LtEq, Gt, GtEq,
				Assign, EOF,
			},
		},
		{
			Name:  "delimiters",
			Input: "( ) [ ] { } , : . ..",
			Expected: []Kind{
				LParen, RParen, LBracket, RBracket,
				LBrace, RBrace, Comma, Colon,
				Dot, DotDot, EOF,
			},
		},
		{
			Name:  "identifiers",
			Input: "foo bar baz",
			Expected: []Kind{
				Ident, Ident, Ident, EOF,
			},
		},
		{
			Name:  "numbers",
			Input: "123 45.67 0.89",
			Expected: []Kind{
				Number, Number, Number, EOF,
			},
		},
		{
			Name:  "strings",
			Input: `"hello" "world"`,
			Expected: []Kind{
				String, String, EOF,
			},
		},
		{
			Name:  "simple_object",
			Input: `{"key": "value", "num": 42}`,
			Expected: []Kind{
				LBrace, String, Colon, String, Comma,
				String, Colon, Number, RBrace,
				EOF,
			},
		},
		//  tokens with mixed content
		{
			Name:  "mixed",
			Input: `let x = 42 if x > 10 { return "big" } else { return "small" }`,
			Expected: []Kind{
				Let, Ident, Assign, Number,
				If, Ident, Gt, Number, LBrace,
				Return, String, RBrace,
				Else, LBrace,
				Return, String, RBrace,
				EOF,
			},
		},
		// special characters
		{
			Name:  "newlines",
			Input: "let x = 42\n\n\n",
			Expected: []Kind{
				Let, Ident, Assign, Number,
				Newline, Newline, Newline,
				EOF,
			},
		},
		{
			Name:  "whitespace + mixed newlines",
			Input: "let a = 1\n\nif a {\n\r\n    let b = \"x\"\n}\r\n",
			Expected: []Kind{
				Let, Ident, Assign, Number,
				Newline, Newline,
				If, Ident, LBrace,
				Newline, Newline,
				Let, Ident, Assign, String,
				Newline,
				RBrace,
				Newline,
				EOF,
			},
		},
		// comments
		{
			Name:  "comments", // comments are ignored by the lexer
			Input: "let x = 42  // this is a comment\nx = x + 1",
			Expected: []Kind{
				Let, Ident, Assign, Number,
				Newline,
				Ident, Assign, Ident, Plus, Number,
				EOF,
			},
		},
		// template strings
		{
			Name:  "template_string",
			Input: `"Hello, ${name}!"`,
			Expected: []Kind{
				Template, EOF,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			validateTokenizerOutput(t, c)
		})
	}
}

func validateTokenizerOutput(t *testing.T, c tests_utils.Case[string, []Kind]) {
	t.Helper()
	toks, err := Tokenize(c.Input)
	if err != nil {
		t.Fatalf("tokenize %q: %v", c.Input, err)
	}

	for i, tok := range toks {
		if c.Expected[i] != tok.Kind {
			t.Errorf(
				"token kind mismatch at index %d: got %v, want %v",
				i, tok.Kind, c.Expected[i],
			)
		}
	}
}

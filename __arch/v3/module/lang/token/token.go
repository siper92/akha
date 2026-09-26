package token

import "strconv"

type Kind int

const (
	EOF Kind = iota
	Newline
	Ident
	Number
	String
	Template

	Let
	Var
	If
	Else
	For
	In
	Range
	Break
	Continue
	Return
	Exit
	And
	Or
	Not
	True
	False
	Null

	Reserved

	Plus
	Minus
	Star
	Slash
	Percent
	Eq
	NotEq
	Lt
	LtEq
	Gt
	GtEq
	Assign

	LParen
	RParen
	LBracket
	RBracket
	LBrace
	RBrace
	Comma
	Colon
	Dot
	DotDot

	NotIn
)

var names = [...]string{
	EOF:      "EOF",
	Newline:  "newline",
	Ident:    "identifier",
	Number:   "number",
	String:   "string",
	Template: "template",

	Let:      "let",
	Var:      "var",
	If:       "if",
	Else:     "else",
	For:      "for",
	In:       "in",
	Range:    "range",
	Break:    "break",
	Continue: "continue",
	Return:   "return",
	Exit:     "exit",
	And:      "and",
	Or:       "or",
	Not:      "not",
	True:     "true",
	False:    "false",
	Null:     "null",

	Reserved: "reserved",

	Plus:    "+",
	Minus:   "-",
	Star:    "*",
	Slash:   "/",
	Percent: "%",
	Eq:      "==",
	NotEq:   "!=",
	Lt:      "<",
	LtEq:    "<=",
	Gt:      ">",
	GtEq:    ">=",
	Assign:  "=",

	LParen:   "(",
	RParen:   ")",
	LBracket: "[",
	RBracket: "]",
	LBrace:   "{",
	RBrace:   "}",
	Comma:    ",",
	Colon:    ":",
	Dot:      ".",
	DotDot:   "..",

	NotIn: "not in",
}

var keywords = func() map[string]Kind {
	m := make(map[string]Kind)
	for k := Let; k <= Null; k++ {
		m[names[k]] = k
	}
	return m
}()

var reserved = map[string]bool{
	"fn":    true,
	"try":   true,
	"catch": true,
}

func (k Kind) String() string {
	if k >= 0 && int(k) < len(names) && names[k] != "" {
		return names[k]
	}
	return "Kind(" + strconv.Itoa(int(k)) + ")"
}

func (k Kind) IsKeyword() bool {
	return k >= Let && k <= Null
}

func Lookup(ident string) Kind {
	if k, ok := keywords[ident]; ok {
		return k
	}
	if reserved[ident] {
		return Reserved
	}
	return Ident
}

type Pos struct {
	Line int
	Col  int
}

type Part struct {
	Text   string
	Expr   string
	IsExpr bool
	Pos    Pos
}

type Token struct {
	Kind  Kind
	Lit   string
	Pos   Pos
	Parts []Part
}

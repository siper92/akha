package token

type Kind int

const (
	ILLEGAL Kind = iota
	EOF
	NEWLINE
	COMMENT
	IDENT
	STRING
	INT
	DOT
	LPAREN
	RPAREN
	COMMA
	ASSIGN
	ELLIPSIS
)

type Pos struct {
	Line int
	Col  int
}

type Token struct {
	Kind Kind
	Lit  string
	Pos  Pos
}

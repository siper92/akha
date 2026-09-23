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
	LBRACE
	RBRACE
	PLUS
	MINUS
	STAR
	SLASH
	PERCENT
	EQ
	NEQ
	LT
	LTE
	GT
	GTE
	LET
	IF
	ELSE
	FOR
	IN
	WHILE
	BREAK
	CONTINUE
	AND
	OR
	NOT
	TRUE
	FALSE
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

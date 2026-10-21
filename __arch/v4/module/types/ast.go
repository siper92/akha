package types

type TokenType int

const (
	TokenTypeUnknown TokenType = iota
	TokenTypeAccess
	TokenTypeRefresh
)

type Token interface {
	Validate() error
	Raw() string
	Type() TokenType
}

type ASTTree interface {
	Parse(source string) error
	Validate() error
	Nodes() []ASTNode
}

type ASTNode interface {
	Validate() error
	Value() DataUnit
}

package lexer

import (
	"iter"

	"github.com/siper92/akha/lang/token"
)

type Lexer interface {
	Next() token.Token
	All() iter.Seq[token.Token]
}

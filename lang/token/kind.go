package token

import (
	"fmt"
	"strconv"
)

var _ fmt.Stringer = Kind(0)

var kindNames = [...]string{
	ILLEGAL:  "illegal",
	EOF:      "eof",
	NEWLINE:  "newline",
	COMMENT:  "comment",
	IDENT:    "ident",
	STRING:   "string",
	INT:      "int",
	DOT:      "dot",
	LPAREN:   "lparen",
	RPAREN:   "rparen",
	COMMA:    "comma",
	ASSIGN:   "assign",
	ELLIPSIS: "ellipsis",
}

func (k Kind) String() string {
	if k >= 0 && int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind(" + strconv.Itoa(int(k)) + ")"
}

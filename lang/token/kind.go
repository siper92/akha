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
	LBRACE:   "lbrace",
	RBRACE:   "rbrace",
	PLUS:     "plus",
	MINUS:    "minus",
	STAR:     "star",
	SLASH:    "slash",
	PERCENT:  "percent",
	EQ:       "eq",
	NEQ:      "neq",
	LT:       "lt",
	LTE:      "lte",
	GT:       "gt",
	GTE:      "gte",
	LET:      "let",
	IF:       "if",
	ELSE:     "else",
	FOR:      "for",
	IN:       "in",
	WHILE:    "while",
	BREAK:    "break",
	CONTINUE: "continue",
	AND:      "and",
	OR:       "or",
	NOT:      "not",
	TRUE:     "true",
	FALSE:    "false",
}

var keywords = map[string]Kind{
	"let":      LET,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"in":       IN,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"true":     TRUE,
	"false":    FALSE,
}

func (k Kind) String() string {
	if k >= 0 && int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind(" + strconv.Itoa(int(k)) + ")"
}

func Lookup(ident string) Kind {
	if k, ok := keywords[ident]; ok {
		return k
	}
	return IDENT
}

package ast

import "github.com/siper92/akha/lang/token"

type Node interface {
	Pos() token.Pos
}

type Expr interface {
	Node
	expr()
}

type Script struct {
	Calls []*Call
}

type Call struct {
	Target Selector
	Args   []Expr
	Kwargs []Kwarg
	P      token.Pos
}

type Selector struct {
	Module string
	Name   string
	P      token.Pos
}

type Kwarg struct {
	Name  string
	Value Expr
	P     token.Pos
}

type Literal struct {
	Kind  token.Kind
	Value string
	P     token.Pos
}

type Spread struct {
	Module string
	P      token.Pos
}

var (
	_ Node = (*Script)(nil)
	_ Node = (*Call)(nil)
	_ Node = (*Selector)(nil)
	_ Node = (*Kwarg)(nil)
	_ Expr = (*Literal)(nil)
	_ Expr = (*Spread)(nil)
)

func (s *Script) Pos() token.Pos {
	if len(s.Calls) == 0 {
		return token.Pos{}
	}
	return s.Calls[0].P
}

func (c *Call) Pos() token.Pos     { return c.P }
func (s *Selector) Pos() token.Pos { return s.P }
func (k *Kwarg) Pos() token.Pos    { return k.P }
func (l *Literal) Pos() token.Pos  { return l.P }
func (s *Spread) Pos() token.Pos   { return s.P }

func (*Literal) expr() {}
func (*Spread) expr()  {}

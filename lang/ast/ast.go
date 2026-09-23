package ast

import "github.com/siper92/akha/lang/token"

type Node interface {
	Pos() token.Pos
}

type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
}

type Script struct {
	Stmts []Stmt
}

type Block struct {
	Stmts []Stmt
	P     token.Pos
}

type CallStmt struct {
	Call *Call
}

type Let struct {
	Name  string
	Value Expr
	P     token.Pos
}

type Assign struct {
	Name  string
	Value Expr
	P     token.Pos
}

type If struct {
	Cond Expr
	Then *Block
	Else Stmt
	P    token.Pos
}

type For struct {
	Var  string
	Iter Expr
	Body *Block
	P    token.Pos
}

type While struct {
	Cond Expr
	Body *Block
	P    token.Pos
}

type Break struct {
	P token.Pos
}

type Continue struct {
	P token.Pos
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

type Ident struct {
	Name string
	P    token.Pos
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

type Binary struct {
	Op token.Kind
	X  Expr
	Y  Expr
	P  token.Pos
}

type Unary struct {
	Op token.Kind
	X  Expr
	P  token.Pos
}

var (
	_ Node = (*Script)(nil)
	_ Node = (*Block)(nil)
	_ Node = (*Selector)(nil)
	_ Node = (*Kwarg)(nil)
	_ Stmt = (*CallStmt)(nil)
	_ Stmt = (*Let)(nil)
	_ Stmt = (*Assign)(nil)
	_ Stmt = (*If)(nil)
	_ Stmt = (*For)(nil)
	_ Stmt = (*While)(nil)
	_ Stmt = (*Break)(nil)
	_ Stmt = (*Continue)(nil)
	_ Expr = (*Call)(nil)
	_ Expr = (*Ident)(nil)
	_ Expr = (*Literal)(nil)
	_ Expr = (*Spread)(nil)
	_ Expr = (*Binary)(nil)
	_ Expr = (*Unary)(nil)
)

func (s *Script) Pos() token.Pos {
	if len(s.Stmts) == 0 {
		return token.Pos{}
	}
	return s.Stmts[0].Pos()
}

func (b *Block) Pos() token.Pos    { return b.P }
func (c *CallStmt) Pos() token.Pos { return c.Call.P }
func (l *Let) Pos() token.Pos      { return l.P }
func (a *Assign) Pos() token.Pos   { return a.P }
func (i *If) Pos() token.Pos       { return i.P }
func (f *For) Pos() token.Pos      { return f.P }
func (w *While) Pos() token.Pos    { return w.P }
func (b *Break) Pos() token.Pos    { return b.P }
func (c *Continue) Pos() token.Pos { return c.P }
func (c *Call) Pos() token.Pos     { return c.P }
func (s *Selector) Pos() token.Pos { return s.P }
func (k *Kwarg) Pos() token.Pos    { return k.P }
func (i *Ident) Pos() token.Pos    { return i.P }
func (l *Literal) Pos() token.Pos  { return l.P }
func (s *Spread) Pos() token.Pos   { return s.P }
func (b *Binary) Pos() token.Pos   { return b.P }
func (u *Unary) Pos() token.Pos    { return u.P }

func (*CallStmt) stmt() {}
func (*Let) stmt()      {}
func (*Assign) stmt()   {}
func (*If) stmt()       {}
func (*For) stmt()      {}
func (*While) stmt()    {}
func (*Break) stmt()    {}
func (*Continue) stmt() {}

func (*Call) expr()    {}
func (*Ident) expr()   {}
func (*Literal) expr() {}
func (*Spread) expr()  {}
func (*Binary) expr()  {}
func (*Unary) expr()   {}

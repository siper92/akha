package ast

import (
	"fmt"

	"github.com/siper92/akha/lang/token"
)

type Pos struct {
	Line int
}

func (p Pos) Position() Pos {
	return p
}

type Node interface {
	fmt.Stringer
	Position() Pos
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
	Pos
	Stmts []Stmt
}

type LetStmt struct {
	Pos
	Name  string
	Value Expr
}

type VarStmt struct {
	Pos
	Name  string
	Value Expr
}

type AssignStmt struct {
	Pos
	Target Expr
	Value  Expr
}

type ExprStmt struct {
	Pos
	X Expr
}

type IfStmt struct {
	Pos
	Cond Expr
	Then *Block
	Else Node
}

type ForInStmt struct {
	Pos
	Mutable bool
	Key     string
	Value   string
	Iter    Expr
	Body    *Block
}

type ForRangeStmt struct {
	Pos
	Mutable bool
	Name    string
	Start   Expr
	End     Expr
	Body    *Block
}

type BreakStmt struct {
	Pos
}

type ContinueStmt struct {
	Pos
}

type ReturnStmt struct {
	Pos
	Exit  bool
	Value Expr
}

type Ident struct {
	Pos
	Name string
}

type NumberLit struct {
	Pos
	Value float64
	Raw   string
}

type StringLit struct {
	Pos
	Value string
}

type TemplatePart struct {
	Text string
	Expr Expr
}

type TemplateLit struct {
	Pos
	Parts []TemplatePart
}

type BoolLit struct {
	Pos
	Value bool
}

type NullLit struct {
	Pos
}

type ArrayLit struct {
	Pos
	Elems []Expr
}

type Entry struct {
	Pos
	Key   string
	Value Expr
}

type ObjectLit struct {
	Pos
	Entries []Entry
}

type UnaryExpr struct {
	Pos
	Op token.Kind
	X  Expr
}

type BinaryExpr struct {
	Pos
	Op    token.Kind
	Left  Expr
	Right Expr
}

type MemberExpr struct {
	Pos
	X    Expr
	Name string
}

type IndexExpr struct {
	Pos
	X     Expr
	Index Expr
}

type CallExpr struct {
	Pos
	Fn   Expr
	Args []Expr
}

var (
	_ Node = (*Block)(nil)

	_ Stmt = (*LetStmt)(nil)
	_ Stmt = (*VarStmt)(nil)
	_ Stmt = (*AssignStmt)(nil)
	_ Stmt = (*ExprStmt)(nil)
	_ Stmt = (*IfStmt)(nil)
	_ Stmt = (*ForInStmt)(nil)
	_ Stmt = (*ForRangeStmt)(nil)
	_ Stmt = (*BreakStmt)(nil)
	_ Stmt = (*ContinueStmt)(nil)
	_ Stmt = (*ReturnStmt)(nil)

	_ Expr = (*Ident)(nil)
	_ Expr = (*NumberLit)(nil)
	_ Expr = (*StringLit)(nil)
	_ Expr = (*TemplateLit)(nil)
	_ Expr = (*BoolLit)(nil)
	_ Expr = (*NullLit)(nil)
	_ Expr = (*ArrayLit)(nil)
	_ Expr = (*ObjectLit)(nil)
	_ Expr = (*UnaryExpr)(nil)
	_ Expr = (*BinaryExpr)(nil)
	_ Expr = (*MemberExpr)(nil)
	_ Expr = (*IndexExpr)(nil)
	_ Expr = (*CallExpr)(nil)
)

func (*LetStmt) stmt()      {}
func (*VarStmt) stmt()      {}
func (*AssignStmt) stmt()   {}
func (*ExprStmt) stmt()     {}
func (*IfStmt) stmt()       {}
func (*ForInStmt) stmt()    {}
func (*ForRangeStmt) stmt() {}
func (*BreakStmt) stmt()    {}
func (*ContinueStmt) stmt() {}
func (*ReturnStmt) stmt()   {}

func (*Ident) expr()       {}
func (*NumberLit) expr()   {}
func (*StringLit) expr()   {}
func (*TemplateLit) expr() {}
func (*BoolLit) expr()     {}
func (*NullLit) expr()     {}
func (*ArrayLit) expr()    {}
func (*ObjectLit) expr()   {}
func (*UnaryExpr) expr()   {}
func (*BinaryExpr) expr()  {}
func (*MemberExpr) expr()  {}
func (*IndexExpr) expr()   {}
func (*CallExpr) expr()    {}

package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/siper92/akha/lang/lexer"
)

var (
	_ fmt.Stringer = (*Script)(nil)
	_ fmt.Stringer = (*Block)(nil)
	_ fmt.Stringer = (*Entry)(nil)
	_ fmt.Stringer = (*TemplatePart)(nil)
	_ fmt.Stringer = (*Kwarg)(nil)

	_ fmt.Stringer = (*LetStmt)(nil)
	_ fmt.Stringer = (*VarStmt)(nil)
	_ fmt.Stringer = (*AssignStmt)(nil)
	_ fmt.Stringer = (*ExprStmt)(nil)
	_ fmt.Stringer = (*IfStmt)(nil)
	_ fmt.Stringer = (*ForInStmt)(nil)
	_ fmt.Stringer = (*ForRangeStmt)(nil)
	_ fmt.Stringer = (*BreakStmt)(nil)
	_ fmt.Stringer = (*ContinueStmt)(nil)
	_ fmt.Stringer = (*ReturnStmt)(nil)

	_ fmt.Stringer = (*Ident)(nil)
	_ fmt.Stringer = (*NumberLit)(nil)
	_ fmt.Stringer = (*StringLit)(nil)
	_ fmt.Stringer = (*TemplateLit)(nil)
	_ fmt.Stringer = (*BoolLit)(nil)
	_ fmt.Stringer = (*NullLit)(nil)
	_ fmt.Stringer = (*ArrayLit)(nil)
	_ fmt.Stringer = (*ObjectLit)(nil)
	_ fmt.Stringer = (*UnaryExpr)(nil)
	_ fmt.Stringer = (*BinaryExpr)(nil)
	_ fmt.Stringer = (*MemberExpr)(nil)
	_ fmt.Stringer = (*IndexExpr)(nil)
	_ fmt.Stringer = (*CallExpr)(nil)
)

const indent = "    "

const (
	precOr = iota + 1
	precAnd
	precNot
	precIn
	precCmp
	precAdd
	precMul
	precNeg
	precPostfix
)

type printer struct {
	b     strings.Builder
	depth int
}

func (s *Script) String() string {
	p := &printer{}
	for i, st := range s.Stmts {
		if i > 0 {
			p.write("\n")
		}
		p.stmt(st)
	}
	return p.b.String()
}

func (b *Block) String() string {
	p := &printer{}
	p.block(b)
	return p.b.String()
}

func (s *LetStmt) String() string      { return stmtString(s) }
func (s *VarStmt) String() string      { return stmtString(s) }
func (s *AssignStmt) String() string   { return stmtString(s) }
func (s *ExprStmt) String() string     { return stmtString(s) }
func (s *IfStmt) String() string       { return stmtString(s) }
func (s *ForInStmt) String() string    { return stmtString(s) }
func (s *ForRangeStmt) String() string { return stmtString(s) }
func (s *BreakStmt) String() string    { return stmtString(s) }
func (s *ContinueStmt) String() string { return stmtString(s) }
func (s *ReturnStmt) String() string   { return stmtString(s) }

func (x *Ident) String() string {
	return x.Name
}

func (x *NumberLit) String() string {
	return strconv.FormatFloat(x.Value, 'f', -1, 64)
}

func (x *StringLit) String() string {
	return `"` + escape(x.Value) + `"`
}

func (x *TemplateLit) String() string {
	var b strings.Builder
	b.WriteByte('"')
	for _, part := range x.Parts {
		b.WriteString(part.String())
	}
	b.WriteByte('"')
	return b.String()
}

func (t TemplatePart) String() string {
	if t.Expr != nil {
		return "${" + t.Expr.String() + "}"
	}
	return escape(t.Text)
}

func (x *BoolLit) String() string {
	return strconv.FormatBool(x.Value)
}

func (x *NullLit) String() string {
	return "null"
}

func (x *ArrayLit) String() string {
	return "[" + join(x.Elems) + "]"
}

func (e Entry) String() string {
	return key(e.Key) + ": " + str(e.Value)
}

func (x *ObjectLit) String() string {
	items := make([]string, len(x.Entries))
	for i, e := range x.Entries {
		items[i] = e.String()
	}
	return "{" + strings.Join(items, ", ") + "}"
}

func (x *UnaryExpr) String() string {
	if x.Op == lexer.Not {
		return "not " + group(x.X, precNot)
	}
	operand := group(x.X, precNeg)
	if strings.HasPrefix(operand, "-") {
		return "- " + operand
	}
	return "-" + operand
}

func (x *BinaryExpr) String() string {
	p := binaryPrec(x.Op)
	left := p
	if p == precIn || p == precCmp {
		left = p + 1
	}
	return group(x.Left, left) + " " + x.Op.String() + " " + group(x.Right, p+1)
}

func (x *MemberExpr) String() string {
	return group(x.X, precPostfix) + "." + x.Name
}

func (x *IndexExpr) String() string {
	return group(x.X, precPostfix) + "[" + str(x.Index) + "]"
}

func (k Kwarg) String() string {
	return k.Name + "=" + str(k.Value)
}

func (x *CallExpr) String() string {
	items := make([]string, 0, len(x.Args)+len(x.Kwargs))
	for _, a := range x.Args {
		items = append(items, str(a))
	}

	for _, k := range x.Kwargs {
		items = append(items, k.String())
	}

	return group(x.Fn, precPostfix) + "(" + strings.Join(items, ", ") + ")"
}

func stmtString(s Stmt) string {
	p := &printer{}
	p.stmt(s)
	return p.b.String()
}

func (p *printer) write(parts ...string) {
	for _, s := range parts {
		p.b.WriteString(s)
	}
}

func (p *printer) stmt(s Stmt) {
	switch s := s.(type) {
	case *LetStmt:
		p.write("let ", s.Name, " = ", str(s.Value))
	case *VarStmt:
		p.write("var ", s.Name)
		if s.Value != nil {
			p.write(" = ", str(s.Value))
		}
	case *AssignStmt:
		p.write(str(s.Target), " = ", str(s.Value))
	case *ExprStmt:
		p.write(str(s.X))
	case *IfStmt:
		p.ifStmt(s)
	case *ForInStmt:
		p.write("for ", varKw(s.Mutable))
		if s.Key != "" {
			p.write(s.Key, ", ")
		}
		p.write(s.Value, " in ", header(s.Iter), " ")
		p.block(s.Body)
	case *ForRangeStmt:
		p.write("for ", varKw(s.Mutable), s.Name, " range [", str(s.Start), "..", str(s.End), "] ")
		p.block(s.Body)
	case *BreakStmt:
		p.write("break")
	case *ContinueStmt:
		p.write("continue")
	case *ReturnStmt:
		if s.Exit {
			p.write("exit")
		} else {
			p.write("return")
		}
		if s.Value != nil {
			p.write(" ", str(s.Value))
		}
	}
}

func (p *printer) ifStmt(s *IfStmt) {
	p.write("if ", header(s.Cond), " ")
	p.block(s.Then)
	switch e := s.Else.(type) {
	case *IfStmt:
		p.write(" else ")
		p.ifStmt(e)
	case *Block:
		p.write(" else ")
		p.block(e)
	}
}

func (p *printer) block(b *Block) {
	p.write("{\n")
	if b != nil {
		p.depth++
		for _, s := range b.Stmts {
			p.write(strings.Repeat(indent, p.depth))
			p.stmt(s)
			p.write("\n")
		}
		p.depth--
	}
	p.write(strings.Repeat(indent, p.depth), "}")
}

func str(x Expr) string {
	if x == nil {
		return ""
	}
	return x.String()
}

func join(xs []Expr) string {
	items := make([]string, len(xs))
	for i, x := range xs {
		items[i] = str(x)
	}
	return strings.Join(items, ", ")
}

func header(x Expr) string {
	s := str(x)
	if strings.HasPrefix(s, "{") {
		return "(" + s + ")"
	}
	return s
}

func group(x Expr, min int) string {
	s := str(x)
	if prec(x) < min {
		return "(" + s + ")"
	}
	return s
}

func prec(x Expr) int {
	switch x := x.(type) {
	case *BinaryExpr:
		return binaryPrec(x.Op)
	case *UnaryExpr:
		if x.Op == lexer.Not {
			return precNot
		}
		return precNeg
	}
	return precPostfix
}

func binaryPrec(op lexer.Kind) int {
	switch op {
	case lexer.Or:
		return precOr
	case lexer.And:
		return precAnd
	case lexer.In, lexer.NotIn:
		return precIn
	case lexer.Eq, lexer.NotEq, lexer.Lt, lexer.LtEq, lexer.Gt, lexer.GtEq:
		return precCmp
	case lexer.Plus, lexer.Minus:
		return precAdd
	case lexer.Star, lexer.Slash, lexer.Percent:
		return precMul
	}
	return precPostfix
}

func varKw(m bool) string {
	if m {
		return "var "
	}
	return ""
}

func key(k string) string {
	if isName(k) {
		return k
	}
	return `"` + escape(k) + `"`
}

func isName(s string) bool {
	if s == "" || lexer.Lookup(s) != lexer.Ident {
		return false
	}
	for i, r := range s {
		switch {
		case r == '_', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case i > 0 && r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

func escape(s string) string {
	var b strings.Builder
	for i, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\t':
			b.WriteString(`\t`)
		case '\r':
			b.WriteString(`\r`)
		case '$':
			if strings.HasPrefix(s[i+1:], "{") {
				b.WriteString(`\$`)
				continue
			}
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

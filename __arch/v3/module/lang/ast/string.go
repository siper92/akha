package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/siper92/akha/lang/token"
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

var (
	_ fmt.Stringer = (*Script)(nil)
	_ fmt.Stringer = Entry{}
	_ fmt.Stringer = TemplatePart{}
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
	if x.Op == token.Not {
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

func (x *CallExpr) String() string {
	return group(x.Fn, precPostfix) + "(" + join(x.Args) + ")"
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
		if x.Op == token.Not {
			return precNot
		}
		return precNeg
	}
	return precPostfix
}

func binaryPrec(op token.Kind) int {
	switch op {
	case token.Or:
		return precOr
	case token.And:
		return precAnd
	case token.In, token.NotIn:
		return precIn
	case token.Eq, token.NotEq, token.Lt, token.LtEq, token.Gt, token.GtEq:
		return precCmp
	case token.Plus, token.Minus:
		return precAdd
	case token.Star, token.Slash, token.Percent:
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
	if s == "" || token.Lookup(strings.ToLower(s)) != token.Ident {
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

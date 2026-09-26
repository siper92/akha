package ast

import (
	"strconv"
	"strings"
)

func Sprint(n any) string {
	var b strings.Builder
	write(&b, n)
	return b.String()
}

func write(b *strings.Builder, n any) {
	switch n := n.(type) {
	case *Script:
		for i, s := range n.Stmts {
			if i > 0 {
				b.WriteByte('\n')
			}
			write(b, s)
		}
	case *Block:
		b.WriteByte('{')
		for i, s := range n.Stmts {
			if i > 0 {
				b.WriteByte(' ')
			}
			write(b, s)
		}
		b.WriteByte('}')
	case *LetStmt:
		sexp(b, "let", n.Name, n.Value)
	case *VarStmt:
		sexp(b, "var", n.Name, n.Value)
	case *AssignStmt:
		sexp(b, "=", n.Target, n.Value)
	case *ExprStmt:
		write(b, n.X)
	case *IfStmt:
		if n.Else == nil {
			sexp(b, "if", n.Cond, n.Then)
			return
		}
		sexp(b, "if", n.Cond, n.Then, "else", n.Else)
	case *ForInStmt:
		sexp(b, "for", mutable(n.Mutable), n.Key, n.Value, "in", n.Iter, n.Body)
	case *ForRangeStmt:
		sexp(b, "for", mutable(n.Mutable), n.Name, "range", n.Start, n.End, n.Body)
	case *BreakStmt:
		sexp(b, "break")
	case *ContinueStmt:
		sexp(b, "continue")
	case *ReturnStmt:
		head := "return"
		if n.Exit {
			head = "exit"
		}
		sexp(b, head, n.Value)
	case *Ident:
		b.WriteString(n.Name)
	case *NumberLit:
		b.WriteString(strconv.FormatFloat(n.Value, 'f', -1, 64))
	case *StringLit:
		b.WriteString(strconv.Quote(n.Value))
	case *TemplateLit:
		items := make([]any, 0, len(n.Parts))
		for _, part := range n.Parts {
			if part.Expr != nil {
				items = append(items, part.Expr)
				continue
			}
			items = append(items, strconv.Quote(part.Text))
		}
		sexp(b, "tpl", items...)
	case *BoolLit:
		b.WriteString(strconv.FormatBool(n.Value))
	case *NullLit:
		b.WriteString("null")
	case *ArrayLit:
		b.WriteByte('[')
		for i, e := range n.Elems {
			if i > 0 {
				b.WriteByte(' ')
			}
			write(b, e)
		}
		b.WriteByte(']')
	case *ObjectLit:
		b.WriteString("(obj")
		for _, e := range n.Entries {
			b.WriteByte(' ')
			b.WriteString(strconv.Quote(e.Key))
			b.WriteByte(':')
			write(b, e.Value)
		}
		b.WriteByte(')')
	case *UnaryExpr:
		sexp(b, n.Op.String(), n.X)
	case *BinaryExpr:
		sexp(b, n.Op.String(), n.Left, n.Right)
	case *MemberExpr:
		sexp(b, ".", n.X, n.Name)
	case *IndexExpr:
		sexp(b, "[]", n.X, n.Index)
	case *CallExpr:
		items := []any{n.Fn}
		for _, a := range n.Args {
			items = append(items, a)
		}
		sexp(b, "call", items...)
	}
}

func sexp(b *strings.Builder, head string, items ...any) {
	b.WriteByte('(')
	b.WriteString(head)
	for _, it := range items {
		if it == nil || it == "" {
			continue
		}
		b.WriteByte(' ')
		if s, ok := it.(string); ok {
			b.WriteString(s)
			continue
		}
		write(b, it)
	}
	b.WriteByte(')')
}

func mutable(m bool) string {
	if m {
		return "var"
	}
	return ""
}

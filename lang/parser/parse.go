package parser

import (
	"fmt"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/token"
)

type parser struct {
	lx    lexer.Lexer
	tok   token.Token
	ahead *token.Token
	errs  Errors
}

var _ Parser = (*parser)(nil)

func New(lx lexer.Lexer) Parser {
	return &parser{lx: lx}
}

func (p *parser) Parse() (*ast.Script, error) {
	p.next()
	s := &ast.Script{}
	for p.tok.Kind != token.EOF {
		if p.tok.Kind == token.NEWLINE {
			p.next()
			continue
		}
		st, ok := p.stmt()
		if !ok {
			p.sync()
			continue
		}
		s.Stmts = append(s.Stmts, st)
		if !p.endStmt() {
			p.sync()
		}
	}
	if len(p.errs) == 0 {
		return s, nil
	}

	return s, p.errs
}

func (p *parser) read() token.Token {
	for {
		t := p.lx.Next()
		if t.Kind != token.COMMENT {
			return t
		}
	}
}

func (p *parser) next() {
	if p.ahead != nil {
		p.tok = *p.ahead
		p.ahead = nil
		return
	}
	p.tok = p.read()
}

func (p *parser) peek() token.Token {
	if p.ahead == nil {
		t := p.read()
		p.ahead = &t
	}
	return *p.ahead
}

func (p *parser) skipNewlines() {
	for p.tok.Kind == token.NEWLINE {
		p.next()
	}
}

func (p *parser) sync() {
	for p.tok.Kind != token.NEWLINE && p.tok.Kind != token.EOF {
		p.next()
	}
	if p.tok.Kind == token.NEWLINE {
		p.next()
	}
}

func (p *parser) errorf(pos token.Pos, format string, args ...any) {
	p.errs = append(p.errs, Error{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

func (p *parser) expect(kind token.Kind, what string) (token.Token, bool) {
	if p.tok.Kind != kind {
		p.errorf(p.tok.Pos, "expected %s, got %s", what, describe(p.tok))
		return p.tok, false
	}
	t := p.tok
	p.next()
	return t, true
}

func (p *parser) endStmt() bool {
	switch p.tok.Kind {
	case token.NEWLINE:
		p.next()
		return true
	case token.EOF:
		return true
	default:
		p.errorf(p.tok.Pos, "expected newline after statement, got %s", describe(p.tok))
		return false
	}
}

func (p *parser) stmt() (ast.Stmt, bool) {
	switch p.tok.Kind {
	case token.IDENT:
		return p.identStmt()
	case token.LET:
		return p.let()
	case token.IF:
		return p.notImplemented()
	case token.FOR:
		return p.notImplemented()
	case token.WHILE:
		return p.notImplemented()
	case token.BREAK:
		return p.notImplemented()
	case token.CONTINUE:
		return p.notImplemented()
	case token.STRING, token.INT, token.TRUE, token.FALSE, token.LPAREN, token.MINUS, token.NOT:
		p.errorf(p.tok.Pos, "expression is not a statement")
		return nil, false
	default:
		p.errorf(p.tok.Pos, "expected statement, got %s", describe(p.tok))
		return nil, false
	}
}

func (p *parser) notImplemented() (ast.Stmt, bool) {
	p.errorf(p.tok.Pos, "%s is not implemented", p.tok.Kind)
	return nil, false
}

func (p *parser) identStmt() (ast.Stmt, bool) {
	switch p.peek().Kind {
	case token.DOT:
		c, ok := p.call()
		if !ok {
			return nil, false
		}
		return &ast.CallStmt{Call: c}, true
	case token.ASSIGN:
		return p.assign()
	default:
		p.next()
		p.errorf(p.tok.Pos, "expected '.' or '=', got %s", describe(p.tok))
		return nil, false
	}
}

func (p *parser) let() (ast.Stmt, bool) {
	kw := p.tok
	p.next()
	name, ok := p.expect(token.IDENT, "variable name")
	if !ok {
		return nil, false
	}
	if _, ok = p.expect(token.ASSIGN, "'='"); !ok {
		return nil, false
	}
	v, ok := p.expr()
	if !ok {
		return nil, false
	}
	return &ast.Let{Name: name.Lit, Value: v, P: kw.Pos}, true
}

func (p *parser) assign() (ast.Stmt, bool) {
	name := p.tok
	p.next()
	p.next()
	v, ok := p.expr()
	if !ok {
		return nil, false
	}
	return &ast.Assign{Name: name.Lit, Value: v, P: name.Pos}, true
}

func (p *parser) block() (*ast.Block, bool) {
	lb, ok := p.expect(token.LBRACE, "'{'")
	if !ok {
		return nil, false
	}
	if _, ok = p.expect(token.NEWLINE, "newline after '{'"); !ok {
		return nil, false
	}
	b := &ast.Block{P: lb.Pos}
	for {
		switch p.tok.Kind {
		case token.NEWLINE:
			p.next()
			continue
		case token.RBRACE:
			p.next()
			return b, true
		case token.EOF:
			p.errorf(p.tok.Pos, "unexpected end of file, expected '}'")
			return nil, false
		}
		st, ok := p.stmt()
		if !ok {
			p.sync()
			continue
		}
		b.Stmts = append(b.Stmts, st)
		if !p.endStmt() {
			p.sync()
		}
	}
}

func (p *parser) call() (*ast.Call, bool) {
	mod, ok := p.expect(token.IDENT, "module name")
	if !ok {
		return nil, false
	}
	if _, ok = p.expect(token.DOT, "'.'"); !ok {
		return nil, false
	}
	name, ok := p.expect(token.IDENT, "function name")
	if !ok {
		return nil, false
	}
	if _, ok = p.expect(token.LPAREN, "'('"); !ok {
		return nil, false
	}
	c := &ast.Call{
		Target: ast.Selector{Module: mod.Lit, Name: name.Lit, P: mod.Pos},
		P:      mod.Pos,
	}
	if !p.args(c) {
		return nil, false
	}
	return c, true
}

func (p *parser) args(c *ast.Call) bool {
	seen := map[string]bool{}
	for {
		p.skipNewlines()
		if p.tok.Kind == token.RPAREN {
			p.next()
			return true
		}
		if p.tok.Kind == token.EOF {
			p.errorf(p.tok.Pos, "unexpected end of file, expected ')'")
			return false
		}
		if !p.arg(c, seen) {
			return false
		}
		p.skipNewlines()
		switch p.tok.Kind {
		case token.COMMA:
			p.next()
		case token.RPAREN:
			p.next()
			return true
		case token.EOF:
			p.errorf(p.tok.Pos, "unexpected end of file, expected ')'")
			return false
		default:
			p.errorf(p.tok.Pos, "expected ')' or ',', got %s", describe(p.tok))
			return false
		}
	}
}

func (p *parser) arg(c *ast.Call, seen map[string]bool) bool {
	if p.tok.Kind == token.IDENT {
		id := p.tok
		switch p.peek().Kind {
		case token.ASSIGN:
			p.next()
			p.next()
			p.skipNewlines()
			v, ok := p.expr()
			if !ok {
				return false
			}
			if seen[id.Lit] {
				p.errorf(id.Pos, "duplicate argument %s", id.Lit)
			}
			seen[id.Lit] = true
			c.Kwargs = append(c.Kwargs, ast.Kwarg{Name: id.Lit, Value: v, P: id.Pos})
			return true
		case token.ELLIPSIS:
			p.next()
			p.next()
			if c.Target.Module != "Ak" || c.Target.Name != "Allow" {
				p.errorf(id.Pos, "spread only allowed in Ak.Allow")
			}
			p.positional(c, &ast.Spread{Module: id.Lit, P: id.Pos})
			return true
		}
	}
	v, ok := p.expr()
	if !ok {
		return false
	}
	p.positional(c, v)
	return true
}

func (p *parser) positional(c *ast.Call, e ast.Expr) {
	if len(c.Kwargs) > 0 {
		p.errorf(e.Pos(), "positional argument after keyword argument")
	}
	c.Args = append(c.Args, e)
}

func (p *parser) expr() (ast.Expr, bool) {
	return p.or()
}

func (p *parser) or() (ast.Expr, bool) {
	return p.binary(p.and, false, token.OR)
}

func (p *parser) and() (ast.Expr, bool) {
	return p.binary(p.not, false, token.AND)
}

func (p *parser) not() (ast.Expr, bool) {
	if p.tok.Kind != token.NOT {
		return p.cmp()
	}
	op := p.tok
	p.next()
	x, ok := p.not()
	if !ok {
		return nil, false
	}
	return &ast.Unary{Op: op.Kind, X: x, P: op.Pos}, true
}

func (p *parser) cmp() (ast.Expr, bool) {
	return p.binary(p.add, true, token.EQ, token.NEQ, token.LT, token.LTE, token.GT, token.GTE)
}

func (p *parser) add() (ast.Expr, bool) {
	return p.binary(p.mul, false, token.PLUS, token.MINUS)
}

func (p *parser) mul() (ast.Expr, bool) {
	return p.binary(p.unary, false, token.STAR, token.SLASH, token.PERCENT)
}

func (p *parser) unary() (ast.Expr, bool) {
	if p.tok.Kind != token.MINUS {
		return p.primary()
	}
	op := p.tok
	p.next()
	x, ok := p.unary()
	if !ok {
		return nil, false
	}
	return &ast.Unary{Op: op.Kind, X: x, P: op.Pos}, true
}

func (p *parser) primary() (ast.Expr, bool) {
	t := p.tok
	switch t.Kind {
	case token.STRING, token.INT, token.TRUE, token.FALSE:
		p.next()
		return &ast.Literal{Kind: t.Kind, Value: t.Lit, P: t.Pos}, true
	case token.IDENT:
		if p.peek().Kind == token.DOT {
			return p.call()
		}
		p.next()
		return &ast.Ident{Name: t.Lit, P: t.Pos}, true
	case token.LPAREN:
		p.next()
		x, ok := p.expr()
		if !ok {
			return nil, false
		}
		if _, ok = p.expect(token.RPAREN, "')'"); !ok {
			return nil, false
		}
		return x, true
	default:
		p.errorf(t.Pos, "expected expression, got %s", describe(t))
		return nil, false
	}
}

func (p *parser) binary(operand func() (ast.Expr, bool), once bool, ops ...token.Kind) (ast.Expr, bool) {
	x, ok := operand()
	if !ok {
		return nil, false
	}
	for isOneOf(p.tok.Kind, ops) {
		op := p.tok
		p.next()
		y, ok := operand()
		if !ok {
			return nil, false
		}
		x = &ast.Binary{Op: op.Kind, X: x, Y: y, P: op.Pos}
		if once {
			break
		}
	}
	return x, true
}

func isOneOf(k token.Kind, ops []token.Kind) bool {
	for _, o := range ops {
		if k == o {
			return true
		}
	}
	return false
}

func describe(t token.Token) string {
	if t.Kind == token.ILLEGAL {
		return fmt.Sprintf("illegal %q", t.Lit)
	}
	return t.Kind.String()
}

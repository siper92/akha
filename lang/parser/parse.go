package parser

import (
	"fmt"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/token"
)

type parser struct {
	lx   lexer.Lexer
	tok  token.Token
	errs Errors
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
		c, ok := p.call()
		if !ok {
			p.sync()
			continue
		}
		s.Calls = append(s.Calls, c)
		switch p.tok.Kind {
		case token.NEWLINE:
			p.next()
		case token.EOF:
		default:
			p.errorf(p.tok.Pos, "expected newline after call, got %s", describe(p.tok))
			p.sync()
		}
	}
	if len(p.errs) == 0 {
		return s, nil
	}
	return s, p.errs
}

func (p *parser) next() {
	for {
		p.tok = p.lx.Next()
		if p.tok.Kind != token.COMMENT {
			return
		}
	}
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
	switch p.tok.Kind {
	case token.IDENT:
		id := p.tok
		p.next()
		switch p.tok.Kind {
		case token.ASSIGN:
			p.next()
			p.skipNewlines()
			v, ok := p.literal()
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
			if c.Target.Module != "Ak" || c.Target.Name != "Allow" {
				p.errorf(id.Pos, "spread only allowed in Ak.Allow")
			}
			p.positional(c, &ast.Spread{Module: id.Lit, P: id.Pos})
			return true
		default:
			p.positional(c, p.identLiteral(id))
			return true
		}
	case token.STRING, token.INT:
		t := p.tok
		p.next()
		p.positional(c, &ast.Literal{Kind: t.Kind, Value: t.Lit, P: t.Pos})
		return true
	default:
		p.errorf(p.tok.Pos, "expected argument, got %s", describe(p.tok))
		return false
	}
}

func (p *parser) positional(c *ast.Call, e ast.Expr) {
	if len(c.Kwargs) > 0 {
		p.errorf(e.Pos(), "positional argument after keyword argument")
	}
	c.Args = append(c.Args, e)
}

func (p *parser) literal() (ast.Expr, bool) {
	t := p.tok
	switch t.Kind {
	case token.STRING, token.INT:
		p.next()
		return &ast.Literal{Kind: t.Kind, Value: t.Lit, P: t.Pos}, true
	case token.IDENT:
		p.next()
		return p.identLiteral(t), true
	default:
		p.errorf(t.Pos, "expected value, got %s", describe(t))
		return nil, false
	}
}

func (p *parser) identLiteral(t token.Token) ast.Expr {
	if t.Lit != "true" && t.Lit != "false" {
		p.errorf(t.Pos, "unexpected identifier")
	}
	return &ast.Literal{Kind: token.IDENT, Value: t.Lit, P: t.Pos}
}

func describe(t token.Token) string {
	if t.Kind == token.ILLEGAL {
		return fmt.Sprintf("illegal %q", t.Lit)
	}
	return t.Kind.String()
}

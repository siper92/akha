package parser

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/diag"
	"github.com/siper92/akha/lang/lexer"
	"github.com/siper92/akha/lang/token"
)

type Parser interface {
	Parse() (*ast.Script, error)
	ParseExpr() (ast.Expr, error)
}

var _ Parser = (*parser)(nil)

type parser struct {
	file  string
	src   string
	toks  []token.Token
	i     int
	tok   token.Token
	nest  int
	loops int
}

type bailout struct {
	err *diag.Error
}

var comparisons = []token.Kind{token.Eq, token.NotEq, token.Lt, token.LtEq, token.Gt, token.GtEq}

func New(file, src string) Parser {
	return &parser{file: file, src: src}
}

func (p *parser) Parse() (_ *ast.Script, err error) {
	defer p.recover(&err)
	p.load(lexer.New(p.src))
	return p.parseScript(), nil
}

func (p *parser) ParseExpr() (_ ast.Expr, err error) {
	defer p.recover(&err)
	p.load(lexer.New(p.src))
	x := p.parseExpr()
	p.skipNewlines()
	if p.tok.Kind != token.EOF {
		p.fail(diag.CodeUnexpectedToken, "", "unexpected %s after expression", describe(p.tok))
	}
	return x, nil
}

func (p *parser) recover(err *error) {
	r := recover()
	if r == nil {
		return
	}
	b, ok := r.(bailout)
	if !ok {
		panic(r)
	}
	b.err.File = p.file
	*err = b.err
}

func (p *parser) load(l lexer.Lexer) {
	p.toks, p.i, p.nest, p.loops = nil, -1, 0, 0
	for {
		t, err := l.Next()
		if err != nil {
			var de *diag.Error
			if !errors.As(err, &de) {
				de = diag.New(diag.ErrLex, token.Pos{}, diag.CodeUnexpectedChar, err.Error(), "")
			}
			panic(bailout{err: de})
		}
		p.toks = append(p.toks, t)
		if t.Kind == token.EOF {
			break
		}
	}
	p.next()
}

func (p *parser) next() {
	for p.i < len(p.toks)-1 {
		p.i++
		p.tok = p.toks[p.i]
		if p.nest == 0 || p.tok.Kind != token.Newline {
			return
		}
	}
}

func (p *parser) peek() token.Token {
	for j := p.i + 1; j < len(p.toks); j++ {
		if p.nest > 0 && p.toks[j].Kind == token.Newline {
			continue
		}
		return p.toks[j]
	}
	return p.toks[len(p.toks)-1]
}

func (p *parser) open() {
	p.nest++
	p.next()
}

func (p *parser) close(kind token.Kind) {
	if p.tok.Kind != kind {
		p.fail(diag.CodeUnexpectedToken, "", "expected %s, got %s", kind, describe(p.tok))
	}
	p.nest--
	p.next()
}

func (p *parser) skipNewlines() {
	for p.tok.Kind == token.Newline {
		p.next()
	}
}

func (p *parser) fail(code, hint, format string, args ...any) {
	p.failAt(p.tok.Pos, code, hint, format, args...)
}

func (p *parser) failAt(pos token.Pos, code, hint, format string, args ...any) {
	panic(bailout{err: diag.New(diag.ErrParse, pos, code, fmt.Sprintf(format, args...), hint)})
}

func (p *parser) parseScript() *ast.Script {
	script := &ast.Script{}
	for {
		p.skipNewlines()
		if p.tok.Kind == token.EOF {
			return script
		}
		script.Stmts = append(script.Stmts, p.parseStmt())
		p.endStmt()
	}
}

func (p *parser) endStmt() {
	switch p.tok.Kind {
	case token.Newline:
		p.next()
	case token.EOF:
	default:
		p.fail(diag.CodeExpectedNewline, "one statement per line", "expected end of line, got %s", describe(p.tok))
	}
}

func (p *parser) parseBlock() *ast.Block {
	switch p.tok.Kind {
	case token.LBrace:
	case token.Newline:
		p.fail(diag.CodeBlockOpen, "put { at the end of the header line", "expected { on the header line")
	default:
		p.fail(diag.CodeUnexpectedToken, "", "expected {, got %s", describe(p.tok))
	}
	open := p.tok
	block := &ast.Block{Pos: line(open)}
	p.next()
	if p.tok.Kind != token.Newline && p.tok.Kind != token.EOF {
		p.fail(diag.CodeBlockNewline, "} must be on its own line", "expected a new line after {, got %s", describe(p.tok))
	}
	for {
		p.skipNewlines()
		switch p.tok.Kind {
		case token.RBrace:
			p.next()
			return block
		case token.EOF:
			p.fail(diag.CodeUnclosedBlock, "", "missing } for the block opened on line %d", open.Pos.Line)
		}
		block.Stmts = append(block.Stmts, p.parseStmt())
		p.endStmt()
	}
}

func (p *parser) parseStmt() ast.Stmt {
	switch p.tok.Kind {
	case token.Let:
		return p.parseLet()
	case token.Var:
		return p.parseVar()
	case token.If:
		return p.parseIf()
	case token.For:
		return p.parseFor()
	case token.Break, token.Continue:
		return p.parseLoopControl()
	case token.Return, token.Exit:
		return p.parseReturn()
	case token.Else:
		p.fail(diag.CodeElsePlacement, "write } else {", "else must be on the same line as the closing }")
	case token.RBrace:
		p.fail(diag.CodeUnexpectedToken, "", "unexpected }")
	case token.Reserved:
		p.fail(diag.CodeReserved, "", "%q is reserved for a later version", p.tok.Lit)
	}
	return p.parseSimple()
}

func (p *parser) parseName() string {
	t := p.tok
	switch {
	case t.Kind == token.Ident:
		p.next()
		return t.Lit
	case t.Kind.IsKeyword():
		p.fail(diag.CodeKeywordName, "", "%q is a keyword and cannot be used as a name", t.Lit)
	case t.Kind == token.Reserved:
		p.fail(diag.CodeReserved, "", "%q is reserved for a later version", t.Lit)
	}
	p.fail(diag.CodeExpectedName, "", "expected a name, got %s", describe(t))
	return ""
}

func (p *parser) parseLet() ast.Stmt {
	pos := line(p.tok)
	p.next()
	name := p.parseName()
	if p.tok.Kind != token.Assign {
		p.fail(diag.CodeLetInit, "use var for a binding without a value", "let %s requires an initializer", name)
	}
	p.next()
	return &ast.LetStmt{Pos: pos, Name: name, Value: p.parseExpr()}
}

func (p *parser) parseVar() ast.Stmt {
	pos := line(p.tok)
	p.next()
	at := p.tok
	name := p.parseName()
	if name == "_" {
		p.failAt(at.Pos, diag.CodeDiscardVar, "use let _ = expr", "_ can only be used with let")
	}
	stmt := &ast.VarStmt{Pos: pos, Name: name}
	if p.tok.Kind == token.Assign {
		p.next()
		stmt.Value = p.parseExpr()
	}
	return stmt
}

func (p *parser) parseIf() *ast.IfStmt {
	stmt := &ast.IfStmt{Pos: line(p.tok)}
	p.next()
	stmt.Cond = p.parseHeader()
	stmt.Then = p.parseBlock()
	if p.tok.Kind != token.Else {
		return stmt
	}
	p.next()
	if p.tok.Kind == token.If {
		stmt.Else = p.parseIf()
		return stmt
	}
	stmt.Else = p.parseBlock()
	return stmt
}

func (p *parser) parseHeader() ast.Expr {
	if p.tok.Kind == token.LBrace {
		if p.peek().Kind == token.Newline {
			p.fail(diag.CodeExpectedExpr, "", "expected a condition before {")
		}
		p.fail(diag.CodeHeaderObject, "wrap it in ( )", "an object literal cannot start a header")
	}
	return p.parseExpr()
}

func (p *parser) parseFor() ast.Stmt {
	pos := line(p.tok)
	p.next()
	if p.tok.Kind == token.Let {
		p.fail(diag.CodeForLet, "use for var to make loop variables mutable", "for let is redundant, loop variables are immutable")
	}
	mutable := p.tok.Kind == token.Var
	if mutable {
		p.next()
	}
	first := p.parseName()
	second := ""
	if p.tok.Kind == token.Comma {
		p.next()
		second = p.parseName()
	}
	switch p.tok.Kind {
	case token.In:
		p.next()
		stmt := &ast.ForInStmt{Pos: pos, Mutable: mutable, Value: first, Iter: p.parseHeader()}
		if second != "" {
			stmt.Key, stmt.Value = first, second
		}
		stmt.Body = p.parseLoopBody()
		return stmt
	case token.Range:
		if second != "" {
			p.fail(diag.CodeRangeVars, "", "range takes one loop variable")
		}
		p.next()
		stmt := &ast.ForRangeStmt{Pos: pos, Mutable: mutable, Name: first}
		stmt.Start, stmt.End = p.parseRange()
		stmt.Body = p.parseLoopBody()
		return stmt
	}
	p.fail(diag.CodeUnexpectedToken, "", "expected in or range, got %s", describe(p.tok))
	return nil
}

func (p *parser) parseRange() (ast.Expr, ast.Expr) {
	if p.tok.Kind != token.LBracket {
		p.fail(diag.CodeRangeSyntax, "use range [start..end]", "expected [ after range, got %s", describe(p.tok))
	}
	p.open()
	start := p.parseExpr()
	if p.tok.Kind != token.DotDot {
		p.fail(diag.CodeRangeSyntax, "use range [start..end]", "expected .. in range, got %s", describe(p.tok))
	}
	p.next()
	end := p.parseExpr()
	p.close(token.RBracket)
	return start, end
}

func (p *parser) parseLoopBody() *ast.Block {
	p.loops++
	body := p.parseBlock()
	p.loops--
	return body
}

func (p *parser) parseLoopControl() ast.Stmt {
	t := p.tok
	if p.loops == 0 {
		p.fail(diag.CodeLoopControl, "", "%s outside of a loop", t.Lit)
	}
	p.next()
	if t.Kind == token.Break {
		return &ast.BreakStmt{Pos: line(t)}
	}
	return &ast.ContinueStmt{Pos: line(t)}
}

func (p *parser) parseReturn() ast.Stmt {
	t := p.tok
	p.next()
	stmt := &ast.ReturnStmt{Pos: line(t), Exit: t.Kind == token.Exit}
	if p.tok.Kind != token.Newline && p.tok.Kind != token.EOF {
		stmt.Value = p.parseExpr()
	}
	return stmt
}

func (p *parser) parseSimple() ast.Stmt {
	start := p.tok
	x := p.parseExpr()
	if p.tok.Kind == token.Assign {
		if !assignable(x) {
			p.failAt(start.Pos, diag.CodeAssignTarget, "", "cannot assign to this expression")
		}
		p.next()
		return &ast.AssignStmt{Pos: line(start), Target: x, Value: p.parseExpr()}
	}
	if _, ok := x.(*ast.CallExpr); !ok {
		p.failAt(start.Pos, diag.CodeUnusedValue, "use let _ = expr to discard a value", "unused value, only a call can stand alone")
	}
	return &ast.ExprStmt{Pos: line(start), X: x}
}

func (p *parser) parseExpr() ast.Expr {
	return p.parseOr()
}

func (p *parser) parseBinary(operand func() ast.Expr, ops ...token.Kind) ast.Expr {
	left := operand()
	for slices.Contains(ops, p.tok.Kind) {
		op := p.tok
		p.next()
		left = &ast.BinaryExpr{Pos: line(op), Op: op.Kind, Left: left, Right: operand()}
	}
	return left
}

func (p *parser) parseOr() ast.Expr {
	return p.parseBinary(p.parseAnd, token.Or)
}

func (p *parser) parseAnd() ast.Expr {
	return p.parseBinary(p.parseNot, token.And)
}

func (p *parser) parseNot() ast.Expr {
	if p.tok.Kind != token.Not {
		return p.parseMembership()
	}
	op := p.tok
	p.next()
	return &ast.UnaryExpr{Pos: line(op), Op: token.Not, X: p.parseNot()}
}

func (p *parser) membershipOp() (token.Kind, bool) {
	switch {
	case p.tok.Kind == token.In:
		return token.In, true
	case p.tok.Kind == token.Not && p.tok.Lit == "not" && p.peek().Kind == token.In:
		return token.NotIn, true
	}
	return 0, false
}

func (p *parser) parseMembership() ast.Expr {
	left := p.parseComparison()
	op, ok := p.membershipOp()
	if !ok {
		return left
	}
	at := p.tok
	if op == token.NotIn {
		p.next()
	}
	p.next()
	x := &ast.BinaryExpr{Pos: line(at), Op: op, Left: left, Right: p.parseComparison()}
	if _, again := p.membershipOp(); again {
		p.fail(diag.CodeChainedIn, "use ( ) to group", "in does not chain")
	}
	return x
}

func (p *parser) parseComparison() ast.Expr {
	left := p.parseAdd()
	if !slices.Contains(comparisons, p.tok.Kind) {
		return left
	}
	op := p.tok
	p.next()
	x := &ast.BinaryExpr{Pos: line(op), Op: op.Kind, Left: left, Right: p.parseAdd()}
	if slices.Contains(comparisons, p.tok.Kind) {
		p.fail(diag.CodeChainedComparison, "use a < b and b < c", "comparisons do not chain")
	}
	return x
}

func (p *parser) parseAdd() ast.Expr {
	return p.parseBinary(p.parseMul, token.Plus, token.Minus)
}

func (p *parser) parseMul() ast.Expr {
	return p.parseBinary(p.parseUnary, token.Star, token.Slash, token.Percent)
}

func (p *parser) parseUnary() ast.Expr {
	if p.tok.Kind != token.Minus {
		return p.parsePostfix()
	}
	op := p.tok
	p.next()
	return &ast.UnaryExpr{Pos: line(op), Op: token.Minus, X: p.parseUnary()}
}

func (p *parser) parsePostfix() ast.Expr {
	x := p.parsePrimary()
	for {
		at := p.tok
		switch at.Kind {
		case token.Dot:
			p.next()
			if p.tok.Kind != token.Ident {
				p.fail(diag.CodeExpectedName, "", "expected a member name after ., got %s", describe(p.tok))
			}
			x = &ast.MemberExpr{Pos: line(at), X: x, Name: p.tok.Lit}
			p.next()
		case token.LBracket:
			p.open()
			index := p.parseExpr()
			if p.tok.Kind == token.Colon {
				p.fail(diag.CodeSlice, "", "slices are not supported in v1")
			}
			p.close(token.RBracket)
			x = &ast.IndexExpr{Pos: line(at), X: x, Index: index}
		case token.LParen:
			p.open()
			x = &ast.CallExpr{Pos: line(at), Fn: x, Args: p.parseList(token.RParen)}
		default:
			return x
		}
	}
}

func (p *parser) parseList(end token.Kind) []ast.Expr {
	var items []ast.Expr
	for p.tok.Kind != end {
		items = append(items, p.parseExpr())
		if p.tok.Kind != token.Comma {
			break
		}
		p.next()
	}
	p.close(end)
	return items
}

func (p *parser) parsePrimary() ast.Expr {
	t := p.tok
	pos := line(t)
	switch t.Kind {
	case token.Number:
		v, err := strconv.ParseFloat(t.Lit, 64)
		if err != nil {
			p.fail(diag.CodeInvalidNumber, "", "invalid number %s", t.Lit)
		}
		p.next()
		return &ast.NumberLit{Pos: pos, Value: v, Raw: t.Lit}
	case token.String:
		p.next()
		return &ast.StringLit{Pos: pos, Value: t.Lit}
	case token.Template:
		p.next()
		return p.parseTemplate(t)
	case token.True, token.False:
		p.next()
		return &ast.BoolLit{Pos: pos, Value: t.Kind == token.True}
	case token.Null:
		p.next()
		return &ast.NullLit{Pos: pos}
	case token.Ident:
		p.next()
		return &ast.Ident{Pos: pos, Name: t.Lit}
	case token.LParen:
		p.open()
		x := p.parseExpr()
		p.close(token.RParen)
		return x
	case token.LBracket:
		p.open()
		return &ast.ArrayLit{Pos: pos, Elems: p.parseList(token.RBracket)}
	case token.LBrace:
		return p.parseObject()
	}
	p.fail(diag.CodeExpectedExpr, "", "expected an expression, got %s", describe(t))
	return nil
}

func (p *parser) parseObject() ast.Expr {
	obj := &ast.ObjectLit{Pos: line(p.tok)}
	p.open()
	seen := make(map[string]bool)
	for p.tok.Kind != token.RBrace {
		key := p.tok
		switch key.Kind {
		case token.Ident, token.String:
		case token.Template:
			p.fail(diag.CodeObjectKey, "", "computed keys are not supported in v1")
		default:
			p.fail(diag.CodeObjectKey, "use a name or a string as key", "invalid object key %s", describe(key))
		}
		if seen[key.Lit] {
			p.fail(diag.CodeDuplicateKey, "", "duplicate key %q", key.Lit)
		}
		seen[key.Lit] = true
		p.next()
		if p.tok.Kind != token.Colon {
			p.fail(diag.CodeUnexpectedToken, "", "expected :, got %s", describe(p.tok))
		}
		p.next()
		obj.Entries = append(obj.Entries, ast.Entry{Pos: line(key), Key: key.Lit, Value: p.parseExpr()})
		if p.tok.Kind != token.Comma {
			break
		}
		p.next()
	}
	p.close(token.RBrace)
	return obj
}

func (p *parser) parseTemplate(t token.Token) ast.Expr {
	lit := &ast.TemplateLit{Pos: line(t)}
	for _, part := range t.Parts {
		if !part.IsExpr {
			lit.Parts = append(lit.Parts, ast.TemplatePart{Text: part.Text})
			continue
		}
		sub := &parser{file: p.file}
		sub.load(lexer.NewAt(part.Expr, part.Pos))
		x := sub.parseExpr()
		if sub.tok.Kind != token.EOF {
			sub.fail(diag.CodeInterpExpr, "", "unexpected %s inside ${ }", describe(sub.tok))
		}
		if !valueRef(x) {
			p.failAt(part.Pos, diag.CodeInterpExpr, "move the expression to a let", "only names, members and indexes are allowed inside ${ }")
		}
		lit.Parts = append(lit.Parts, ast.TemplatePart{Expr: x})
	}
	return lit
}

func assignable(x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.Ident:
		return x.Name != "_"
	case *ast.MemberExpr:
		return assignable(x.X)
	case *ast.IndexExpr:
		return assignable(x.X)
	}
	return false
}

func valueRef(x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.Ident:
		return true
	case *ast.MemberExpr:
		return valueRef(x.X)
	case *ast.IndexExpr:
		switch x.Index.(type) {
		case *ast.NumberLit, *ast.StringLit:
			return valueRef(x.X)
		}
		return valueRef(x.X) && valueRef(x.Index)
	}
	return false
}

func line(t token.Token) ast.Pos {
	return ast.Pos{Line: t.Pos.Line}
}

func describe(t token.Token) string {
	switch t.Kind {
	case token.EOF:
		return "end of file"
	case token.Newline:
		return "end of line"
	case token.Ident, token.Number, token.Reserved:
		return t.Kind.String() + " " + t.Lit
	case token.String, token.Template:
		return "string " + strconv.Quote(t.Lit)
	}
	return strconv.Quote(t.Lit)
}

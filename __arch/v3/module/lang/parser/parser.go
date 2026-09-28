package parser

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
)

type Parser interface {
	Parse() (*ast.Script, error)
}

var _ Parser = (*parser)(nil)

const MaxDepth = 1000

type parser struct {
	file  string
	src   string
	toks  []lexer.Token
	i     int
	tok   lexer.Token
	nest  int
	loops int
	depth int
}

type bailout struct {
	err *lexer.Error
}

const (
	precOr = iota
	precAnd
	precNot
	precIn
	precCmp
	precAdd
	precMul
	precUnary
)

var levels = [precUnary][]lexer.Kind{
	precOr:  {lexer.Or},
	precAnd: {lexer.And},
	precIn:  {lexer.In, lexer.NotIn},
	precCmp: {lexer.Eq, lexer.NotEq, lexer.Lt, lexer.LtEq, lexer.Gt, lexer.GtEq},
	precAdd: {lexer.Plus, lexer.Minus},
	precMul: {lexer.Star, lexer.Slash, lexer.Percent},
}

var precedence = func() map[lexer.Kind]int {
	m := make(map[lexer.Kind]int)
	for level, ops := range levels {
		for _, op := range ops {
			m[op] = level
		}
	}

	return m
}()

type chainErr struct {
	code, hint, msg string
}

var nonAssoc = map[int]chainErr{
	precIn:  {code: lexer.CodeChainedIn, hint: "use ( ) to group", msg: "in does not chain"},
	precCmp: {code: lexer.CodeChainedComparison, hint: "use a < b and b < c", msg: "comparisons do not chain"},
}

func New(file, src string) Parser {
	return &parser{file: file, src: src}
}

func (p *parser) Parse() (_ *ast.Script, err error) {
	defer p.recover(&err)

	if p.file != "" {
		content, err := os.ReadFile(p.file)
		if err != nil {
			return nil, err
		}

		p.src = string(content)
	}

	p.load(lexer.New(p.src))

	return p.parseScript(), nil
}

func (p *parser) recover(err *error) {
	switch r := recover().(type) {
	case nil:
	case bailout:
		r.err.File = p.file
		*err = r.err
	default:
		de := lexer.NewLexError(lexer.ErrParse, p.tok.Pos, lexer.CodeInternal, fmt.Sprintf("internal error: %v", r), "")
		de.File = p.file
		*err = de
	}
}

func (p *parser) enter() {
	p.depth++
	if p.depth > MaxDepth {
		p.fail(lexer.CodeNesting, "", "excessive nesting")
	}
}

func (p *parser) leave() {
	p.depth--
}

func (p *parser) load(l lexer.Lexer) {
	p.toks, p.i, p.nest, p.loops, p.depth = nil, -1, 0, 0, 0
	for {
		t, err := l.Next()
		if err != nil {
			var de *lexer.Error
			if !errors.As(err, &de) {
				de = lexer.NewLexError(lexer.ErrLex, lexer.Pos{}, lexer.CodeUnexpectedChar, err.Error(), "")
			}
			panic(bailout{err: de})
		}

		p.toks = append(p.toks, t)
		if t.Kind == lexer.EOF {
			break
		}
	}

	p.next()
}

func (p *parser) next() {
	for p.i < len(p.toks)-1 {
		p.i++
		p.tok = p.toks[p.i]
		if p.nest == 0 || p.tok.Kind != lexer.Newline {
			return
		}
	}
}

func (p *parser) peek() lexer.Token {
	for j := p.i + 1; j < len(p.toks); j++ {
		if p.nest > 0 && p.toks[j].Kind == lexer.Newline {
			continue
		}
		return p.toks[j]
	}
	return p.toks[len(p.toks)-1]
}

func (p *parser) consume(kind lexer.Kind) lexer.Token {
	t := p.tok
	if t.Kind != kind {
		p.fail(lexer.CodeUnexpectedToken, "", "expected %s, got %s", kind, t.Describe())
	}
	p.next()
	return t
}

func (p *parser) open() {
	p.nest++
	p.next()
}

func (p *parser) close(kind lexer.Kind) {
	if p.tok.Kind != kind {
		p.fail(lexer.CodeUnexpectedToken, "", "expected %s, got %s", kind, p.tok.Describe())
	}
	p.nest--
	p.next()
}

func (p *parser) skipNewlines() {
	for p.tok.Kind == lexer.Newline {
		p.next()
	}
}

func (p *parser) fail(code, hint, format string, args ...any) {
	p.failAt(p.tok.Pos, code, hint, format, args...)
}

func (p *parser) failAt(pos lexer.Pos, code, hint, format string, args ...any) {
	panic(bailout{err: lexer.NewLexError(lexer.ErrParse, pos, code, fmt.Sprintf(format, args...), hint)})
}

// Script = {Line} eof .
func (p *parser) parseScript() *ast.Script {
	script := &ast.Script{}
	for {
		p.skipNewlines()
		if p.tok.Kind == lexer.EOF {
			return script
		}

		script.Stmts = append(script.Stmts, p.parseStmt())
		p.endStmt()
	}
}

// Line = [Statement] newline .
func (p *parser) endStmt() {
	switch p.tok.Kind {
	case lexer.Newline:
		p.next()
	case lexer.EOF:
	default:
		p.fail(lexer.CodeExpectedNewline, "one statement per line", "expected end of line, got %s", p.tok.Describe())
	}
}

// Block = '{' newline {Line} '}' .
func (p *parser) parseBlock() *ast.Block {
	p.enter()
	defer p.leave()

	switch p.tok.Kind {
	case lexer.LBrace:
	case lexer.Newline:
		p.fail(lexer.CodeBlockOpen, "put { at the end of the header line", "expected { on the header line")
	default:
		p.fail(lexer.CodeUnexpectedToken, "", "expected {, got %s", p.tok.Describe())
	}

	open := p.tok
	block := &ast.Block{Pos: line(open)}
	p.next()
	if p.tok.Kind != lexer.Newline && p.tok.Kind != lexer.EOF {
		p.fail(lexer.CodeBlockNewline, "} must be on its own line", "expected a new line after {, got %s", p.tok.Describe())
	}

	for {
		p.skipNewlines()
		switch p.tok.Kind {
		case lexer.RBrace:
			p.next()
			return block
		case lexer.EOF:
			p.fail(lexer.CodeUnclosedBlock, "", "missing } for the block opened on line %d", open.Pos.Line)
		}
		block.Stmts = append(block.Stmts, p.parseStmt())
		p.endStmt()
	}
}

// Statement = LetStmt | VarStmt | IfStmt | ForStmt | BreakStmt | ContinueStmt | ReturnStmt | AssignStmt | CallStmt .
func (p *parser) parseStmt() ast.Stmt {
	switch p.tok.Kind {
	case lexer.Let:
		return p.parseLet()
	case lexer.Var:
		return p.parseVar()
	case lexer.If:
		return p.parseIf()
	case lexer.For:
		return p.parseFor()
	case lexer.Break, lexer.Continue:
		return p.parseLoopControl()
	case lexer.Return, lexer.Exit:
		return p.parseReturn()
	case lexer.Else:
		p.fail(lexer.CodeElsePlacement, "write } else {", "else must be on the same line as the closing }")
	case lexer.RBrace:
		p.fail(lexer.CodeUnexpectedToken, "", "unexpected }")
	case lexer.Reserved:
		p.fail(lexer.CodeReserved, "", "%q is reserved for a later version", p.tok.Lit)
	}
	return p.parseSimple()
}

func (p *parser) parseName() string {
	t := p.tok
	switch {
	case t.Kind == lexer.Ident:
		p.next()
		return t.Lit
	case t.Kind.IsKeyword():
		p.fail(lexer.CodeKeywordName, "", "%q is a keyword and cannot be used as a name", t.Lit)
	case t.Kind == lexer.Reserved:
		p.fail(lexer.CodeReserved, "", "%q is reserved for a later version", t.Lit)
	}
	p.fail(lexer.CodeExpectedName, "", "expected a name, got %s", t.Describe())
	return ""
}

// LetStmt = 'let' identifier '=' Expr .
func (p *parser) parseLet() ast.Stmt {
	pos := line(p.tok)
	p.next()
	name := p.parseName()
	if p.tok.Kind != lexer.Assign {
		p.fail(lexer.CodeLetInit, "use var for a binding without a value", "let %s requires an initializer", name)
	}
	p.next()
	return &ast.LetStmt{Pos: pos, Name: name, Value: p.parseExpr()}
}

// VarStmt = 'var' identifier ['=' Expr] .
func (p *parser) parseVar() ast.Stmt {
	pos := line(p.tok)
	p.next()
	at := p.tok
	name := p.parseName()
	if name == "_" {
		p.failAt(at.Pos, lexer.CodeDiscardVar, "use let _ = expr", "_ can only be used with let")
	}
	stmt := &ast.VarStmt{Pos: pos, Name: name}
	if p.tok.Kind == lexer.Assign {
		p.next()
		stmt.Value = p.parseExpr()
	}
	return stmt
}

// IfStmt = 'if' Header Block ['else' (IfStmt | Block)] .
func (p *parser) parseIf() *ast.IfStmt {
	p.enter()
	defer p.leave()

	stmt := &ast.IfStmt{Pos: line(p.tok)}
	p.next()
	stmt.Cond = p.parseHeader()
	stmt.Then = p.parseBlock()
	if p.tok.Kind != lexer.Else {
		return stmt
	}
	p.next()
	if p.tok.Kind == lexer.If {
		stmt.Else = p.parseIf()
		return stmt
	}
	stmt.Else = p.parseBlock()
	return stmt
}

// Header = Expr .
func (p *parser) parseHeader() ast.Expr {
	if p.tok.Kind == lexer.LBrace {
		if p.peek().Kind == lexer.Newline {
			p.fail(lexer.CodeExpectedExpr, "", "expected a condition before {")
		}
		p.fail(lexer.CodeHeaderObject, "wrap it in ( )", "an object literal cannot start a header")
	}
	return p.parseExpr()
}

// ForStmt = 'for' ['var'] LoopVariables ('in' Header | 'range' Range) Block .
func (p *parser) parseFor() ast.Stmt {
	pos := line(p.tok)
	p.next()
	if p.tok.Kind == lexer.Let {
		p.fail(lexer.CodeForLet, "use for var to make loop variables mutable", "for let is redundant, loop variables are immutable")
	}
	mutable := p.tok.Kind == lexer.Var
	if mutable {
		p.next()
	}

	first, second := p.parseLoopVars()
	switch p.tok.Kind {
	case lexer.In:
		p.next()
		stmt := &ast.ForInStmt{Pos: pos, Mutable: mutable, Value: first, Iter: p.parseHeader()}
		if second != "" {
			stmt.Key, stmt.Value = first, second
		}
		stmt.Body = p.parseLoopBody()
		return stmt
	case lexer.Range:
		if second != "" {
			p.fail(lexer.CodeRangeVars, "", "range takes one loop variable")
		}
		p.next()
		stmt := &ast.ForRangeStmt{Pos: pos, Mutable: mutable, Name: first}
		stmt.Start, stmt.End = p.parseRange()
		stmt.Body = p.parseLoopBody()
		return stmt
	default:
		p.fail(lexer.CodeUnexpectedToken, "", "expected in or range, got %s", p.tok.Describe())
		return nil
	}
}

// LoopVariables = identifier [',' identifier] .
func (p *parser) parseLoopVars() (string, string) {
	first := p.parseName()
	if p.tok.Kind != lexer.Comma {
		return first, ""
	}
	p.next()

	return first, p.parseName()
}

// Range = '[' Expr '..' Expr ']' .
func (p *parser) parseRange() (ast.Expr, ast.Expr) {
	if p.tok.Kind != lexer.LBracket {
		p.fail(lexer.CodeRangeSyntax, "use range [start..end]", "expected [ after range, got %s", p.tok.Describe())
	}
	p.open()

	start := p.parseExpr()
	if p.tok.Kind != lexer.DotDot {
		p.fail(lexer.CodeRangeSyntax, "use range [start..end]", "expected .. in range, got %s", p.tok.Describe())
	}
	p.next()

	end := p.parseExpr()
	p.close(lexer.RBracket)

	return start, end
}

func (p *parser) parseLoopBody() *ast.Block {
	p.loops++
	body := p.parseBlock()
	p.loops--
	return body
}

// BreakStmt = 'break' .
// ContinueStmt = 'continue' .
func (p *parser) parseLoopControl() ast.Stmt {
	t := p.tok
	if p.loops == 0 {
		p.fail(lexer.CodeLoopControl, "", "%s outside of a loop", t.Kind)
	}

	p.next()
	if t.Kind == lexer.Break {
		return &ast.BreakStmt{Pos: line(t)}
	}

	return &ast.ContinueStmt{Pos: line(t)}
}

// ReturnStmt = ('return' | 'exit') [Expr] .
func (p *parser) parseReturn() ast.Stmt {
	t := p.tok
	p.next()

	stmt := &ast.ReturnStmt{Pos: line(t), Exit: t.Kind == lexer.Exit}
	if p.tok.Kind != lexer.Newline && p.tok.Kind != lexer.EOF {
		stmt.Value = p.parseExpr()
	}

	return stmt
}

// AssignStmt = Target '=' Expr .
// CallStmt = PostfixExpr .
func (p *parser) parseSimple() ast.Stmt {
	start := p.tok
	x := p.parseExpr()
	if p.tok.Kind == lexer.Assign {
		if start.Kind != lexer.Ident || !assignable(x) {
			p.failAt(start.Pos, lexer.CodeAssignTarget, "", "cannot assign to this expression")
		}
		p.next()
		return &ast.AssignStmt{Pos: line(start), Target: x, Value: p.parseExpr()}
	}
	if _, ok := x.(*ast.CallExpr); !ok {
		p.failAt(start.Pos, lexer.CodeUnusedValue, "use let _ = expr to discard a value", "unused value, only a call can stand alone")
	}
	return &ast.ExprStmt{Pos: line(start), X: x}
}

// Expr = OrExpr .
func (p *parser) parseExpr() ast.Expr {
	return p.parsePrec(precOr)
}

// NotExpr = ('not' | '!') NotExpr | InExpr .
func (p *parser) parsePrec(prec int) ast.Expr {
	p.enter()
	defer p.leave()

	if prec >= precUnary {
		return p.parseUnary()
	}

	if prec == precNot && p.tok.Kind == lexer.Not {
		op := p.tok
		p.next()
		return &ast.UnaryExpr{Pos: line(op), Op: lexer.Not, X: p.parsePrec(precNot)}
	}

	return p.parseBinop(prec)
}

// OrExpr = AndExpr {('or' | '||') AndExpr} .
// AndExpr = NotExpr {('and' | '&&') NotExpr} .
// InExpr = CmpExpr [('in' | 'not' 'in') CmpExpr] .
// CmpExpr = AddExpr [('==' | '!=' | '<' | '<=' | '>' | '>=') AddExpr] .
// AddExpr = MulExpr {('+' | '-') MulExpr} .
// MulExpr = UnaryExpr {('*' | '/' | '%') UnaryExpr} .
func (p *parser) parseBinop(prec int) ast.Expr {
	x := p.parsePrec(prec + 1)
	for first := true; ; first = false {
		op, ok := p.binop()
		level := precedence[op]
		if !ok || level < prec {
			return x
		}
		if chain, ok := nonAssoc[level]; ok && !first {
			p.fail(chain.code, chain.hint, "%s", chain.msg)
		}
		at := p.tok
		if op == lexer.NotIn {
			p.next()
		}
		p.next()
		x = &ast.BinaryExpr{Pos: line(at), Op: op, Left: x, Right: p.parsePrec(level + 1)}
	}
}

func (p *parser) binop() (lexer.Kind, bool) {
	if p.tok.Kind == lexer.Not && p.tok.Lit != "!" && p.peek().Kind == lexer.In {
		return lexer.NotIn, true
	}
	_, ok := precedence[p.tok.Kind]
	return p.tok.Kind, ok
}

// UnaryExpr = '-' UnaryExpr | PostfixExpr .
func (p *parser) parseUnary() ast.Expr {
	p.enter()
	defer p.leave()

	if p.tok.Kind != lexer.Minus {
		return p.parsePostfix()
	}
	op := p.tok
	p.next()

	return &ast.UnaryExpr{Pos: line(op), Op: lexer.Minus, X: p.parseUnary()}
}

// PostfixExpr = Operand {DotSuffix | IndexSuffix | CallSuffix} .
// DotSuffix = '.' identifier .
// IndexSuffix = '[' Expr ']' .
// CallSuffix = '(' [List] ')' .
func (p *parser) parsePostfix() ast.Expr {
	x := p.parsePrimary()
	for {
		at := p.tok
		switch at.Kind {
		case lexer.Dot:
			if !objectOperand(x) {
				p.fail(lexer.CodeMemberKind, "use . on objects only", "member access works only on objects")
			}
			p.next()
			if p.tok.Kind != lexer.Ident {
				p.fail(lexer.CodeExpectedName, "", "expected a member name after ., got %s", p.tok.Describe())
			}
			x = &ast.MemberExpr{Pos: line(at), X: x, Name: p.tok.Lit}
			p.next()
		case lexer.LBracket:
			p.open()
			index := p.parseExpr()
			if p.tok.Kind == lexer.Colon {
				p.fail(lexer.CodeSlice, "", "slices are not supported in v1")
			}
			p.close(lexer.RBracket)
			x = &ast.IndexExpr{Pos: line(at), X: x, Index: index}
		case lexer.LParen:
			p.open()
			x = &ast.CallExpr{Pos: line(at), Fn: x, Args: p.parseList(lexer.RParen)}
		default:
			return x
		}
	}
}

// List = Expr {',' Expr} [','] .
func (p *parser) parseList(end lexer.Kind) []ast.Expr {
	var items []ast.Expr
	for p.tok.Kind != end {
		items = append(items, p.parseExpr())
		if p.tok.Kind != lexer.Comma {
			break
		}
		p.next()
	}
	p.close(end)
	return items
}

// Operand = identifier | number | String | 'true' | 'false' | 'null' | ArrayExpr | ObjectExpr | '(' Expr ')' .
// String = string | Template .
// ArrayExpr = '[' [List] ']' .
func (p *parser) parsePrimary() ast.Expr {
	t := p.tok
	pos := line(t)
	switch t.Kind {
	case lexer.Number:
		v, err := strconv.ParseFloat(t.Lit, 64)
		if err != nil {
			p.fail(lexer.CodeInvalidNumber, "", "invalid number %s", t.Lit)
		}
		p.next()
		return &ast.NumberLit{Pos: pos, Value: v, Raw: t.Lit}
	case lexer.String:
		p.next()
		return &ast.StringLit{Pos: pos, Value: t.Lit}
	case lexer.Template:
		p.next()
		return p.parseTemplate(t)
	case lexer.True, lexer.False:
		p.next()
		return &ast.BoolLit{Pos: pos, Value: t.Kind == lexer.True}
	case lexer.Null:
		p.next()
		return &ast.NullLit{Pos: pos}
	case lexer.Ident:
		p.next()
		return &ast.Ident{Pos: pos, Name: t.Lit}
	case lexer.LParen:
		p.open()
		x := p.parseExpr()
		p.close(lexer.RParen)
		return x
	case lexer.LBracket:
		p.open()
		return &ast.ArrayLit{Pos: pos, Elems: p.parseList(lexer.RBracket)}
	case lexer.LBrace:
		return p.parseObject()
	}

	p.fail(lexer.CodeExpectedExpr, "", "expected an expression, got %s", t.Describe())
	return nil
}

// ObjectExpr = '{' [Entry {',' Entry} [',']] '}' .
// Entry = (identifier | string) ':' Expr .
func (p *parser) parseObject() ast.Expr {
	obj := &ast.ObjectLit{Pos: line(p.tok)}
	p.open()
	seen := make(map[string]bool)
	for p.tok.Kind != lexer.RBrace {
		key := p.tok
		switch key.Kind {
		case lexer.Ident, lexer.String:
		case lexer.Template:
			p.fail(lexer.CodeObjectKey, "", "computed keys are not supported in v1")
		default:
			p.fail(lexer.CodeObjectKey, "use a name or a string as key", "invalid object key %s", key.Describe())
		}
		if seen[key.Lit] {
			p.fail(lexer.CodeDuplicateKey, "", "duplicate key %q", key.Lit)
		}
		seen[key.Lit] = true
		p.next()
		p.consume(lexer.Colon)
		obj.Entries = append(obj.Entries, ast.Entry{Pos: line(key), Key: key.Lit, Value: p.parseExpr()})
		if p.tok.Kind != lexer.Comma {
			break
		}
		p.next()
	}
	p.close(lexer.RBrace)
	return obj
}

// Template = '"' {char | escape | '${' ValueRef '}'} '"' .
func (p *parser) parseTemplate(t lexer.Token) ast.Expr {
	lit := &ast.TemplateLit{Pos: line(t)}
	for _, part := range t.Parts {
		if !part.IsExpr {
			lit.Parts = append(lit.Parts, ast.TemplatePart{Text: part.Text})
			continue
		}
		sub := &parser{file: p.file}
		sub.load(lexer.NewAt(part.Expr, part.Pos))
		x := sub.parseExpr()
		if sub.tok.Kind != lexer.EOF {
			sub.fail(lexer.CodeInterpExpr, "", "unexpected %s inside ${ }", sub.tok.Describe())
		}

		if !valueRef(x) {
			p.failAt(part.Pos, lexer.CodeInterpExpr, "move the expression to a let", "only names, members and indexes are allowed inside ${ }")
		}
		lit.Parts = append(lit.Parts, ast.TemplatePart{Expr: x})
	}

	return lit
}

// Target = identifier {DotSuffix | IndexSuffix} .
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

// ValueRef = identifier {DotSuffix | '[' (number | string | ValueRef) ']'} .
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

// ObjectOperand = PostfixExpr that is not a number, string, bool, null or array literal .
func objectOperand(x ast.Expr) bool {
	switch x.(type) {
	case *ast.NumberLit, *ast.StringLit, *ast.TemplateLit, *ast.BoolLit, *ast.NullLit, *ast.ArrayLit:
		return false
	}
	return true
}

func line(t lexer.Token) ast.Pos {
	return ast.Pos{Line: t.Pos.Line}
}

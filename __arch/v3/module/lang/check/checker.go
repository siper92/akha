package check

import (
	"context"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
)

var _ Checker = (*checker)(nil)

type symKind int

const (
	symLet symKind = iota
	symVar
	symLoop
	symInput
)

type symbol struct {
	kind    symKind
	mutable bool
	known   ast.Expr
}

type scope struct {
	parent  *scope
	names   map[string]*symbol
	pending map[string]bool
}

type checker struct {
	file   string
	values []ValueChecker
}

type run struct {
	ctx    context.Context
	values []ValueChecker
	scope  *scope
}

type bailout struct {
	err *Error
}

func New(file string) Checker {
	return &checker{file: file, values: Values()}
}

func (c *checker) Check(ctx context.Context, script *ast.Script) (err error) {
	r := &run{ctx: ctx, values: c.values}
	defer func() {
		switch p := recover().(type) {
		case nil:
		case bailout:
			p.err.File = c.file
			err = p.err
		default:
			panic(p)
		}
	}()

	r.push(script.Stmts)
	r.scope.names["input"] = &symbol{kind: symInput}
	r.stmts(script.Stmts)

	return nil
}

func (r *run) fail(pos ast.Pos, code, hint, format string, args ...any) {
	panic(bailout{err: newError(pos, code, hint, format, args...)})
}

func (r *run) push(stmts []ast.Stmt) {
	s := &scope{parent: r.scope, names: map[string]*symbol{}, pending: map[string]bool{}}
	for _, st := range stmts {
		switch st := st.(type) {
		case *ast.LetStmt:
			s.pending[st.Name] = true
		case *ast.VarStmt:
			s.pending[st.Name] = true
		}
	}
	r.scope = s
}

func (r *run) pop() {
	r.scope = r.scope.parent
}

func (r *run) declare(pos ast.Pos, name string, sym *symbol) {
	if name == "_" {
		return
	}
	if _, ok := r.scope.names[name]; ok {
		r.fail(pos, lexer.CodeRedeclared, "assign with name = value", "%s is already declared in this block", name)
	}
	r.scope.names[name] = sym
}

func (r *run) lookup(pos ast.Pos, name string) *symbol {
	for s := r.scope; s != nil; s = s.parent {
		if sym, ok := s.names[name]; ok {
			return sym
		}
	}
	for s := r.scope; s != nil; s = s.parent {
		if s.pending[name] {
			r.fail(pos, lexer.CodeUseBeforeDecl, "move the declaration up", "%s is used before its declaration", name)
		}
	}
	r.fail(pos, lexer.CodeUndeclared, "declare it with let or var", "%s is not declared", name)
	return nil
}

func (r *run) known(x ast.Expr) ast.Expr {
	return Resolve(x, func(name string) ast.Expr {
		for s := r.scope; s != nil; s = s.parent {
			if sym, ok := s.names[name]; ok {
				return sym.known
			}
		}
		return nil
	})
}

func (r *run) block(b *ast.Block) {
	r.push(b.Stmts)
	r.stmts(b.Stmts)
	r.pop()
}

func (r *run) stmts(stmts []ast.Stmt) {
	for i, st := range stmts {
		if err := r.ctx.Err(); err != nil {
			r.fail(st.Position(), lexer.CodeInternal, "", "check canceled: %v", err)
		}
		r.stmt(st)
		if _, ok := st.(*ast.ReturnStmt); ok && i < len(stmts)-1 {
			r.fail(stmts[i+1].Position(), lexer.CodeUnreachable, "remove the code after return", "unreachable code")
		}
	}
}

func (r *run) stmt(st ast.Stmt) {
	switch st := st.(type) {
	case *ast.LetStmt:
		r.expr(st.Value)
		r.declare(st.Pos, st.Name, &symbol{kind: symLet, known: r.known(st.Value)})
	case *ast.VarStmt:
		if st.Value != nil {
			r.expr(st.Value)
		}
		r.declare(st.Pos, st.Name, &symbol{kind: symVar, mutable: true})
	case *ast.AssignStmt:
		r.expr(st.Value)
		r.assign(st.Target)
	case *ast.ExprStmt:
		r.expr(st.X)
	case *ast.IfStmt:
		r.ifStmt(st)
	case *ast.ForInStmt:
		r.expr(st.Iter)
		r.push(nil)
		if st.Key != "" {
			r.declare(st.Pos, st.Key, &symbol{kind: symLoop, mutable: st.Mutable})
		}
		r.declare(st.Pos, st.Value, &symbol{kind: symLoop, mutable: st.Mutable})
		r.block(st.Body)
		r.pop()
	case *ast.ForRangeStmt:
		r.expr(st.Start)
		r.expr(st.End)
		r.push(nil)
		r.declare(st.Pos, st.Name, &symbol{kind: symLoop, mutable: st.Mutable})
		r.block(st.Body)
		r.pop()
	case *ast.ReturnStmt:
		if st.Value != nil {
			r.expr(st.Value)
		}
	}
}

func (r *run) ifStmt(st *ast.IfStmt) {
	r.expr(st.Cond)
	r.block(st.Then)
	switch e := st.Else.(type) {
	case *ast.IfStmt:
		r.ifStmt(e)
	case *ast.Block:
		r.block(e)
	}
}

func (r *run) assign(target ast.Expr) {
	root := target
	for {
		switch t := root.(type) {
		case *ast.MemberExpr:
			root = t.X
			continue
		case *ast.IndexExpr:
			r.expr(t.Index)
			root = t.X
			continue
		}
		break
	}

	id, ok := root.(*ast.Ident)
	if !ok {
		r.fail(target.Position(), lexer.CodeAssignTarget, "", "cannot assign to this expression")
	}

	sym := r.lookup(id.Pos, id.Name)
	if sym.mutable {
		return
	}

	switch sym.kind {
	case symInput:
		r.fail(id.Pos, lexer.CodeImmutable, "copy it with var x = input", "cannot assign to input")
	case symLoop:
		r.fail(id.Pos, lexer.CodeImmutable, "use for var to make loop variables mutable", "cannot assign to loop variable %s", id.Name)
	}
	r.fail(id.Pos, lexer.CodeImmutable, "use var for a mutable binding", "cannot assign to let %s", id.Name)
}

func (r *run) expr(x ast.Expr) {
	switch x := x.(type) {
	case *ast.Ident:
		r.lookup(x.Pos, x.Name)
	case *ast.TemplateLit:
		for _, part := range x.Parts {
			if part.Expr != nil {
				r.expr(part.Expr)
			}
		}
	case *ast.ArrayLit:
		for _, e := range x.Elems {
			r.expr(e)
		}
	case *ast.ObjectLit:
		for _, e := range x.Entries {
			r.expr(e.Value)
		}
	case *ast.UnaryExpr:
		r.expr(x.X)
	case *ast.BinaryExpr:
		r.expr(x.Left)
		r.expr(x.Right)
	case *ast.MemberExpr:
		r.expr(x.X)
	case *ast.IndexExpr:
		r.expr(x.X)
		r.expr(x.Index)
	case *ast.CallExpr:
		r.fail(x.Pos, lexer.CodeNoCallable, "calls come with modules in a later version", "unknown callee %s, v1 has no callables", x.Fn)
	}

	for _, v := range r.values {
		if err := v.CheckValue(r.ctx, x, r.known); err != nil {
			if ce, ok := err.(*Error); ok {
				panic(bailout{err: ce})
			}
			r.fail(x.Position(), lexer.CodeInternal, "", "%v", err)
		}
	}
}

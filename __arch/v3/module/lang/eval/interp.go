package eval

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/lexer"
)

var (
	_ Evaluator = (*evaluator)(nil)
	_ Interp    = (*interp)(nil)
	_ Target    = (*interp)(nil)
)

type evaluator struct {
	opts Options
}

type interp struct {
	opts  Options
	ops   Operators
	stmts int
	iters int
	ret   Value
}

func New(opts Options) Evaluator {
	return &evaluator{opts: opts}
}

func (e *evaluator) Run(ctx context.Context, script *ast.Script, input Value) (Value, error) {
	in := &interp{opts: e.opts, ops: NewOperators(), ret: Null{}}
	if input == nil {
		input = Null{}
	}

	root := NewEnv()
	if err := root.Declare("input", input.Clone(), false); err != nil {
		return nil, err
	}

	for _, s := range script.Stmts {
		flow, err := in.Exec(ctx, root, s)
		if err != nil {
			return nil, err
		}
		if flow == FlowReturn {
			break
		}
	}

	return in.ret, nil
}

func (in *interp) fail(pos ast.Pos, err error) error {
	var re *RuntimeError
	if !errors.As(err, &re) {
		re = newError(CodeInternal, "%v", err)
	}
	if re.Line == 0 {
		re.Line = pos.Line
		re.File = in.opts.File
	}
	return re
}

func (in *interp) tick(ctx context.Context, pos ast.Pos, count *int, max int, what string) error {
	if err := ctx.Err(); err != nil {
		return in.fail(pos, newError(CodeCanceled, "canceled: %v", err))
	}
	*count++
	if max > 0 && *count > max {
		return in.fail(pos, newError(CodeLimit, "max %s %d exceeded", what, max))
	}
	return nil
}

func (in *interp) store(pos ast.Pos, v Value) (Value, error) {
	v = v.Clone()
	if s, ok := v.(Sizer); ok && in.opts.MaxValueSize > 0 && s.Size() > in.opts.MaxValueSize {
		return nil, in.fail(pos, newError(CodeLimit, "max value size %d exceeded", in.opts.MaxValueSize))
	}
	return v, nil
}

func (in *interp) Exec(ctx context.Context, env Env, s ast.Stmt) (Flow, error) {
	pos := s.Position()
	if err := in.tick(ctx, pos, &in.stmts, in.opts.MaxStmts, "statements"); err != nil {
		return FlowNext, err
	}

	switch s := s.(type) {
	case *ast.LetStmt:
		return FlowNext, in.declare(ctx, env, pos, s.Name, s.Value, false)
	case *ast.VarStmt:
		return FlowNext, in.declare(ctx, env, pos, s.Name, s.Value, true)
	case *ast.AssignStmt:
		v, err := in.value(ctx, env, s.Value)
		if err != nil {
			return FlowNext, err
		}
		return FlowNext, in.Set(ctx, env, s.Target, v)
	case *ast.ExprStmt:
		_, err := in.Eval(ctx, env, s.X)
		return FlowNext, err
	case *ast.IfStmt:
		return in.ifStmt(ctx, env, s)
	case *ast.ForInStmt:
		return in.forIn(ctx, env, s)
	case *ast.ForRangeStmt:
		return in.forRange(ctx, env, s)
	case *ast.BreakStmt:
		return FlowBreak, nil
	case *ast.ContinueStmt:
		return FlowContinue, nil
	case *ast.ReturnStmt:
		in.ret = Null{}
		if s.Value != nil {
			v, err := in.value(ctx, env, s.Value)
			if err != nil {
				return FlowNext, err
			}
			in.ret = v
		}
		return FlowReturn, nil
	}

	return FlowNext, in.fail(pos, newError(CodeInternal, "unknown statement %T", s))
}

func (in *interp) ExecBlock(ctx context.Context, env Env, b *ast.Block) (Flow, error) {
	scope := env.Child()
	for _, s := range b.Stmts {
		flow, err := in.Exec(ctx, scope, s)
		if err != nil || flow != FlowNext {
			return flow, err
		}
	}
	return FlowNext, nil
}

func (in *interp) value(ctx context.Context, env Env, x ast.Expr) (Value, error) {
	v, err := in.Eval(ctx, env, x)
	if err != nil {
		return nil, err
	}
	return in.store(x.Position(), v)
}

func (in *interp) declare(ctx context.Context, env Env, pos ast.Pos, name string, x ast.Expr, mutable bool) error {
	var v Value = Null{}
	if x != nil {
		var err error
		if v, err = in.value(ctx, env, x); err != nil {
			return err
		}
	}
	if err := env.Declare(name, v, mutable); err != nil {
		return in.fail(pos, err)
	}
	return nil
}

func (in *interp) ifStmt(ctx context.Context, env Env, s *ast.IfStmt) (Flow, error) {
	cond, err := in.Eval(ctx, env, s.Cond)
	if err != nil {
		return FlowNext, err
	}
	if cond.Truth() {
		return in.ExecBlock(ctx, env, s.Then)
	}
	switch e := s.Else.(type) {
	case *ast.IfStmt:
		return in.ifStmt(ctx, env, e)
	case *ast.Block:
		return in.ExecBlock(ctx, env, e)
	}
	return FlowNext, nil
}

func (in *interp) forIn(ctx context.Context, env Env, s *ast.ForInStmt) (Flow, error) {
	v, err := in.Eval(ctx, env, s.Iter)
	if err != nil {
		return FlowNext, err
	}
	iter, ok := v.Clone().(Iterable)
	if !ok {
		return FlowNext, in.fail(s.Pos, newError(CodeNotIterable, "cannot iterate %s", v.Kind()))
	}

	for _, item := range iter.Items() {
		loop := env.Child()
		switch {
		case s.Key != "":
			_ = loop.Declare(s.Key, item.Key, s.Mutable)
			_ = loop.Declare(s.Value, item.Value.Clone(), s.Mutable)
		case v.Kind() == KindObject:
			_ = loop.Declare(s.Value, item.Key, s.Mutable)
		default:
			_ = loop.Declare(s.Value, item.Value.Clone(), s.Mutable)
		}

		flow, err := in.iterate(ctx, loop, s.Pos, s.Body)
		if err != nil || flow == FlowReturn {
			return flow, err
		}
		if flow == FlowBreak {
			break
		}
	}
	return FlowNext, nil
}

func (in *interp) forRange(ctx context.Context, env Env, s *ast.ForRangeStmt) (Flow, error) {
	start, err := in.bound(ctx, env, s.Start)
	if err != nil {
		return FlowNext, err
	}
	end, err := in.bound(ctx, env, s.End)
	if err != nil {
		return FlowNext, err
	}

	for i := start; i < end; i++ {
		loop := env.Child()
		_ = loop.Declare(s.Name, Number(i), s.Mutable)

		flow, err := in.iterate(ctx, loop, s.Pos, s.Body)
		if err != nil || flow == FlowReturn {
			return flow, err
		}
		if flow == FlowBreak {
			break
		}
	}
	return FlowNext, nil
}

func (in *interp) iterate(ctx context.Context, loop Env, pos ast.Pos, body *ast.Block) (Flow, error) {
	if err := in.tick(ctx, pos, &in.iters, in.opts.MaxIterations, "iterations"); err != nil {
		return FlowNext, err
	}
	flow, err := in.ExecBlock(ctx, loop, body)
	if flow == FlowContinue {
		flow = FlowNext
	}
	return flow, err
}

func (in *interp) bound(ctx context.Context, env Env, x ast.Expr) (float64, error) {
	v, err := in.Eval(ctx, env, x)
	if err != nil {
		return 0, err
	}
	n, ok := v.(Number)
	if !ok {
		return 0, in.fail(x.Position(), newError(CodeKindMismatch, "range bound must be a number, got %s", v.Kind()))
	}
	if float64(n) != math.Trunc(float64(n)) {
		return 0, in.fail(x.Position(), newError(CodeNonIntegral, "range bound %s is not integral", n))
	}
	return float64(n), nil
}

func (in *interp) Set(ctx context.Context, env Env, t ast.Expr, v Value) error {
	switch t := t.(type) {
	case *ast.Ident:
		if err := env.Assign(t.Name, v); err != nil {
			return in.fail(t.Pos, err)
		}
		return nil
	case *ast.MemberExpr:
		if err := in.mutableRoot(env, t); err != nil {
			return err
		}
		x, err := in.Eval(ctx, env, t.X)
		if err != nil {
			return err
		}
		ms, ok := x.(MemberSetter)
		if !ok {
			return in.fail(t.Pos, accessError(x, "set member "+t.Name+" on"))
		}
		return in.failIf(t.Pos, ms.SetMember(t.Name, v))
	case *ast.IndexExpr:
		if err := in.mutableRoot(env, t); err != nil {
			return err
		}
		x, err := in.Eval(ctx, env, t.X)
		if err != nil {
			return err
		}
		idx, err := in.Eval(ctx, env, t.Index)
		if err != nil {
			return err
		}
		is, ok := x.(IndexSetter)
		if !ok {
			return in.fail(t.Pos, accessError(x, "set an index on"))
		}
		return in.failIf(t.Pos, is.SetIndex(idx, v))
	}
	return in.fail(t.Position(), newError(CodeInternal, "cannot assign to %s", t))
}

func (in *interp) mutableRoot(env Env, t ast.Expr) error {
	root := t
	for {
		switch r := root.(type) {
		case *ast.MemberExpr:
			root = r.X
			continue
		case *ast.IndexExpr:
			root = r.X
			continue
		}
		break
	}
	id, ok := root.(*ast.Ident)
	if !ok {
		return in.fail(t.Position(), newError(CodeInternal, "cannot assign to %s", t))
	}
	b, ok := env.Lookup(id.Name)
	if !ok {
		return in.fail(id.Pos, newError(CodeUndeclared, "%s is not declared", id.Name))
	}
	if !b.Mutable {
		return in.fail(id.Pos, newError(CodeImmutable, "cannot assign to %s", id.Name))
	}
	return nil
}

func (in *interp) failIf(pos ast.Pos, err error) error {
	if err != nil {
		return in.fail(pos, err)
	}
	return nil
}

func (in *interp) Eval(ctx context.Context, env Env, x ast.Expr) (Value, error) {
	switch x := x.(type) {
	case *ast.NumberLit:
		return Number(x.Value), nil
	case *ast.StringLit:
		return String(x.Value), nil
	case *ast.BoolLit:
		return Bool(x.Value), nil
	case *ast.NullLit:
		return Null{}, nil
	case *ast.Ident:
		b, ok := env.Lookup(x.Name)
		if !ok {
			return nil, in.fail(x.Pos, newError(CodeUndeclared, "%s is not declared", x.Name))
		}
		return b.Value, nil
	case *ast.TemplateLit:
		return in.template(ctx, env, x)
	case *ast.ArrayLit:
		arr := &Array{Elems: make([]Value, 0, len(x.Elems))}
		for _, e := range x.Elems {
			v, err := in.value(ctx, env, e)
			if err != nil {
				return nil, err
			}
			arr.Elems = append(arr.Elems, v)
		}
		return arr, nil
	case *ast.ObjectLit:
		obj := NewObject()
		for _, e := range x.Entries {
			v, err := in.value(ctx, env, e.Value)
			if err != nil {
				return nil, err
			}
			_ = obj.SetMember(e.Key, v)
		}
		return obj, nil
	case *ast.UnaryExpr:
		v, err := in.Eval(ctx, env, x.X)
		if err != nil {
			return nil, err
		}
		r, err := in.ops.Unary(x.Op, v)
		return r, in.failIf(x.Pos, err)
	case *ast.BinaryExpr:
		return in.binary(ctx, env, x)
	case *ast.MemberExpr:
		v, err := in.Eval(ctx, env, x.X)
		if err != nil {
			return nil, err
		}
		m, ok := v.(Memberer)
		if !ok {
			return nil, in.fail(x.Pos, accessError(v, "access member "+x.Name+" on"))
		}
		r, err := m.Member(x.Name)
		return r, in.failIf(x.Pos, err)
	case *ast.IndexExpr:
		v, err := in.Eval(ctx, env, x.X)
		if err != nil {
			return nil, err
		}
		idx, err := in.Eval(ctx, env, x.Index)
		if err != nil {
			return nil, err
		}
		ix, ok := v.(Indexer)
		if !ok {
			return nil, in.fail(x.Pos, accessError(v, "index"))
		}
		r, err := ix.Index(idx)
		return r, in.failIf(x.Pos, err)
	case *ast.CallExpr:
		return nil, in.fail(x.Pos, newError(CodeNoCallable, "no callables in v1"))
	}
	return nil, in.fail(x.Position(), newError(CodeInternal, "unknown expression %T", x))
}

func (in *interp) binary(ctx context.Context, env Env, x *ast.BinaryExpr) (Value, error) {
	l, err := in.Eval(ctx, env, x.Left)
	if err != nil {
		return nil, err
	}
	switch x.Op {
	case lexer.And:
		if !l.Truth() {
			return Bool(false), nil
		}
	case lexer.Or:
		if l.Truth() {
			return Bool(true), nil
		}
	}

	r, err := in.Eval(ctx, env, x.Right)
	if err != nil {
		return nil, err
	}
	if x.Op == lexer.And || x.Op == lexer.Or {
		return Bool(r.Truth()), nil
	}

	v, err := in.ops.Binary(x.Op, l, r)
	return v, in.failIf(x.Pos, err)
}

func (in *interp) template(ctx context.Context, env Env, x *ast.TemplateLit) (Value, error) {
	var b strings.Builder
	scope := env.Child()
	for _, part := range x.Parts {
		if part.Expr == nil {
			b.WriteString(part.Text)
			continue
		}
		v, err := in.Eval(ctx, scope, part.Expr)
		if err != nil {
			return nil, err
		}
		b.WriteString(format(v))
	}
	return String(b.String()), nil
}

func accessError(v Value, what string) *RuntimeError {
	if v.Kind() == KindNull {
		return newError(CodeNullAccess, "access on null")
	}
	return newError(CodeMemberAccess, "cannot %s %s", what, v.Kind())
}

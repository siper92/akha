package types

import (
	"github.com/siper92/akha/lang/ast"
	"github.com/siper92/akha/lang/eval"
	"github.com/siper92/akha/lang/module"
)

var (
	_ module.IType = String{}
	_ module.IType = Any{}
)

type String struct{}

type Any struct{}

func (Any) Name() string { return "value" }

func (Any) Accepts(ast.Expr) bool { return true }

func (Any) Valid(v module.IValue) bool {
	_, ok := v.(eval.Value)
	return ok
}

func (String) Name() string { return "string" }

func (String) Accepts(known ast.Expr) bool {
	switch known.(type) {
	case nil, *ast.StringLit:
		return true
	default:
		return false
	}
}

func (String) Valid(v module.IValue) bool {
	_, ok := v.(eval.String)
	return ok
}

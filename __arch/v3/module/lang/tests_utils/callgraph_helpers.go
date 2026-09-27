package tests_utils

import (
	goast "go/ast"
	goparser "go/parser"
	gotoken "go/token"
	"testing"
)

type CallGraph struct {
	Calls    map[string]map[string]bool
	Breaking map[string]bool
}

func ReadCallGraph(t *testing.T, path, recvType, recvVar, breaker string) CallGraph {
	t.Helper()
	file, err := goparser.ParseFile(gotoken.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	g := CallGraph{Calls: make(map[string]map[string]bool), Breaking: make(map[string]bool)}
	for _, decl := range file.Decls {
		fn, ok := decl.(*goast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		current := methodName(fn, recvType)
		if current == "" {
			continue
		}
		goast.Inspect(fn.Body, func(n goast.Node) bool {
			call, ok := n.(*goast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*goast.SelectorExpr)
			if !ok {
				return true
			}
			if id, ok := sel.X.(*goast.Ident); !ok || id.Name != recvVar {
				return true
			}
			if sel.Sel.Name == breaker {
				g.Breaking[current] = true
			}
			if g.Calls[current] == nil {
				g.Calls[current] = make(map[string]bool)
			}
			g.Calls[current][sel.Sel.Name] = true
			return true
		})
	}

	return g
}

func methodName(fn *goast.FuncDecl, recvType string) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	star, ok := fn.Recv.List[0].Type.(*goast.StarExpr)
	if !ok {
		return ""
	}
	if id, ok := star.X.(*goast.Ident); ok && id.Name == recvType {
		return fn.Name.Name
	}

	return ""
}

func RunCallGraphCycles(t *testing.T, path, recvType, recvVar, breaker string) {
	t.Helper()
	g := ReadCallGraph(t, path, recvType, recvVar, breaker)
	if len(g.Breaking) == 0 {
		t.Fatalf("no method of %s calls %s.%s", recvType, recvVar, breaker)
	}
	for root := range g.Calls {
		color := make(map[string]int)
		var stack []string
		var visit func(string)
		visit = func(f string) {
			switch color[f] {
			case 0:
				color[f] = 1
				if !g.Breaking[f] {
					stack = append(stack, f)
					for callee := range g.Calls[f] {
						visit(callee)
					}
					stack = stack[:len(stack)-1]
				}
				color[f] = 2
			case 1:
				t.Errorf("cycle through %s without %s.%s: %v", f, recvVar, breaker, stack)
			}
		}
		visit(root)
	}
}

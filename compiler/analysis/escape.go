package analysis

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/types"
)

func EscapingObjects(n ast.Node, info *types.Info) map[*types.Var]bool {
	v := escapeAnalysis{
		info:         info,
		escaping:     make(map[*types.Var]bool),
		topScope:     info.Scopes[n],
		bottomScopes: make(map[*types.Scope]bool),
	}
	ast.Walk(&v, n)
	return v.escaping
}

type escapeAnalysis struct {
	info         *types.Info
	escaping     map[*types.Var]bool
	topScope     *types.Scope
	bottomScopes map[*types.Scope]bool
}

func (v *escapeAnalysis) Visit(node ast.Node) (w ast.Visitor) {
	// huge overapproximation
	switch n := node.(type) {
	case *ast.UnaryExpr:
		if n.Op == token.AND {
			if _, ok := n.X.(*ast.Ident); ok {
				return &escapingObjectCollector{v}
			}
		}
	case *ast.FuncLit:
		v.bottomScopes[v.info.Scopes[n.Type]] = true
		return &escapingObjectCollector{v}
	case *ast.ForStmt, *ast.RangeStmt:
		v.bottomScopes[v.info.Scopes[n]] = true
	}
	return v
}

type escapingObjectCollector struct {
	analysis *escapeAnalysis
}

func (v *escapingObjectCollector) Visit(node ast.Node) (w ast.Visitor) {
	if id, ok := node.(*ast.Ident); ok {
		if obj, ok := v.analysis.info.Uses[id].(*types.Var); ok {
			for s := obj.Parent(); s != nil; s = s.Parent() {
				if s == v.analysis.topScope {
					v.analysis.escaping[obj] = true
					break
				}
				if v.analysis.bottomScopes[s] {
					break
				}
			}
		}
	}
	return v
}

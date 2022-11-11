package visitors

import (
	"github.com/midbel/buddy/ast"
)

type loopVisitor struct {
	level int
}

func Loop() Visitor {
	return &loopVisitor{}
}

func (v *loopVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	return expr, v.visit(expr)
}

func (v *loopVisitor) visit(expr ast.Expression) error {
	switch expr.(type) {
	case ast.Literal:
	case ast.Double:
	case ast.Integer:
	case ast.Boolean:
	case ast.Variable:
	case ast.Array:
	case ast.Dict:
	case ast.Index:
	case ast.Slice:
	case ast.Path:
	case ast.Call:
	case ast.Parameter:
	case ast.Assert:
	case ast.Assign:
	case ast.Unary:
	case ast.Binary:
	case ast.ListComp:
	case ast.DictComp:
	case ast.Test:
	case ast.While:
	case ast.For:
	case ast.ForEach:
	case ast.Import:
	case ast.Script:
	case ast.Function:
	case ast.Return:
	case ast.Break:
	case ast.Continue:
	default:
	}
	return nil
}

func (v *loopVisitor) enter() {
	v.level++
}

func (v *loopVisitor) leave() {
	v.level--
}

func (v *loopVisitor) inLoop() bool {
	return v.level > 0
}

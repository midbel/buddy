package visitors

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
)

type importVisitor struct {
	env *Counter[string]
	list  faults.ErrorList
	limit int
}

func Import() Visitor {
	return &importVisitor{
		env: EmptyCounter[string](),
		list:  make(faults.ErrorList, 0, faults.MaxErrorCount),
		limit: faults.MaxErrorCount,
	}
}

func (i *importVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	return expr, nil
}

func (i *importVisitor) visit(expr ast.Expression) error {
	switch e := expr.(type) {
	case ast.Literal:
	case ast.Double:
	case ast.Integer:
	case ast.Boolean:
	case ast.Variable:
	case ast.Array:
	case ast.Dict:
	case ast.Index:
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
	case ast.Return:
	case ast.Break:
	case ast.Continue:
	default:
	}
	return nil
}

func unusedImport(ident string) error {
	return fmt.Errorf("%s: module imported but not used", ident)
}
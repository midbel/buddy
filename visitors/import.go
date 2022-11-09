package visitors

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
)

type importVisitor struct {
	env   *Counter[string]
	list  faults.ErrorList
	limit int
}

func Import() Visitor {
	return &importVisitor{
		env:   EmptyCounter[string](),
		list:  make(faults.ErrorList, 0, faults.MaxErrorCount),
		limit: faults.MaxErrorCount,
	}
}

func (v *importVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	err := v.visit(expr)
	v.unused()
	if err == nil && v.list.Size() > 0 {
		err = &v.list
	}
	return expr, err
}

func (v *importVisitor) visit(expr ast.Expression) error {
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
		if err := v.exists(e.Ident); err != nil {
			v.list.Append(err)
		} else {
			v.env.Incr(e.Ident)
		}
		if err := v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.Call:
	case ast.Parameter:
	case ast.Assert:
	case ast.Assign:
		if err := v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.Unary:
	case ast.Binary:
	case ast.ListComp:
	case ast.DictComp:
	case ast.Test:
	case ast.While:
	case ast.For:
	case ast.ForEach:
	case ast.Import:
		v.env.Incr(e.Alias)
	case ast.Script:
		for i := range e.List {
			if err := v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Return:
		if err := v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.Break:
	case ast.Continue:
	default:
	}
	if !v.noLimit() && v.list.Size() > v.limit {
		return &v.list
	}
	return nil
}

func (v importVisitor) exists(ident string) error {
	ok := v.env.Exists(ident)
	if !ok {
		return undefinedImport(ident)
	}
	return nil
}

func (v importVisitor) unused() {
	// TBW
}

func (v importVisitor) noLimit() bool {
	return v.limit <= 0
}

func unusedImport(ident string) error {
	return fmt.Errorf("%s: module imported but not used", ident)
}

func undefinedImport(ident string) error {
	return fmt.Errorf("%s: module used but not imported!", ident)
}

package visitors

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
)

type variableVisitor struct {
	env   map[string]int
	list  faults.ErrorList
	limit int
}

func Variable() Visitor {
	return &variableVisitor{
		env:   make(map[string]int),
		list:  make(faults.ErrorList, 0, faults.MaxErrorCount),
		limit: faults.MaxErrorCount,
	}
}

func (v *variableVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	err := v.visit(expr)
	v.unused()
	if err == nil && v.list.Size() > 0 {
		err = &v.list
	}
	return expr, err
}

func (v *variableVisitor) visit(expr ast.Expression) error {
	var err error
	switch e := expr.(type) {
	case ast.Literal:
		// PASS: to be removed later
	case ast.Double:
		// PASS: to be removed later
	case ast.Integer:
		// PASS: to be removed later
	case ast.Boolean:
		// PASS: to be removed later
	case ast.Variable:
		err = v.exists(e.Ident)
		if err != nil {
			v.list.Append(err)
		}
	case ast.Array:
		for i := range e.List {
			if err = v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Dict:
		for i := range e.List {
			if err = v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Index:
	case ast.Path:
	case ast.Call:
		for i := range e.Args {
			if err = v.visit(e.Args[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Parameter:
		if err = v.visit(e.Expr); err != nil {
			v.list.Append(err)
		}
	case ast.Assert:
		if err = v.visit(e.Expr); err != nil {
			v.list.Append(err)
		}
	case ast.Assign:
		err = v.visit(e.Right)
		if err != nil {
			v.list.Append(err)
		}
		if i, ok := e.Ident.(ast.Variable); ok {
			v.set(i.Ident)
		}
	case ast.Unary:
		err = v.visit(e.Right)
		if err != nil {
			v.list.Append(err)
		}
	case ast.Binary:
		if err = v.visit(e.Left); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.ListComp:
		// sub := Variable()
		// for i := range e.List {
		// 	if _, err = sub.Visit(e.List[i]); err != nil {
		// 		v.list.Append(err)
		// 	}
		// }
		// if _, err = sub.Visit(e.Body); err != nil {
		// 	v.list.Append(err)
		// }
	case ast.DictComp:
		// sub := Variable()
		// for i := range e.List {
		// 	if _, err = sub.Visit(e.List[i]); err != nil {
		// 		v.list.Append(err)
		// 	}
		// }
		// if _, err = sub.Visit(e.Key); err != nil {
		// 	v.list.Append(err)
		// }
		// if _, err = sub.Visit(e.Val); err != nil {
		// 	v.list.Append(err)
		// }
	case ast.CompItem:
	case ast.Test:
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		sub1 := Variable()
		if _, err = sub1.Visit(e.Csq); err != nil {
			v.list.Append(err)
		}
		sub2 := Variable()
		if _, err = sub2.Visit(e.Alt); err != nil {
			v.list.Append(err)
		}
	case ast.While:
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		sub := Variable()
		if _, err = sub.Visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.For:
		sub := Variable()
		if _, err = sub.Visit(e.Init); err != nil {
			v.list.Append(err)
		}
		if _, err = sub.Visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		if _, err = sub.Visit(e.Incr); err != nil {
			v.list.Append(err)
		}
		sub2 := Variable()
		if _, err = sub2.Visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.ForEach:
		if err = v.visit(e.Iter); err != nil {
			v.list.Append(err)
		}
		sub := Variable()
		if _, err = sub.Visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.Import:
		// PASS: to be removed later
	case ast.Script:
		for i := range e.List {
			err = v.visit(e.List[i])
			if err != nil {
				v.list.Append(err)
			}
		}
	case ast.Function:
	case ast.Return:
		if err = v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.Break:
	case ast.Continue:
	default:
		// PASS: unknown/unsupported expression type and/or nil Expression
	}
	if !v.noLimit() && v.list.Size() > v.limit {
		return &v.list
	}
	return nil
}

func (v variableVisitor) unused() {
	for id := range v.env {
		c := v.env[id]
		if c == 1 {
			v.list.Append(unusedVar(id))
		}
	}
}

func (v variableVisitor) exists(ident string) error {
	_, ok := v.env[ident]
	if !ok {
		return undefinedVar(ident)
	}
	return nil
}

func (v variableVisitor) set(ident string) {
	v.env[ident]++
}

func (v variableVisitor) noLimit() bool {
	return v.limit <= 0
}

func undefinedVar(ident string) error {
	return fmt.Errorf("%s: variable used before being defined!", ident)
}

func unusedVar(ident string) error {
	return fmt.Errorf("%s: variable defined but not used", ident)
}

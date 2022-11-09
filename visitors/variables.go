package visitors

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
)

type variableVisitor struct {
	env   *Counter[string]
	list  faults.ErrorList
	limit int
}

func Variable() Visitor {
	return &variableVisitor{
		env:   EmptyCounter[string](),
		list:  make(faults.ErrorList, 0, faults.MaxErrorCount),
		limit: faults.MaxErrorCount,
	}
}

func (v *variableVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	err := v.visit(expr)
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
		if i, ok := e.Arr.(ast.Variable); ok {
			if err = v.exists(i.Ident); err != nil {
				v.list.Append(err)
			}
		}
		for i := range e.List {
			if err = v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Slice:
		if err = v.visit(e.Start); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.End); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Step); err != nil {
			v.list.Append(err)
		}
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
			v.env.Add(i.Ident)
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
		v.enter()
		defer v.leave()
		for i := range e.List {
			if err = v.visit(e.List[i].Iter); err != nil {
				v.list.Append(err)
			}
			v.env.Add(e.List[i].Ident)
			for j := range e.List[i].Cdt {
				if err = v.visit(e.List[i].Cdt[j]); err != nil {
					v.list.Append(err)
				}
			}
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.DictComp:
		v.enter()
		defer v.leave()
		for i := range e.List {
			if err = v.visit(e.List[i].Iter); err != nil {
				v.list.Append(err)
			}
			v.env.Add(e.List[i].Ident)
			for j := range e.List[i].Cdt {
				if err = v.visit(e.List[i].Cdt[j]); err != nil {
					v.list.Append(err)
				}
			}
		}
		if err = v.visit(e.Key); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Val); err != nil {
			v.list.Append(err)
		}
	case ast.Test:
		v.enter()
		defer v.leave()
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Csq); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Alt); err != nil {
			v.list.Append(err)
		}
	case ast.While:
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		v.enter()
		defer v.leave()
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.For:
		v.enter()
		defer v.leave()
		if err = v.visit(e.Init); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Incr); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.ForEach:
		v.enter()
		defer v.leave()
		if err = v.visit(e.Iter); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Body); err != nil {
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
		v.enter()
		defer v.leave()
		for i := range e.Params {
			p, ok := e.Params[i].(ast.Parameter)
			if !ok {
				continue
			}
			v.env.Add(p.Ident)
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
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

func (v *variableVisitor) enter() {
	v.env = v.env.Wrap()
}

func (v *variableVisitor) leave() {
	v.env = v.env.Unwrap()
}

func (v variableVisitor) exists(ident string) error {
	ok := v.env.Exists(ident)
	if !ok {
		return undefinedVar(ident)
	}
	return nil
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

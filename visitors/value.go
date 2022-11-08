package visitors

import (
	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/types"
)

type valueVisitor struct {
	*types.Environ
}

func Value() Visitor {
	return valueVisitor{
		Environ: types.EmptyEnv(),
	}
}

func (v valueVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	return v.visit(expr, v)
}

func (v valueVisitor) visit(expr ast.Expression, ctx types.Context) (ast.Expression, error) {
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
		if res, err1 := ctx.Resolve(ast.Ident); err1 == nil {
			e, err = ast.CreatePrimitive(res.Raw())
		}
		return e, err
	case ast.Array:
		for i := range e.List {
			if e.List[i], err = v.visit(e.List[i], ctx); err != nil {
				break
			}
		}
		return e, err
	case ast.Dict:
		for k := range e.List {
			if e.List[k], err = v.visit(e.List[k], ctx); err != nil {
				break
			}
		}
		return e, err
	case ast.Index:
		if e.Arr, err = v.visit(e.Arr, ctx); err != nil {
			return nil, err
		}
		for i := range e.List {
			if e.List[i], err = v.visit(e.List[i], ctx); err != nil {
				break
			}
		}
		return e, nil
	case ast.Path:
		e.Right, err = v.visit(e.Right, ctx)
		return e, err
	case ast.Call:
		for i := range e.Args {
			if e.Args[i], err = v.visit(e.Args[i], ctx); err != nil {
				break
			}
		}
		return e, nil
	case ast.Parameter:
		e.Expr, err = v.visit(e.Expr, ctx)
		return e, err
	case ast.Assert:
		e.Expr, err = v.visit(e.Expr, ctx)
		return e, err
	case ast.Assign:
		e.Right, err = v.visit(e.Right, ctx)
		return e, err
	case ast.Unary:
		if e.Right, err = v.visit(e.Right, ctx); err != nil {
			return nil, err
		}
		if e.Right.IsValue() {

		}
		return e, err
	case ast.Binary:
		if e.Left, err = v.visit(e.Left, ctx); err != nil {
			return nil, err
		}
		if e.Right, err = v.visit(e.Right, ctx); err != nil {
			return nil, err
		}
		if e.Left.IsValue() && e.Right.IsValue() {

		}
		return e, err
	case ast.ListComp:
		if e.Body, err = v.visit(e.Body, ctx); err != nil {
			return err
		}
		for i := range e.List {
			if e.List[i], err = v.visit(e.List[i], ctx); err != nil {
				break
			}
		}
		return e, err
	case ast.DictComp:
		if e.Key, err = v.visit(e.Key, ctx); err != nil {
			return err
		}
		if e.Val, err = v.visit(e.Val, ctx); err != nil {
			return err
		}
		for i := range e.List {
			if e.List[i], err = v.visit(e.List[i], ctx); err != nil {
				break
			}
		}
		return e, err
	case ast.CompItem:
		if e.Iter, e = v.visit(e.Iter, ctx); err != nil {
			return nil, err
		}
		for i := range e.Cdt {
			if e.Cdt[i], err = v.visit(e.Cdt[i], ctx); err != nil {
				break
			}
		}
		return e, nil
	case ast.Test:
		if e.Cdt, err = v.visit(e.Cdt, ctx); err != nil {
			return nil, err
		}
		if e.Csq, err = v.visit(e.Csq, ctx); err != nil {
			return nil, err
		}
		e.Alt, err = v.visit(e.Alt, ctx)
		return e, err
	case ast.While:
		if e.Cdt, err = v.visit(e.Cdt, ctx); err != nil {
			return nil, err
		}
		e.Body, err = v.visit(e.Body, ctx)
		return e, err
	case ast.For:
		if e.Init, err = v.visit(e.Init, ctx); err != nil {
			return nil, err
		}
		if e.Cdt, err = v.visit(e.Cdt, ctx); err != nil {
			return nil, err
		}
		if e.Incr, err = v.visit(e.Incr, ctx); err != nil {
			return nil, err
		}
		e.Body, err = v.visit(e.Body, ctx)
		return e, err
	case ast.ForEach:
		if e.Iter, err = v.visit(e.Iter, ctx); err != nil {
			return nil, err
		}
		e.Body, err = v.visit(e.Body, ctx)
		return e, err
	case ast.Import:
		// PASS: to be removed later
	case ast.Script:
		for i := range e.List {
			if e.List[i], err = v.visit(e.List[i], ctx); err != nil {
				break
			}
		}
		return e, err
	case ast.Function:
		e.Body, err = v.visit(e.Body, ctx)
		return e, err
	case ast.Return:
		e.Right, err = v.visit(e.Right, ctx)
		return e, err
	case ast.Break:
		// PASS: to be removed later
	case ast.Continue:
		// PASS: to be removed later
	default:
		// PASS: unknown/unsupported expression type and/or nil Expression
	}
	return expr, nil
}

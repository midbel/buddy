package visitors

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/token"
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
	// TODO: create sub Environ for Test, For, While, ForEach, Function
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
		if res, err := ctx.Resolve(e.Ident); err == nil {
			return ast.CreatePrimitive(e.Token, res.Raw())
		}
		return e, nil
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
	case ast.Slice:
		if e.Start, err = v.visit(e.Start, ctx); err != nil {
			return nil, err
		}
		if e.End, err = v.visit(e.End, ctx); err != nil {
			return nil, err
		}
		if e.Step, err = v.visit(e.Step, ctx); err != nil {
			return nil, err
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
		if res, err := evalExpression(e.Right); err == nil {
			if v, ok := e.Ident.(ast.Variable); ok {
				ctx.Define(v.Ident, res)
			}
		}
		return e, err
	case ast.Unary:
		if e.Right, err = v.visit(e.Right, ctx); err != nil {
			return nil, err
		}
		if e.Right.IsValue() {
			return evalUnary(e, ctx)
		}
		return e, nil
	case ast.Binary:
		if e.Left, err = v.visit(e.Left, ctx); err != nil {
			return nil, err
		}
		if e.Right, err = v.visit(e.Right, ctx); err != nil {
			return nil, err
		}
		if e.Left.IsValue() && e.Right.IsValue() {
			return evalBinary(e, ctx)
		}
		return e, nil
	case ast.ListComp:
		if e.Body, err = v.visit(e.Body, ctx); err != nil {
			return e, err
		}
		var ci ast.Expression
		for i := range e.List {
			ci, err = v.visit(e.List[i], ctx)
			if err != nil {
				break
			}
			e.List[i] = ci.(ast.CompItem)
		}
		return e, err
	case ast.DictComp:
		if e.Key, err = v.visit(e.Key, ctx); err != nil {
			return nil, err
		}
		if e.Val, err = v.visit(e.Val, ctx); err != nil {
			return nil, err
		}
		var ci ast.Expression
		for i := range e.List {
			ci, err = v.visit(e.List[i], ctx)
			if err != nil {
				break
			}
			e.List[i] = ci.(ast.CompItem)
		}
		return e, err
	case ast.CompItem:
		if e.Iter, err = v.visit(e.Iter, ctx); err != nil {
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

func evalExpression(e ast.Expression) (types.Primitive, error) {
	var res types.Primitive
	switch e := e.(type) {
	case ast.Literal:
		res = types.CreateString(e.Str)
	case ast.Double:
		res = types.CreateFloat(e.Value)
	case ast.Integer:
		res = types.CreateInt(e.Value)
	case ast.Boolean:
		res = types.CreateBool(e.Value)
	default:
		return nil, fmt.Errorf("expression is not a value")
	}
	return res, nil
}

var unaryActions = map[rune]func(ast.Unary) ast.Expression{
	token.Sub:    evalRev,
	token.Not:    evalNot,
	token.BinNot: evalBnot,
}

func evalUnary(u ast.Unary, ctx types.Context) (ast.Expression, error) {
	fn, ok := unaryActions[u.Op]
	if !ok {
		return nil, fmt.Errorf("unary operator not recognized")
	}
	e := fn(u)
	if e == nil {
		return u, nil
	}
	return e, nil
}

func evalNot(u ast.Unary) ast.Expression {
	var b bool
	switch r := u.Right.(type) {
	case ast.Literal:
		b = len(r.Str) == 0
	case ast.Double:
		b = r.Value == 0
	case ast.Integer:
		b = r.Value == 0
	case ast.Boolean:
		b = !r.Value
	default:
		return u
	}
	return ast.CreateBoolean(u.Token, b)
}

func evalRev(u ast.Unary) ast.Expression {
	switch r := u.Right.(type) {
	case ast.Double:
		r.Value = -r.Value
		return r
	case ast.Integer:
		r.Value = -r.Value
		return r
	default:
		return u
	}
}

func evalBnot(u ast.Unary) ast.Expression {
	x, ok := u.Right.(ast.Integer)
	if ok {
		x.Value = ^x.Value
		return x
	}
	return u
}

var binaryActions = map[rune]func(ast.Binary) ast.Expression{
	token.Add:    evalAdd,
	token.Sub:    evalSub,
	token.Mul:    evalMul,
	token.Div:    evalDiv,
	token.Pow:    evalPow,
	token.Mod:    evalMod,
	token.Lshift: evalLshift,
	token.Rshift: evalRshift,
	token.BinAnd: evalBand,
	token.BinOr:  evalBor,
	token.Eq:     evalEq,
	token.Ne:     evalNe,
	token.Lt:     evalLt,
	token.Le:     evalLe,
	token.Gt:     evalGt,
	token.Ge:     evalGe,
	token.And:    evalAnd,
	token.Or:     evalOr,
}

func evalBinary(b ast.Binary, ctx types.Context) (ast.Expression, error) {
	fn, ok := binaryActions[b.Op]
	if !ok {
		return nil, fmt.Errorf("binary operator not recognized")
	}
	e := fn(b)
	if e == nil {
		return b, nil
	}
	return e, nil
}

func evalAdd(b ast.Binary) ast.Expression {
	add, ok := b.Left.(interface {
		Add(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return add.Add(b.Right)
}

func evalSub(b ast.Binary) ast.Expression {
	sub, ok := b.Left.(interface {
		Sub(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return sub.Sub(b.Right)
}

func evalDiv(b ast.Binary) ast.Expression {
	div, ok := b.Left.(interface {
		Div(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return div.Div(b.Right)
}

func evalMul(b ast.Binary) ast.Expression {
	mul, ok := b.Left.(interface {
		Mul(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return mul.Mul(b.Right)
}

func evalMod(b ast.Binary) ast.Expression {
	mod, ok := b.Left.(interface {
		Mod(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return mod.Mod(b.Right)
}

func evalPow(b ast.Binary) ast.Expression {
	pow, ok := b.Left.(interface {
		Pow(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return pow.Pow(b.Right)
}

func evalLshift(b ast.Binary) ast.Expression {
	shift, ok := b.Left.(interface {
		Lshift(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return shift.Lshift(b.Right)
}

func evalRshift(b ast.Binary) ast.Expression {
	shift, ok := b.Left.(interface {
		Rshift(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return shift.Rshift(b.Right)
}

func evalBand(b ast.Binary) ast.Expression {
	bin, ok := b.Left.(interface {
		And(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return bin.And(b.Right)
}

func evalBor(b ast.Binary) ast.Expression {
	bin, ok := b.Left.(interface {
		Or(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return bin.Or(b.Right)
}

func evalEq(b ast.Binary) ast.Expression {
	eq, ok := b.Left.(interface {
		Eq(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return eq.Eq(b.Right)
}

func evalNe(b ast.Binary) ast.Expression {
	eq, ok := b.Left.(interface {
		Ne(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return eq.Ne(b.Right)
}

func evalLt(b ast.Binary) ast.Expression {
	eq, ok := b.Left.(interface {
		Lt(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return eq.Lt(b.Right)
}

func evalLe(b ast.Binary) ast.Expression {
	eq, ok := b.Left.(interface {
		Le(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return eq.Le(b.Right)
}

func evalGt(b ast.Binary) ast.Expression {
	eq, ok := b.Left.(interface {
		Gt(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return eq.Gt(b.Right)
}

func evalGe(b ast.Binary) ast.Expression {
	eq, ok := b.Left.(interface {
		Ge(ast.Expression) ast.Expression
	})
	if !ok {
		return b
	}
	return eq.Ge(b.Right)
}

func evalAnd(b ast.Binary) ast.Expression {
	res := ast.IsTrue(b.Left) && ast.IsTrue(b.Right)
	return ast.CreateBoolean(b.Token, res)
}

func evalOr(b ast.Binary) ast.Expression {
	res := ast.IsTrue(b.Left) || ast.IsTrue(b.Right)
	return ast.CreateBoolean(b.Token, res)
}

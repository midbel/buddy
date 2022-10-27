package buddy

import (
	"fmt"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

type Callable interface {
	Call(*Resolver, ...types.Primitive) (types.Primitive, error)
	Arity() int
	Variadic() bool
	at(int) (string, error)
	index(string) (int, error)
}

type callBuiltin struct {
	builtin builtins.Builtin
}

func callableFromBuiltin(b builtins.Builtin) Callable {
	return callBuiltin{
		builtin: b,
	}
}

func (c callBuiltin) Arity() int {
	return len(c.builtin.Params)
}

func (c callBuiltin) Variadic() bool {
	return c.builtin.Variadic
}

func (c callBuiltin) Call(_ *Resolver, args ...types.Primitive) (types.Primitive, error) {
	return c.builtin.Run(args...)
}

func (c callBuiltin) at(i int) (string, error) {
	if i > len(c.builtin.Params) || i < 0 {
		return "", fmt.Errorf("index out of range")
	}
	return c.builtin.Params[i].Name, nil
}

func (c callBuiltin) index(ident string) (int, error) {
	x := slices.Index(c.builtin.Params, func(p builtins.Parameter) bool {
		return p.Name == ident
	})
	if x < 0 {
		return x, fmt.Errorf("%s: parameter not defined", ident)
	}
	return x, nil
}

type callExpr struct {
	fun function
	ctx Module
}

func callableFromExpression(e Expression) (Callable, error) {
	var (
		fun function
		mod Module
	)
	switch e := e.(type) {
	case function:
		fun = e
	case link:
		f, ok := e.Expression.(function)
		if !ok {
			return nil, fmt.Errorf("expression is not a function")	
		}
		fun = f
		mod = e.Module
	default:
		return nil, fmt.Errorf("expression is not a function")
	}
	return callExpr{
		fun: fun,
		ctx: mod,
	}, nil
}

func (c callExpr) Arity() int {
	return len(c.fun.params)
}

func (_ callExpr) Variadic() bool {
	return false
}

func (c callExpr) Call(res *Resolver, args ...types.Primitive) (types.Primitive, error) {
	if len(args) > len(c.fun.params) {
		return nil, fmt.Errorf("%s: invalid number of arguments given", c.fun.ident)
	}
	env := types.EmptyEnv()
	for i := range c.fun.params {
		var (
			p, _ = c.fun.params[i].(parameter)
			a    types.Primitive
		)
		if i < len(args) && args[i] != nil {
			a = args[i]
		} else {
			if p.expr == nil {
				return nil, fmt.Errorf("%s: parameter not set", p.ident)
			}
			v, err := eval(p.expr, res)
			if err != nil {
				return nil, err
			}
			a = v
		}
		env.Define(p.ident, a)
	}
	if c.ctx != nil {
		old := res.current
		defer func() {
			res.current = old
		}()
		res.current = c.ctx
	}
	return eval(c.fun.body, res.Sub(env))
}

func (c callExpr) at(i int) (string, error) {
	if i > len(c.fun.params) || i < 0 {
		return "", fmt.Errorf("index out of range")
	}
	e := c.fun.params[i]
	return e.(parameter).ident, nil
}

func (c callExpr) index(ident string) (int, error) {
	x := slices.Index(c.fun.params, func(e Expression) bool {
		return e.(parameter).ident == ident
	})
	if x < 0 {
		return x, fmt.Errorf("%s: parameter not defined", ident)
	}
	return x, nil
}

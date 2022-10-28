package buddy

import (
	"fmt"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

type Callable interface {
	Call(...types.Primitive) (types.Primitive, error)
	Arity() int
	Variadic() bool
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

func (c callBuiltin) Call(args ...types.Primitive) (types.Primitive, error) {
	return c.builtin.Run(args...)
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
	ctx *Resolver
}

func callableFromExpression(ctx *Resolver, expr Expression) (Callable, error) {
	fun, ok := expr.(function)
	if !ok {
		return nil, fmt.Errorf("expression is not a function")
	}
	return callExpr{
		fun: fun,
		ctx: ctx,
	}, nil
}

func (c callExpr) Arity() int {
	return len(c.fun.params)
}

func (_ callExpr) Variadic() bool {
	return false
}

func (c callExpr) Call(args ...types.Primitive) (types.Primitive, error) {
	if len(args) > len(c.fun.params) {
		return nil, fmt.Errorf("%s: invalid number of arguments given", c.fun.ident)
	}
	old := c.ctx.Environ
	defer func() {
		c.ctx.Environ = old
	}()
	c.ctx.Environ = types.EmptyEnv()
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
			v, err := eval(p.expr, c.ctx)
			if err != nil {
				return nil, err
			}
			a = v
		}
		c.ctx.Define(p.ident, a)
	}
	return eval(c.fun.body, c.ctx)
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

type Module interface {
	Lookup(string) (Callable, error)
}

type builtinModule struct {
	module builtins.Module
}

func moduleFromBuiltin(mod builtins.Module) Module {
	return builtinModule{
		module: mod,
	}
}

func (m builtinModule) Lookup(name string) (Callable, error) {
	b, err := m.module.Lookup(name)
	if err != nil {
		return nil, err
	}
	return callableFromBuiltin(b), nil
}

package eval

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/types"
)

type mutableModule interface {
	Get(string) (types.Module, error)
	Register(string, types.Module) error
}

type userModule struct {
	name      string
	callables map[string]types.Callable
	modules   map[string]types.Module
	*types.Environ
}

func emptyModule(ident string) *userModule {
	return &userModule{
		name:      ident,
		callables: make(map[string]types.Callable),
		modules:   make(map[string]types.Module),
		Environ:   types.EmptyEnv(),
	}
}

func (m *userModule) Id() string {
	return m.name
}

func (m *userModule) Get(ident string) (types.Module, error) {
	mod, ok := m.modules[ident]
	if !ok {
		return nil, fmt.Errorf("%s: module not found", ident)
	}
	return mod, nil
}

func (m *userModule) Register(ident string, mod types.Module) error {
	if _, ok := m.modules[ident]; ok {
		return fmt.Errorf("%s: module already imported", ident)
	}
	m.modules[ident] = mod
	return nil
}

func (m *userModule) Append(ident string, call types.Callable) error {
	if _, ok := m.callables[ident]; ok {
		return fmt.Errorf("%s: function already defined", ident)
	}
	m.callables[ident] = call
	return nil
}

func (m *userModule) Lookup(mod, ident string) (types.Callable, error) {
	if mod == "" {
		call, ok := m.callables[ident]
		if !ok {
			return nil, fmt.Errorf("%s: function not defined in %s", ident, m.name)
		}
		return call, nil
	}
	sub, ok := m.modules[mod]
	if !ok {
		return nil, fmt.Errorf("%s: module not found", mod)
	}
	return sub.Lookup("", ident)
}

type userCallable struct {
	fun ast.Function
}

func callableFromExpression(expr ast.Expression) (types.Callable, error) {
	fun, ok := expr.(ast.Function)
	if !ok {
		return nil, fmt.Errorf("expression is not a function definition")
	}
	call := userCallable{
		fun: fun,
	}
	return call, nil
}

func (c userCallable) Call(ctx types.Context, args ...types.Primitive) (types.Primitive, error) {
	i, ok := ctx.(*Interpreter)
	if !ok {
		return nil, fmt.Errorf("temporary hack")
	}
	old := i.Environ
	defer func() {
		i.Environ = old
	}()
	i.Environ = types.EnclosedEnv(old)
	for j := range c.fun.Params {
		var (
			par, _ = c.fun.Params[j].(ast.Parameter)
			val    types.Primitive
			err    error
		)
		if j < len(args) {
			val = args[j]
		} else if par.Expr != nil {
			val, err = eval(par.Expr, i)
		} else {
			_ = err
		}
		if err != nil {
			return nil, err
		}
		i.Define(par.Ident, val)
	}
	res, err := eval(c.fun.Body, i)
	if err != nil {
		err = fmt.Errorf("%s: %w", c.fun.Ident, err)
	}
	return res, err
}

func (c userCallable) Arity() int {
	return len(c.fun.Params)
}

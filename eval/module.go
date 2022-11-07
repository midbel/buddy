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

func (c userCallable) Call(ctx types.Context, args []types.Argument) (types.Primitive, error) {
	i, ok := ctx.(*Interpreter)
	if !ok {
		return nil, fmt.Errorf("temporary hack")
	}

	old := i.Environ
	defer func() {
		i.Environ = old
	}()
	i.Environ = types.EnclosedEnv(old)
	if err := c.setDefault(i); err != nil {
		return nil, err
	}
	var (
		ptr int
		set = make(map[string]struct{})
	)
	for ; ptr < len(args); ptr++ {
		if args[ptr].Name != "" {
			break
		}
		if ptr >= len(c.fun.Params) {
			return nil, fmt.Errorf("variadic argument not supported")
		}
		par := c.fun.Params[ptr].(ast.Parameter)
		if _, ok := set[par.Ident]; ok {
			return nil, fmt.Errorf("%s: argument already given", par.Ident)
		}
		set[par.Ident] = struct{}{}
		i.Define(par.Ident, args[ptr].Value)
	}
	for ; ptr < len(args); ptr++ {
		par, err := c.findParameter(args[ptr].Name)
		if err != nil {
			return nil, err
		}
		if _, ok := set[par.Ident]; ok {
			return nil, fmt.Errorf("%s: argument already given", par.Ident)
		}
		set[par.Ident] = struct{}{}
		i.Define(par.Ident, args[ptr].Value)
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

func (c userCallable) findParameter(ident string) (ast.Parameter, error) {
	for _, e := range c.fun.Params {
		p := e.(ast.Parameter)
		if p.Ident == ident {
			return p, nil
		}
	}
	return ast.Parameter{}, fmt.Errorf("%s: parameter not found", ident)
}

func (c userCallable) setDefault(i *Interpreter) error {
	for _, e := range c.fun.Params {
		p, ok := e.(ast.Parameter)
		if !ok || p.Expr == nil {
			continue
		}
		res, err := eval(p.Expr, i)
		if err != nil {
			return err
		}
		if err := i.Define(p.Ident, res); err != nil {
			return err
		}
	}
	return nil
}

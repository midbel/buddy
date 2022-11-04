package buddy

import (
	"errors"
	"fmt"
	"io"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/token"
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var (
	errBreak    = errors.New("break")
	errContinue = errors.New("continue")
	errReturn   = errors.New("return")
)

func Eval(r io.Reader) (types.Primitive, error) {
	return EvalEnv(r, types.EmptyEnv())
}

func EvalEnv(r io.Reader, env *types.Environ) (types.Primitive, error) {
	e, err := Parse(r)
	if err != nil {
		return nil, err
	}
	return Execute(e, env)
}

func Execute(expr Expression, env *types.Environ) (types.Primitive, error) {
	res := ResolveEnv(env)
	return execute(expr, res)
}

func execute(expr Expression, env *Resolver) (types.Primitive, error) {
	var err error
	if s, ok := expr.(script); ok {
		for k, e := range s.symbols {
			s.symbols[k], err = traverse(e, env, visitors)
			if err != nil {
				return nil, err
			}
		}
		env.symbols = s.symbols
	}
	if expr, err = traverse(expr, env, visitors); err != nil {
		return nil, err
	}
	return eval(expr, env)
}

func evalCall(c call, env *Resolver) (types.Primitive, error) {
	call, err := env.Lookup(c.ident)
	if err != nil {
		return nil, err
	}
	args, err := evalArguments(call, c.args, env)
	if err != nil {
		return nil, err
	}
	return call.Call(args...)
}

func evalPath(pat path, env *Resolver) (types.Primitive, error) {
	switch right := pat.right.(type) {
	case call:
		call, err := env.LookupCallable(right.ident, pat.ident)
		if err != nil {
			return nil, err
		}
		args, err := evalArguments(call, right.args, env)
		if err != nil {
			return nil, err
		}
		return call.Call(args...)
	case variable:
		res, err := env.Resolve(pat.ident)
		if err != nil {
			return nil, err
		}
		c, ok := res.(types.Container)
		if !ok {
			return nil, fmt.Errorf("%s is not a container", res)
		}
		return c.Get(types.CreateString(right.ident))
	case dict:
		return nil, nil
	default:
		return nil, fmt.Errorf("unexpected expression type")
	}
}

func evalArguments(call Callable, args []Expression, env *Resolver) ([]types.Primitive, error) {
	if len(args) > call.Arity() && !call.Variadic() {
		return nil, fmt.Errorf("too many arguments given")
	}
	var (
		err    error
		ptr    int
		values = make([]types.Primitive, len(args))
	)
	for ; ptr < len(args); ptr++ {
		e := args[ptr]
		if _, ok := e.(parameter); ok {
			break
		}
		values[ptr], err = eval(e, env)
		if err != nil {
			return nil, err
		}
	}
	for ; ptr < len(args); ptr++ {
		e, ok := args[ptr].(parameter)
		if !ok {
			return nil, fmt.Errorf("positional parameter should be given before named parameter")
		}
		i, err := call.index(e.ident)
		if err != nil {
			return nil, err
		}
		if values[i] != nil {
			return nil, fmt.Errorf("%s: parameter already set", e.ident)
		}
		values[i], err = eval(e.expr, env)
		if err != nil {
			return nil, err
		}
	}
	return values, nil
}

func evalImport(mod module, env *Resolver) (types.Primitive, error) {
	symbols := make(map[string]string)
	for _, s := range mod.symbols {
		if s.alias == "" {
			s.alias = s.ident
		}
		symbols[s.ident] = s.alias
	}
	if mod.alias == "" {
		mod.alias = slices.Lst(mod.ident)
	}
	err := env.Load(mod.ident, mod.alias, symbols)
	return nil, err
}

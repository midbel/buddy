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

func eval(expr Expression, env *Resolver) (types.Primitive, error) {
	var (
		res types.Primitive
		err error
	)
	switch e := expr.(type) {
	case script:
		res, err = evalScript(e, env)
	case call:
		res, err = evalCall(e, env)
		if errors.Is(err, errReturn) {
			err = nil
		}
		return res, err
	case array:
		return evalArray(e, env)
	case dict:
		return evalDict(e, env)
	case index:
		return evalIndex(e, env)
	case module:
		return evalImport(e, env)
	case path:
		return evalPath(e, env)
	case listcomp:
		return evalListcomp(e, env)
	case dictcomp:
		return evalDictcomp(e, env)
	case assert:
		res, err = evalAssert(e, env)
	case unary:
		res, err = evalUnary(e, env)
	case binary:
		res, err = evalBinary(e, env)
	case walrus:
		res, err = evalWalrus(e, env)
	case assign:
		res, err = evalAssign(e, env)
	case test:
		res, err = evalTest(e, env)
	case while:
		res, err = evalWhile(e, env)
	case forloop:
		res, err = evalFor(e, env)
	case foreach:
		res, err = evalForeach(e, env)
	}
	return res, err
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

func evalDictcomp(cmp dictcomp, env *Resolver) (types.Primitive, error) {
	var (
		arr      = types.CreateDict()
		evalItem func([]compitem) error
		old      = env.Environ
	)
	env.Environ = types.EnclosedEnv(old)
	defer func() {
		env.Environ = old
	}()

	evalItem = func(list []compitem) error {
		if len(list) == 0 {
			return nil
		}
		curr := slices.Fst(list)
		it, err := eval(curr.iter, env)
		if err != nil {
			return err
		}
		iter, ok := it.(types.Iterable)
		if !ok {
			return fmt.Errorf("can not iterate over %T", it)
		}
		err = iter.Iter(func(p types.Primitive) error {
			env.Define(curr.ident, p)
			for i := range curr.cdt {
				res, err := eval(curr.cdt[i], env)
				if err != nil {
					return err
				}
				if !res.True() {
					return nil
				}
			}
			if len(list) > 1 {
				return evalItem(slices.Rest(list))
			}
			key, err := eval(cmp.key, env)
			if err != nil {
				return err
			}
			val, err := eval(cmp.val, env)
			if err != nil {
				return err
			}
			arr.(types.Dict).Set(key, val)
			return nil
		})
		return err
	}
	if err := evalItem(cmp.list); err != nil {
		return nil, err
	}
	return arr, nil
}

func evalListcomp(cmp listcomp, env *Resolver) (types.Primitive, error) {
	var (
		arr      []types.Primitive
		evalItem func([]compitem) error
		old      = env.Environ
	)
	env.Environ = types.EnclosedEnv(old)
	defer func() {
		env.Environ = old
	}()

	evalItem = func(list []compitem) error {
		if len(list) == 0 {
			return nil
		}
		curr := slices.Fst(list)
		it, err := eval(curr.iter, env)
		if err != nil {
			return err
		}
		iter, ok := it.(types.Iterable)
		if !ok {
			return fmt.Errorf("can not iterate over %T", it)
		}
		err = iter.Iter(func(p types.Primitive) error {
			env.Define(curr.ident, p)
			for i := range curr.cdt {
				res, err := eval(curr.cdt[i], env)
				if err != nil {
					return err
				}
				if !res.True() {
					return nil
				}
			}
			if len(list) > 1 {
				return evalItem(slices.Rest(list))
			}
			res, err := eval(cmp.body, env)
			if err == nil {
				arr = append(arr, res)
			}
			return err
		})
		return err
	}
	if err := evalItem(cmp.list); err != nil {
		return nil, err
	}
	return types.CreateArray(arr), nil
}

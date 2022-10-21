package buddy

import (
	"errors"
	"fmt"
	"io"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/types"
)

var (
	errBreak    = errors.New("break")
	errContinue = errors.New("continue")
	errReturn   = errors.New("return")
)

func Eval(r io.Reader) (types.Primitive, error) {
	return EvalEnv(r, EmptyEnv())
}

func EvalEnv(r io.Reader, env *Environ) (types.Primitive, error) {
	e, err := Parse(r)
	if err != nil {
		return nil, err
	}
	return Execute(e, env)
}

func Execute(expr Expression, env *Environ) (types.Primitive, error) {
	resolv := ResolveEnv(env)
	if s, ok := expr.(script); ok {
		resolv.symbols = s.symbols
	}
	return execute(expr, resolv)
}

func execute(expr Expression, env *Resolver) (types.Primitive, error) {
	var (
		err  error
		list = []visitFunc{
			trackVariables,
			replaceFunctionArgs,
			inlineFunctionCall,
			replaceValue,
		}
	)
	if expr, err = traverse(expr, env, list); err != nil {
		return nil, err
	}
	for k, e := range env.symbols {
		env.symbols[k], err = traverse(e, env, list)
		if err != nil {
			return nil, err
		}
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
	case literal:
		return types.CreateString(e.str), nil
	case number:
		return types.CreateFloat(e.value), nil
	case boolean:
		return types.CreateBool(e.value), nil
	case variable:
		return env.Resolve(e.ident)
	case unary:
		res, err = evalUnary(e, env)
	case binary:
		res, err = evalBinary(e, env)
	case assign:
		res, err = evalAssign(e, env)
	case test:
		res, err = evalTest(e, env)
	case while:
		res, err = evalWhile(e, env)
	case returned:
		if e.right == nil {
			return nil, errReturn
		}
		res, err = eval(e.right, env)
		return res, errReturn
	case breaked:
		return nil, errBreak
	case continued:
		return nil, errContinue
	}
	return res, err
}

func evalScript(s script, env *Resolver) (types.Primitive, error) {
	var (
		res types.Primitive
		err error
	)
	for _, e := range s.list {
		res, err = eval(e, env)
		if err != nil && !errors.Is(err, errReturn) {
			if builtins.IsExit(err) {
				return res, err
			}
			break
		}
		if errors.Is(err, errReturn) {
			return res, nil
		}
	}
	return res, err
}

func evalUnary(u unary, env *Resolver) (types.Primitive, error) {
	res, err := eval(u.right, env)
	if err != nil {
		return nil, err
	}
	switch u.op {
	case Not:
		return res.Not()
	case Sub:
		return res.Rev()
	default:
		return nil, fmt.Errorf("unsupported unary operator")
	}
}

func evalBinary(b binary, env *Resolver) (types.Primitive, error) {
	left, err := eval(b.left, env)
	if err != nil {
		return nil, err
	}
	right, err := eval(b.right, env)
	if err != nil {
		return nil, err
	}
	switch b.op {
	default:
		return nil, fmt.Errorf("unsupported binary operator")
	case Add:
		return left.Add(right)
	case Sub:
		return left.Sub(right)
	case Mul:
		return left.Mul(right)
	case Div:
		return left.Div(right)
	case Pow:
		return left.Pow(right)
	case Mod:
		return left.Mod(right)
	case And:
		return types.And(left, right)
	case Or:
		return types.Or(left, right)
	case Eq:
		return left.Eq(right)
	case Ne:
		return left.Ne(right)
	case Lt:
		return left.Lt(right)
	case Le:
		return left.Le(right)
	case Gt:
		return left.Gt(right)
	case Ge:
		return left.Ge(right)
	}
}

func evalTest(t test, env *Resolver) (types.Primitive, error) {
	res, err := eval(t.cdt, env)
	if err != nil {
		return nil, err
	}
	if res.True() {
		return eval(t.csq, env)
	}
	if t.alt == nil {
		return nil, nil
	}
	return eval(t.alt, env)
}

func evalWhile(w while, env *Resolver) (types.Primitive, error) {
	var (
		res types.Primitive
		err error
	)
	for {
		res, err = eval(w.cdt, env)
		if err != nil {
			return nil, err
		}
		if !res.True() {
			break
		}
		res, err = eval(w.body, env)
		if err != nil {
			if errors.Is(err, errBreak) {
				break
			}
			if errors.Is(err, errContinue) {
				continue
			}
			return nil, err
		}
	}
	return res, nil
}

func evalAssign(a assign, env *Resolver) (types.Primitive, error) {
	res, err := eval(a.right, env)
	if err != nil {
		return nil, err
	}
	env.Define(a.ident, res)
	return nil, nil
}

func evalCall(c call, env *Resolver) (types.Primitive, error) {
	var (
		args []types.Primitive
		res  types.Primitive
		err  error
	)
	for _, a := range c.args {
		res, err = eval(a, env)
		if err != nil {
			return nil, err
		}
		args = append(args, res)
	}
	call, err := env.Lookup(c.ident)
	if err != nil {
		return nil, err
	}
	return call.Call(env, args...)
}

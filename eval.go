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
	if s, ok := expr.(script); ok {
		var err error
		for k, e := range s.symbols {
			s.symbols[k], err = traverse(e, res, visitors)
			if err != nil {
				return nil, err
			}
		}
		res.symbols = s.symbols
	}
	return execute(expr, res)
}

func execute(expr Expression, env *Resolver) (types.Primitive, error) {
	var err error
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
		return evalModule(e, env)
	case path:
		return evalPath(e, env)
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
	case foreach:
		res, err = evalForeach(e, env)
	case returned:
		if e.right == nil {
			return nil, errReturn
		}
		res, err = eval(e.right, env)
		if err != nil {
			return nil, err
		}
		return res, errReturn
	case breaked:
		return nil, errBreak
	case continued:
		return nil, errContinue
	default:
		return nil, fmt.Errorf("unsupported node type %T", expr)
	}
	return res, err
}

func evalScript(s script, env *Resolver) (types.Primitive, error) {
	var (
		res types.Primitive
		err error
	)
	for i := range s.list {
		res, err = eval(s.list[i], env)
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
		cal, ok := res.(types.Calculable)
		if !ok {
			return nil, fmt.Errorf("%w: value can not be reversed", types.ErrOperation)
		}
		return cal.Rev()
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
	var do binaryFunc
	if isArithmetic(b.op) {
		do = doArithmetic
	} else if isComparison(b.op) {
		do = doComparison
	} else if isRelational(b.op) {
		do = doRelational
	} else {
		return nil, fmt.Errorf("unsupported binary operator")
	}
	return do(left, right, b.op)
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

func evalForeach(f foreach, env *Resolver) (types.Primitive, error) {
	it, err := eval(f.iter, env)
	if err != nil {
		return nil, err
	}
	iter, ok := it.(types.Iterable)
	if !ok {
		return nil, fmt.Errorf("can not iterate on %T", it)
	}
	var (
		old = env.Environ
		res types.Primitive
	)
	defer func() {
		env.Environ = old
	}()
	env.Environ = types.EnclosedEnv(old)
	err = iter.Iter(func(p types.Primitive) error {
		env.Define(f.ident, p)
		res, err = eval(f.body, env)
		return err
	})
	return res, err
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
	switch a := a.ident.(type) {
	case variable:
		env.Define(a.ident, res)
	case index:
		return assignIndex(a, res, env)
	default:
		return nil, fmt.Errorf("can not assign to %T", a)
	}
	return res, nil
}

func evalCall(c call, env *Resolver) (types.Primitive, error) {
	call, err := env.Lookup(c.ident)
	if err != nil {
		return nil, err
	}

	if len(c.args) > call.Arity() && !call.Variadic() {
		return nil, fmt.Errorf("too many arguments given")
	}
	var (
		ptr  int
		args = make([]types.Primitive, len(c.args))
	)
	for ; ptr < len(c.args); ptr++ {
		e := c.args[ptr]
		if _, ok := e.(parameter); ok {
			break
		}
		args[ptr], err = eval(e, env)
		if err != nil {
			return nil, err
		}
	}
	for ; ptr < len(c.args); ptr++ {
		e, ok := c.args[ptr].(parameter)
		if !ok {
			return nil, fmt.Errorf("positional parameter should be given before named parameter")
		}
		i, err := call.index(e.ident)
		if err != nil {
			return nil, err
		}
		if args[i] != nil {
			return nil, fmt.Errorf("%s: parameter already set", e.ident)
		}
		args[i], err = eval(e.expr, env)
		if err != nil {
			return nil, err
		}
	}
	return call.Call(env, args...)
}

func evalPath(pat path, env *Resolver) (types.Primitive, error) {
	return nil, fmt.Errorf("not yet implemented")
}

func evalModule(mod module, env *Resolver) (types.Primitive, error) {
	if mod.alias == "" {
		mod.alias = slices.Lst(mod.ident)
	}
	err := env.Load(mod.ident, mod.alias)
	return nil, err
}

func evalArray(arr array, env *Resolver) (types.Primitive, error) {
	var list []types.Primitive
	for i := range arr.list {
		v, err := eval(arr.list[i], env)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return types.CreateArray(list), nil
}

func evalDict(arr dict, env *Resolver) (types.Primitive, error) {
	d := types.CreateDict()
	for k, v := range arr.list {
		kp, err := eval(k, env)
		if err != nil {
			return nil, err
		}
		vp, err := eval(v, env)
		if err != nil {
			return nil, err
		}
		d, _ = d.(types.Container).Set(kp, vp)
	}
	return d, nil
}

func evalIndex(idx index, env *Resolver) (types.Primitive, error) {
	p, err := eval(idx.arr, env)
	if err != nil {
		return nil, err
	}
	c, ok := p.(types.Container)
	if !ok {
		return nil, fmt.Errorf("%T is not a container!", p)
	}
	var res types.Primitive
	for j, e := range idx.list {
		ix, err := eval(e, env)
		if err != nil {
			return nil, err
		}
		res, err = c.Get(ix)
		if err != nil {
			return nil, err
		}
		c, ok = res.(types.Container)
		if !ok && j == len(idx.list)-1 {
			return nil, fmt.Errorf("%T is not a container")
		}
	}
	return res, nil
}

func assignIndex(idx index, value types.Primitive, env *Resolver) (types.Primitive, error) {
	ix, err := eval(idx.list[0], env)
	if err != nil {
		return nil, err
	}
	switch i := idx.arr.(type) {
	case variable:
		v, err := env.Resolve(i.ident)
		if err != nil {
			return nil, err
		}
		c, ok := v.(types.Container)
		if !ok {
			return nil, fmt.Errorf("%s is not a container!", i.ident)
		}
		v, err = c.Set(ix, value)
		if err == nil {
			err = env.Define(i.ident, v)
		}
		return v, err
	case array:
		arr, err := evalArray(i, env)
		if err != nil {
			return nil, err
		}
		c, ok := arr.(types.Container)
		if !ok {
			return nil, fmt.Errorf("value is not a container!")
		}
		return c.Set(ix, value)
	case dict:
	default:
		return nil, fmt.Errorf("can not assign to %T", idx.arr)
	}
	return nil, nil
}

type binaryFunc func(types.Primitive, types.Primitive, rune) (types.Primitive, error)

func doArithmetic(left, right types.Primitive, op rune) (types.Primitive, error) {
	cal, ok := left.(types.Calculable)
	if !ok {
		return nil, fmt.Errorf("%w: values can not be calculated", types.ErrOperation)
	}
	var err error
	switch op {
	case Add:
		left, err = cal.Add(right)
	case Sub:
		left, err = cal.Sub(right)
	case Mul:
		left, err = cal.Mul(right)
	case Div:
		left, err = cal.Div(right)
	case Pow:
		left, err = cal.Pow(right)
	case Mod:
		left, err = cal.Mod(right)
	default:
		err = fmt.Errorf("unsupported binary operator")
	}
	return left, err
}

func doComparison(left, right types.Primitive, op rune) (types.Primitive, error) {
	cmp, ok := left.(types.Comparable)
	if !ok {
		return nil, fmt.Errorf("%w: values can not be compared", types.ErrOperation)
	}
	var err error
	switch op {
	case Eq:
		left, err = cmp.Eq(right)
	case Ne:
		left, err = cmp.Ne(right)
	case Lt:
		left, err = cmp.Lt(right)
	case Le:
		left, err = cmp.Le(right)
	case Gt:
		left, err = cmp.Gt(right)
	case Ge:
		left, err = cmp.Ge(right)
	default:
		err = fmt.Errorf("unsupported binary operator")
	}
	return left, err
}

func doRelational(left, right types.Primitive, op rune) (types.Primitive, error) {
	if op == And {
		return types.And(left, right)
	}
	return types.Or(left, right)
}

func isRelational(op rune) bool {
	return op == And || op == Or
}

func isArithmetic(op rune) bool {
	switch op {
	case Add, Sub, Div, Mul, Pow, Mod:
		return true
	default:
		return false
	}
}

func isComparison(op rune) bool {
	switch op {
	case Eq, Ne, Lt, Le, Gt, Ge:
		return true
	default:
		return false
	}
}

package eval

import (
	"errors"
	"fmt"
	"io"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/parse"
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var (
	errBreak       = errors.New("break")
	errContinue    = errors.New("continue")
	errReturn      = errors.New("return")
	errEval        = errors.New("expression can not be evaluated in the current context")
	errImplemented = errors.New("not yet implemented")
)

func Eval(r io.Reader) (types.Primitive, error) {
	return EvalEnv(r, types.EmptyEnv())
}

func EvalEnv(r io.Reader, env *types.Environ) (types.Primitive, error) {
	expr, err := parse.New(r).Parse()
	if err != nil {
		return nil, err
	}
	bud := New(env)
	if s, ok := expr.(ast.Script); ok {
		mod, ok := bud.stack.Top().(*userModule)
		if !ok {
			return nil, fmt.Errorf("fail to initialize main module")
		}
		for k, expr := range s.Symbols {
			call, err := callableFromExpression(expr)
			if err != nil {
				return nil, err
			}
			if err := mod.Append(k, call); err != nil {
				return nil, err
			}
		}
	}
	return eval(expr, bud)
}

func eval(expr ast.Expression, env *Interpreter) (types.Primitive, error) {
	var (
		res types.Primitive
		err error
	)
	switch e := expr.(type) {
	case ast.Literal:
		res = types.CreateString(e.Str)
	case ast.Double:
		res = types.CreateFloat(e.Value)
	case ast.Integer:
		res = types.CreateInt(e.Value)
	case ast.Boolean:
		res = types.CreateBool(e.Value)
	case ast.Variable:
		res, err = env.Resolve(e.Ident)
	case ast.Array:
		res, err = evalArray(e, env)
	case ast.Dict:
		res, err = evalDict(e, env)
	case ast.Index:
		res, err = evalIndex(e, env)
	case ast.Path:
		res, err = evalPath(e, env)
	case ast.Call:
		res, err = evalCall(e, env)
	case ast.Parameter:
		res, err = eval(e.Expr, env)
	case ast.Assert:
		res, err = evalAssert(e, env)
	case ast.Assign:
		res, err = evalAssign(e, env)
	case ast.Unary:
		res, err = evalUnary(e, env)
	case ast.Binary:
		res, err = evalBinary(e, env)
	case ast.ListComp:
		res, err = evalListComp(e, env)
	case ast.DictComp:
		res, err = evalDictComp(e, env)
	case ast.Test:
		res, err = evalTest(e, env)
	case ast.While:
		res, err = evalWhile(e, env)
	case ast.For:
		res, err = evalFor(e, env)
	case ast.ForEach:
		res, err = evalForeach(e, env)
	case ast.Import:
		res, err = evalImport(e, env)
	case ast.Script:
		res, err = evalScript(e, env)
	case ast.Return:
		res, err = evalReturn(e, env)
	case ast.Break:
		return nil, errBreak
	case ast.Continue:
		return nil, errContinue
	default:
		return nil, fmt.Errorf("eval: %w", errEval)
	}
	return res, err
}

func evalArray(a ast.Array, env *Interpreter) (types.Primitive, error) {
	var list []types.Primitive
	for i := range a.List {
		v, err := eval(a.List[i], env)
		if err != nil {
			return nil, err
		}
		list = append(list, v)
	}
	return types.CreateArray(list), nil
}

func evalDict(a ast.Dict, env *Interpreter) (types.Primitive, error) {
	d := types.CreateDict()
	for k, v := range a.List {
		kp, err := eval(k, env)
		if err != nil {
			return nil, err
		}
		vp, err := eval(v, env)
		if err != nil {
			return nil, err
		}
		d, err = d.(types.Container).Set(kp, vp)
		if err != nil {
			return nil, err
		}
	}
	return d, nil
}

func evalIndex(i ast.Index, env *Interpreter) (types.Primitive, error) {
	p, err := eval(i.Arr, env)
	if err != nil {
		return nil, err
	}
	c, ok := p.(types.Container)
	if !ok {
		return nil, types.ContainerError(p)
	}
	var res types.Primitive
	for _, e := range slices.Slice(i.List) {
		ix, err := eval(e, env)
		if err != nil {
			return nil, err
		}
		if res, err = c.Get(ix); err != nil {
			return nil, err
		}
		c, ok = res.(types.Container)
		if !ok {
			return nil, types.ContainerError(res)
		}
	}
	res, err = eval(slices.Lst(i.List), env)
	if err != nil {
		return nil, err
	}
	return c.Get(res)
}

func evalPath(p ast.Path, env *Interpreter) (types.Primitive, error) {
	switch right := p.Right.(type) {
	case ast.Call:
		args, err := evalArguments(right, env)
		if err != nil {
			return nil, err
		}
		return env.Call(p.Ident, right.Ident, func(call types.Callable) (types.Primitive, error) {
			return call.Call(env, args)
		})
	case ast.Variable:
		res, err := env.Resolve(p.Ident)
		if err != nil {
			return nil, err
		}
		c, ok := res.(types.Container)
		if !ok {
			return nil, types.ContainerError(res)
		}
		return c.Get(types.CreateString(right.Ident))
	default:
		return nil, fmt.Errorf("path: %w", errEval)
	}
}

func evalCall(c ast.Call, env *Interpreter) (types.Primitive, error) {
	args, err := evalArguments(c, env)
	if err != nil {
		return nil, err
	}
	return env.Call("", c.Ident, func(call types.Callable) (types.Primitive, error) {
		return call.Call(env, args)
	})
}

func evalArguments(c ast.Call, env *Interpreter) ([]types.Argument, error) {
	var (
		ptr  int
		args = make([]types.Argument, len(c.Args))
	)
	for ; ptr < len(c.Args); ptr++ {
		if _, ok := c.Args[ptr].(ast.Parameter); ok {
			break
		}
		tmp, err := eval(c.Args[ptr], env)
		if err != nil {
			return nil, err
		}
		args[ptr] = types.NamedArg("", ptr, tmp)
	}
	for ; ptr < len(c.Args); ptr++ {
		tmp, err := eval(c.Args[ptr], env)
		if err != nil {
			return nil, err
		}
		args[ptr] = types.NamedArg("", ptr, tmp)
		if p, ok := c.Args[ptr].(ast.Parameter); !ok {
			return nil, fmt.Errorf("expected named argument")
		} else {
			args[ptr].Name = p.Ident
		}
	}
	return args, nil
}

func evalAssert(a ast.Assert, env *Interpreter) (types.Primitive, error) {
	res, err := eval(a.Expr, env)
	if err != nil {
		return nil, err
	}
	if !res.True() {
		return nil, types.ErrAssert
	}
	return res, nil
}

func evalAssign(a ast.Assign, env *Interpreter) (types.Primitive, error) {
	res, err := eval(a.Right, env)
	if err != nil {
		return nil, err
	}
	switch a := a.Ident.(type) {
	case ast.Variable:
		env.Define(a.Ident, res)
	case ast.Index:
		err = assignIndex(a, res, env)
	default:
		return nil, fmt.Errorf("assignment: %w", errEval)
	}
	return res, err
}

func evalUnary(u ast.Unary, env *Interpreter) (types.Primitive, error) {
	res, err := eval(u.Right, env)
	if err != nil {
		return nil, err
	}
	return executeUnary(u.Op, res)
}

func evalBinary(b ast.Binary, env *Interpreter) (types.Primitive, error) {
	left, err := eval(b.Left, env)
	if err != nil {
		return nil, err
	}
	right, err := eval(b.Right, env)
	if err != nil {
		return nil, err
	}
	return executeBinary(b.Op, left, right)
}

func evalListComp(lc ast.ListComp, env *Interpreter) (types.Primitive, error) {
	var (
		arr []types.Primitive
		err error
		old = env.Environ
	)
	defer func() {
		env.Environ = old
	}()
	env.Environ = types.EnclosedEnv(old)

	err = evalCompItem(lc.List, env, func() error {
		res, err := eval(lc.Body, env)
		if err == nil {
			arr = append(arr, res)
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return types.CreateArray(arr), nil
}

func evalDictComp(dc ast.DictComp, env *Interpreter) (types.Primitive, error) {
	var (
		dict = types.CreateDict()
		err  error
		old  = env.Environ
	)
	defer func() {
		env.Environ = old
	}()
	env.Environ = types.EnclosedEnv(old)

	err = evalCompItem(dc.List, env, func() error {
		key, err := eval(dc.Key, env)
		if err != nil {
			return err
		}
		val, err := eval(dc.Val, env)
		if err != nil {
			return err
		}
		dict.(types.Dict).Set(key, val)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dict, nil
}

func evalCompItem(cis []ast.CompItem, env *Interpreter, do func() error) error {
	if len(cis) == 0 {
		return nil
	}
	curr := slices.Fst(cis)
	it, err := eval(curr.Iter, env)
	if err != nil {
		return err
	}
	iter, ok := it.(types.Iterable)
	if !ok {
		return types.IterationError(it)
	}
	return iter.Iter(func(p types.Primitive) error {
		env.Define(curr.Ident, p)
		for i := range curr.Cdt {
			res, err := eval(curr.Cdt[i], env)
			if err != nil {
				return err
			}
			if !res.True() {
				return nil
			}
		}
		if len(cis) > 1 {
			old := env.Environ
			defer func() {
				env.Environ = old
			}()
			env.Environ = types.EnclosedEnv(old)
			return evalCompItem(slices.Rest(cis), env, do)
		}
		return do()
	})
}

func evalTest(t ast.Test, env *Interpreter) (types.Primitive, error) {
	res, err := eval(t.Cdt, env)
	if err != nil {
		return nil, err
	}
	old := env.Environ
	defer func() {
		env.Environ = old
	}()
	env.Environ = types.EnclosedEnv(old)
	if res.True() {
		return eval(t.Csq, env)
	}
	if t.Alt == nil {
		return nil, nil
	}
	return eval(t.Alt, env)
}

func evalWhile(w ast.While, env *Interpreter) (types.Primitive, error) {
	return evalLoop(w.Body, w.Cdt, nil, env)
}

func evalFor(f ast.For, env *Interpreter) (types.Primitive, error) {
	old := env.Environ
	defer func() {
		env.Environ = old
	}()
	env.Environ = types.EnclosedEnv(old)
	if f.Init != nil {
		_, err := eval(f.Init, env)
		if err != nil {
			return nil, err
		}
	}
	return evalLoop(f.Body, f.Cdt, f.Incr, env)
}

func evalLoop(body, cdt, incr ast.Expression, env *Interpreter) (types.Primitive, error) {
	execBody := func() (types.Primitive, error) {
		old := env.Environ
		defer func() {
			env.Environ = old
		}()
		env.Environ = types.EnclosedEnv(old)
		return eval(body, env)
	}
	var (
		res types.Primitive
		err error
	)
	for {
		tmp, err1 := eval(cdt, env)
		if err1 != nil {
			return nil, err1
		}
		if !tmp.True() {
			break
		}
		res, err = execBody()
		if err != nil {
			if errors.Is(err, errBreak) {
				break
			}
			if errors.Is(err, errContinue) {
				continue
			}
			return nil, err
		}
		if incr != nil {
			_, err := eval(incr, env)
			if err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}

func evalForeach(f ast.ForEach, env *Interpreter) (types.Primitive, error) {
	it, err := eval(f.Iter, env)
	if err != nil {
		return nil, err
	}
	iter, ok := it.(types.Iterable)
	if !ok {
		return nil, types.IterationError(it)
	}
	var res types.Primitive
	err = iter.Iter(func(p types.Primitive) error {
		old := env.Environ
		defer func() {
			env.Environ = old
		}()
		env.Environ = types.EnclosedEnv(old)
		env.Define(f.Ident, p)
		res, err = eval(f.Body, env)
		return err
	})
	return res, err
}

func evalImport(i ast.Import, env *Interpreter) (types.Primitive, error) {
	return nil, env.Load(i.Ident, i.Alias)
}

func evalScript(s ast.Script, env *Interpreter) (types.Primitive, error) {
	var (
		res types.Primitive
		err error
	)
	for i := range s.List {
		res, err = eval(s.List[i], env)
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

func evalReturn(ret ast.Return, env *Interpreter) (types.Primitive, error) {
	var (
		res types.Primitive
		err error
	)
	if ret.Right != nil {
		res, err = eval(ret.Right, env)
		if err == nil {
			err = errReturn
		}
	}
	return res, err
}

func assignIndex(i ast.Index, value types.Primitive, env *Interpreter) error {
	var (
		res types.Primitive
		err error
	)
	switch a := i.Arr.(type) {
	case ast.Variable:
		res, err = env.Resolve(a.Ident)
	case ast.Array:
		res, err = eval(a, env)
	case ast.Dict:
		res, err = eval(a, env)
	default:
		return fmt.Errorf("index assignment: %w", errEval)
	}
	if err != nil {
		return err
	}
	c, ok := res.(types.Container)
	if !ok {
		return types.ContainerError(res)
	}
	for _, e := range slices.Slice(i.List) {
		x, err := eval(e, env)
		if err != nil {
			return err
		}
		if res, err = c.Get(x); err != nil {
			return err
		}
		c, ok = res.(types.Container)
		if !ok {
			return types.ContainerError(res)
		}
	}
	if res, err = eval(slices.Lst(i.List), env); err != nil {
		return err
	}
	_, err = c.Set(res, value)
	return err
}

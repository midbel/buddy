package buddy

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/midbel/buddy/builtins"
)

func Eval(r io.Reader) (interface{}, error) {
	expr, err := Parse(r)
	if err != nil {
		return nil, err
	}
	env := EmptyEnv[any]()
	return Execute(expr, env)
}

func Execute(expr Expression, env *Environ[any]) (interface{}, error) {
	var (
		err  error
		list = []visitFunc{replaceValue}
	)
	if expr, err = traverse(expr, env, list); err != nil {
		return nil, err
	}
	return eval(expr, env)
}

func eval(expr Expression, env *Environ[any]) (interface{}, error) {
	var (
		res interface{}
		err error
	)
	switch e := expr.(type) {
	case script:
		for _, e := range e.list {
			res, err = eval(e, env)
			if err != nil {
				break
			}
		}
	case call:
		res, err = evalCall(e, env)
	case literal:
		return e.str, nil
	case number:
		return e.value, nil
	case boolean:
		return e.value, nil
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
	}
	return res, err
}

func evalUnary(u unary, env *Environ[any]) (interface{}, error) {
	res, err := eval(u.right, env)
	if err != nil {
		return nil, err
	}
	switch u.op {
	case Not:
		return !isTruthy(res), nil
	case Sub:
		f, ok := res.(float64)
		if !ok {
			return nil, fmt.Errorf("expected float")
		}
		return -f, nil
	default:
		return nil, fmt.Errorf("unsupported unary operator")
	}
}

func evalBinary(b binary, env *Environ[any]) (interface{}, error) {
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
		return execAdd(left, right)
	case Sub:
		return execSub(left, right)
	case Mul:
		return execMul(left, right)
	case Div:
		return execDiv(left, right)
	case Pow:
		return execPow(left, right)
	case Mod:
		return execMod(left, right)
	case And:
		return execAnd(left, right)
	case Or:
		return execOr(left, right)
	case Eq:
		return execEqual(left, right, false)
	case Ne:
		return execEqual(left, right, true)
	case Lt:
		return execLesser(left, right, false)
	case Le:
		return execLesser(left, right, true)
	case Gt:
		return execGreater(left, right, false)
	case Ge:
		return execGreater(left, right, true)
	}
}

func evalTest(t test, env *Environ[any]) (interface{}, error) {
	res, err := eval(t.cdt, env)
	if err != nil {
		return nil, err
	}
	if isTruthy(res) {
		return eval(t.csq, env)
	}
	if t.alt == nil {
		return nil, nil
	}
	return eval(t.alt, env)
}

func evalWhile(w while, env *Environ[any]) (interface{}, error) {
	var (
		res interface{}
		err error
	)
	for {
		res, err = eval(w.cdt, env)
		if err != nil {
			return nil, err
		}
		if !isTruthy(res) {
			break
		}
		res, err = eval(w.body, env)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func evalAssign(a assign, env *Environ[any]) (interface{}, error) {
	res, err := eval(a.right, env)
	if err != nil {
		return nil, err
	}
	env.Define(a.ident, res)
	return nil, nil
}

func evalCall(c call, env *Environ[any]) (interface{}, error) {
	var (
		args []interface{}
		res  interface{}
		err  error
	)
	for _, a := range c.args {
		res, err = eval(a, env)
		if err != nil {
			return nil, err
		}
		args = append(args, res)
	}
	fn, ok := builtins.Builtins[c.ident]
	if !ok {
		return nil, fmt.Errorf("%s: function not defined")
	}
	return fn(args...)
}

func execLesser(left, right interface{}, eq bool) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		y, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isLesser(x, y, eq), nil
	case string:
		y, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isLesser(x, y, eq), nil
	default:
		return nil, fmt.Errorf("type can not be compared")
	}
}

func execGreater(left, right interface{}, eq bool) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		y, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isGreater(x, y, eq), nil
	case string:
		y, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type for comparison")
		}
		return isGreater(x, y, eq), nil
	default:
		return nil, fmt.Errorf("type can not be compared")
	}
}

func execEqual(left, right interface{}, ne bool) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		y, ok := right.(float64)
		if !ok {
			return nil, fmt.Errorf("incompatible type for equality")
		}
		return isEqual(x, y, ne), nil
	case string:
		y, ok := right.(string)
		if !ok {
			return nil, fmt.Errorf("incompatible type for equality")
		}
		return isEqual(x, y, ne), nil
	case bool:
		y, ok := right.(bool)
		if !ok {
			return nil, fmt.Errorf("incompatible type for equality")
		}
		return isEqual(x, y, ne), nil
	default:
		return nil, fmt.Errorf("type can not be compared")
	}
	return nil, nil
}

func isEqual[T float64 | string | bool](left, right T, ne bool) bool {
	ok := left == right
	if ne {
		ok = !ok
	}
	return ok
}

func isLesser[T float64 | string](left, right T, eq bool) bool {
	ok := left < right
	if !ok && eq {
		ok = left == right
	}
	return ok
}

func isGreater[T float64 | string](left, right T, eq bool) bool {
	ok := left > right
	if !ok && eq {
		ok = left == right
	}
	return ok
}

func execAnd(left, right interface{}) (interface{}, error) {
	return isTruthy(left) && isTruthy(right), nil
}
func execOr(left, right interface{}) (interface{}, error) {
	return isTruthy(left) || isTruthy(right), nil
}

func execAdd(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			return x + y, nil
		}
		if y, ok := right.(string); ok {
			return fmt.Sprintf("%f%s", x, y), nil
		}
		return nil, fmt.Errorf("incompatible type for addition")
	case string:
		if y, ok := right.(float64); ok {
			return fmt.Sprintf("%s%f", x, y), nil
		}
		if y, ok := right.(string); ok {
			return x + y, nil
		}
		return nil, fmt.Errorf("incompatible type for addition")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execSub(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			return x - y, nil
		}
		return nil, fmt.Errorf("incompatible type for subtraction")
	case string:
		if y, ok := right.(float64); ok {
			if y < 0 && int(math.Abs(y)) < len(x) {
				y = math.Abs(y)
				return x[int(y):], nil
			}
			if y > 0 && int(y) < len(x) {
				return x[:int(y)], nil
			}
		}
		return nil, fmt.Errorf("incompatible type for subtraction")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execMul(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			return x * y, nil
		}
		if y, ok := right.(string); ok {
			return strings.Repeat(y, int(x)), nil
		}
		return nil, fmt.Errorf("incompatible type for multiply")
	case string:
		if y, ok := right.(float64); ok {
			return strings.Repeat(x, int(y)), nil
		}
		return nil, fmt.Errorf("incompatible type for multiply")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execDiv(left, right interface{}) (interface{}, error) {
	switch x := left.(type) {
	case float64:
		if y, ok := right.(float64); ok {
			if y < 0 {
				return nil, fmt.Errorf("division by zero")
			}
			return x / y, nil
		}
		return nil, fmt.Errorf("incompatible type for division")
	case string:
		if y, ok := right.(float64); ok && y > 0 {
			z := len(x) / int(y)
			return x[:z], nil
		}
		return nil, fmt.Errorf("incompatible type for division")
	default:
		return nil, fmt.Errorf("left value should be literal or number")
	}
}

func execMod(left, right interface{}) (interface{}, error) {
	x, ok1 := left.(float64)
	y, ok2 := right.(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("incompatible type for modulo")
	}
	if y == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return math.Mod(x, y), nil
}

func execPow(left, right interface{}) (interface{}, error) {
	x, ok1 := left.(float64)
	y, ok2 := right.(float64)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("incompatible type for power")
	}
	return math.Pow(x, y), nil
}

func isTruthy(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		return x != ""
	default:
		return v != nil
	}
}

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
	at(int) (string, error)
	index(string) (int, error)
}

type callBuiltin struct {
	builtin builtins.Builtin
}

func makeCallFromBuiltin(b builtins.Builtin) Callable {
	return callBuiltin{
		builtin: b,
	}
}

func (c callBuiltin) Arity() int {
	return len(c.builtin.Params)
}

func (c callBuiltin) Call(res *Resolver, args ...types.Primitive) (types.Primitive, error) {
	if err := res.enter(); err != nil {
		return nil, err
	}
	defer res.leave()
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
}

func makeCallFromExpr(e Expression) (Callable, error) {
	f, ok := e.(function)
	if !ok {
		return nil, fmt.Errorf("expression is not a function")
	}
	return callExpr{
		fun: f,
	}, nil
}

func (c callExpr) Arity() int {
	return len(c.fun.params)
}

func (c callExpr) Call(res *Resolver, args ...types.Primitive) (types.Primitive, error) {
	if len(args) > len(c.fun.params) {
		return nil, fmt.Errorf("%s: invalid number of arguments given", c.fun.ident)
	}
	env := EmptyEnv()
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
	if err := res.enter(); err != nil {
		return nil, err
	}
	old := res.Environ
	defer func() {
		res.Environ = old
		res.leave()
	}()
	res.Environ = env
	return eval(c.fun.body, res)
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

type Expression interface {
	isPrimitive() bool
}

func createPrimitive(res interface{}) (Expression, error) {
	switch r := res.(type) {
	case int64:
		return createNumber(float64(r)), nil
	case float64:
		return createNumber(r), nil
	case bool:
		return createBoolean(r), nil
	case string:
		return createLiteral(r), nil
	default:
		return nil, fmt.Errorf("unexpected primitive type: %T", res)
	}
}

type variable struct {
	ident string
}

func createVariable(ident string) variable {
	return variable{
		ident: ident,
	}
}

func (_ variable) isPrimitive() bool {
	return false
}

type literal struct {
	str string
}

func createLiteral(str string) literal {
	return literal{
		str: str,
	}
}

func (_ literal) isPrimitive() bool {
	return true
}

type boolean struct {
	value bool
}

func createBoolean(b bool) boolean {
	return boolean{
		value: b,
	}
}

func (_ boolean) isPrimitive() bool {
	return true
}

type number struct {
	value float64
}

func createNumber(f float64) number {
	return number{
		value: f,
	}
}

func (_ number) isPrimitive() bool {
	return true
}

type array struct {
	list []Expression
}

func (_ array) isPrimitive() bool {
	return false
}

type dict struct {
	list map[Expression]Expression
}

func (_ dict) isPrimitive() bool {
	return false
}

type index struct {
	arr  Expression
	expr Expression
}

func (_ index) isPrimitive() bool {
	return false
}

type parameter struct {
	ident string
	expr  Expression
}

func createParameter(ident string) parameter {
	return parameter{
		ident: ident,
	}
}

func (_ parameter) isPrimitive() bool {
	return false
}

type function struct {
	ident  string
	params []Expression
	body   Expression
}

func (_ function) isPrimitive() bool {
	return false
}

type assign struct {
	ident Expression
	right Expression
}

func (_ assign) isPrimitive() bool {
	return false
}

type call struct {
	ident string
	args  []Expression
}

func (_ call) isPrimitive() bool {
	return false
}

type returned struct {
	right Expression
}

func (_ returned) isPrimitive() bool {
	return false
}

type unary struct {
	op    rune
	right Expression
}

func (_ unary) isPrimitive() bool {
	return false
}

type binary struct {
	op    rune
	left  Expression
	right Expression
}

func (_ binary) isPrimitive() bool {
	return false
}

type script struct {
	list    []Expression
	symbols map[string]Expression
}

func (_ script) isPrimitive() bool {
	return false
}

type while struct {
	cdt  Expression
	body Expression
}

func (_ while) isPrimitive() bool {
	return false
}

type breaked struct{}

func (_ breaked) isPrimitive() bool {
	return false
}

type continued struct{}

func (_ continued) isPrimitive() bool {
	return false
}

type test struct {
	cdt Expression
	csq Expression
	alt Expression
}

func (_ test) isPrimitive() bool {
	return false
}

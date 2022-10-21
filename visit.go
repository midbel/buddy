package buddy

import (
	"fmt"
	"strings"
)

type errorsList []error

func (e errorsList) Error() string {
	var str strings.Builder
	for i := range e {
		if i > 0 {
			str.WriteString("\n")
		}
		str.WriteString(e[i].Error())
	}
	return str.String()
}

type visitFunc func(Expression, *Resolver) (Expression, error)

func traverse(expr Expression, env *Resolver, visit []visitFunc) (Expression, error) {
	var err error
	for _, v := range visit {
		expr, err = v(expr, env)
		if err != nil {
			break
		}
	}
	return expr, err
}

func replaceExprList(list []Expression, env *Resolver) ([]Expression, error) {
	var err error
	for i := range list {
		list[i], err = replaceValue(list[i], env)
		if err != nil {
			return nil, err
		}
	}
	return list, err
}

func replaceValue(expr Expression, env *Resolver) (Expression, error) {
	var err error
	switch e := expr.(type) {
	case script:
		e.list, err = replaceExprList(e.list, env)
		return e, err
	case call:
		e.args, err = replaceExprList(e.args, env)
		return e, err
	case unary:
		e.right, err = replaceValue(e.right, env)
		if err != nil {
			return nil, err
		}
		if e.right.isPrimitive() {
			res, err := evalUnary(e, env)
			if err != nil {
				return nil, err
			}
			return createPrimitive(res.Raw())
		}
		return e, nil
	case binary:
		e.left, err = replaceValue(e.left, env)
		if err != nil {
			return nil, err
		}
		e.right, err = replaceValue(e.right, env)
		if err != nil {
			return nil, err
		}
		if e.left.isPrimitive() && e.right.isPrimitive() {
			res, err := evalBinary(e, env)
			if err != nil {
				return nil, err
			}
			return createPrimitive(res.Raw())
		}
		return e, nil
	case assign:
		e.right, err = replaceValue(e.right, env)
		return e, err
	case test:
		e.cdt, err = replaceValue(e.cdt, env)
		if err != nil {
			return nil, err
		}
		e.csq, err = replaceValue(e.csq, env)
		if err != nil {
			return nil, err
		}
		if e.alt == nil {
			return e, nil
		}
		e.alt, err = replaceValue(e.alt, env)
		return e, err
	case while:
		e.cdt, err = replaceValue(e.cdt, env)
		if err != nil {
			return nil, err
		}
		e.body, err = replaceValue(e.body, env)
		return e, err
	case returned:
		e.right, err = replaceValue(e.right, env)
		return e, nil
	case function:
		e.params, err = replaceExprList(e.params, env)
		if err != nil {
			return nil, err
		}
		e.body, err = replaceValue(e.body, env)
		return e, nil
	case parameter:
		e.expr, err = replaceValue(e.expr, env)
		return e, err
	case dict:
		for k, v := range e.list {
			e.list[k], err = replaceValue(v, env)
			if err != nil {
				return nil, err
			}
		}
		return e, nil
	case array:
		e.list, err = replaceExprList(e.list, env)
		return e, err
	case index:
		e.expr, err = replaceValue(e.expr, env)
		return e, nil
	default:
		return expr, nil
	}
}

type counter struct {
	identifiers map[string]int
}

func createCounter() *counter {
	return &counter{
		identifiers: make(map[string]int),
	}
}

func (c *counter) unused() error {
	var list errorsList
	for ident, count := range c.identifiers {
		if count <= 1 {
			list = append(list, unusedVar(ident))
		}
	}
	if len(list) == 0 {
		return nil
	}
	return list
}

func (c *counter) set(ident string) {
	c.identifiers[ident] += 1
}

func (c counter) exists(ident string) bool {
	return c.identifiers[ident] > 0
}

func checkVariables(expr Expression, env *Resolver) (Expression, error) {
	return expr, checkVars(expr, env, createCounter())
}

func checkVars(expr Expression, env *Resolver, count *counter) error {
	var err error
	switch e := expr.(type) {
	case script:
		for i := range e.list {
			if err = checkVars(e.list[i], env, count); err != nil {
				break
			}
		}
		if err == nil {
			err = count.unused()
		}
	case parameter:
		if err = checkVars(e.expr, env, count); err != nil {
			break
		}
		count.set(e.ident)
	case function:
		tmp := createCounter()
		for i := range e.params {
			if err = checkVars(e.params[i], env, tmp); err != nil {
				break
			}
		}
		if err == nil {
			err = checkVars(e.body, env, tmp)
		}
		if err == nil {
			err = tmp.unused()
		}
	case returned:
		err = checkVars(e.right, env, count)
	case binary:
		if err = checkVars(e.left, env, count); err != nil {
			break
		}
		err = checkVars(e.right, env, count)
	case unary:
		err = checkVars(e.right, env, count)
	case while:
		if err = checkVars(e.cdt, env, count); err != nil {
			break
		}
		err = checkVars(e.body, env, count)
	case test:
		if err = checkVars(e.cdt, env, count); err != nil {
			break
		}
		if err = checkVars(e.csq, env, count); err != nil {
			break
		}
		err = checkVars(e.alt, env, count)
	case call:
		for i := range e.args {
			if err = checkVars(e.args[i], env, count); err != nil {
				break
			}
		}
	case assign:
		count.set(e.ident)
		err = checkVars(e.right, env, count)
	case variable:
		if !count.exists(e.ident) {
			err = undefinedVar(e.ident)
		} else {
			count.set(e.ident)
		}
	case array:
		for i := range e.list {
			if err = checkVars(e.list[i], env, count); err != nil {
				break
			}
		}
	case dict:
		for _, v := range e.list {
			if err = checkVars(v, env, count); err != nil {
				break
			}
		}
	case index:
		if err = checkVars(e.arr, env, count); err != nil {
			break
		}
		err = checkVars(e.expr, env, count)
	default:
	}
	return err
}

func replaceFunctionArgs(expr Expression, env *Resolver) (Expression, error) {
	return expr, nil
}

func inlineFunctionCall(expr Expression, env *Resolver) (Expression, error) {
	return expr, nil
}

func undefinedVar(ident string) error {
	return fmt.Errorf("%s: undefined variable!", ident)
}

func unusedVar(ident string) error {
	return fmt.Errorf("%s: variable defined but never used", ident)
}

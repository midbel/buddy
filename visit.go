package buddy

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

func inlineFunctionArgs(expr Expression, env *Resolver) (Expression, error) {
	return expr, nil
}

func inlineFunctionCall(expr Expression, env *Resolver) (Expression, error) {
	return expr, nil
}

package buddy

import (
	"fmt"
)

func init() {
	visitors = []visitFunc{
		trackVariables,
		trackImport,
		trackCyclic,
		trackLoop,
		replaceFunctionArgs,
		inlineFunctionCall,
		trackValue,
	}
}

type visitFunc func(Expression, *Resolver) (Expression, error)

var visitors []visitFunc

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
		list[i], err = trackValue(list[i], env)
		if err != nil {
			return nil, err
		}
	}
	return list, err
}

func trackValue(expr Expression, env *Resolver) (Expression, error) {
	var err error
	switch e := expr.(type) {
	case script:
		e.list, err = replaceExprList(e.list, env)
		return e, err
	case call:
		e.args, err = replaceExprList(e.args, env)
		return e, err
	case unary:
		e.right, err = trackValue(e.right, env)
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
		e.left, err = trackValue(e.left, env)
		if err != nil {
			return nil, err
		}
		e.right, err = trackValue(e.right, env)
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
		e.right, err = trackValue(e.right, env)
		return e, err
	case test:
		e.cdt, err = trackValue(e.cdt, env)
		if err != nil {
			return nil, err
		}
		e.csq, err = trackValue(e.csq, env)
		if err != nil {
			return nil, err
		}
		if e.alt == nil {
			return e, nil
		}
		e.alt, err = trackValue(e.alt, env)
		return e, err
	case while:
		e.cdt, err = trackValue(e.cdt, env)
		if err != nil {
			return nil, err
		}
		e.body, err = trackValue(e.body, env)
		return e, err
	case foreach:
		e.iter, err = trackValue(e.iter, env)
		if err != nil {
			return nil, err
		}
		e.body, err = trackValue(e.body, env)
		return e, err
	case returned:
		e.right, err = trackValue(e.right, env)
		return e, nil
	case function:
		e.params, err = replaceExprList(e.params, env)
		if err != nil {
			return nil, err
		}
		e.body, err = trackValue(e.body, env)
		return e, nil
	case parameter:
		e.expr, err = trackValue(e.expr, env)
		return e, err
	case dict:
		for k, v := range e.list {
			e.list[k], err = trackValue(v, env)
			if err != nil {
				return nil, err
			}
		}
		return e, nil
	case array:
		e.list, err = replaceExprList(e.list, env)
		return e, err
	case index:
		e.expr, err = trackValue(e.expr, env)
		return e, nil
	default:
		return expr, nil
	}
}

func trackCyclic(expr Expression, env *Resolver) (Expression, error) {
	return expr, nil
}

func trackImport(expr Expression, env *Resolver) (Expression, error) {
	var track func(Expression, bool) error
	track = func(expr Expression, global bool) error {
		switch e := expr.(type) {
		case script:
			for i := range e.list {
				if err := track(e.list[i], global); err != nil {
					return err
				}
			}
		case function:
			return track(e.body, false)
		case while:
			return track(e.body, false)
		case test:
			err := track(e.csq, false)
			if err != nil {
				return err
			}
			if e.alt != nil {
				err = track(e.alt, false)
			}
			return err
		case module:
			if !global {
				return fmt.Errorf("import expression only allows at global level")
			}
		default:
		}
		return nil
	}
	return expr, track(expr, true)
}

func trackLoop(expr Expression, env *Resolver) (Expression, error) {
	var track func(expr Expression, inloop bool) error
	track = func(expr Expression, inloop bool) error {
		switch e := expr.(type) {
		case script:
			for i := range e.list {
				if err := track(e.list[i], inloop); err != nil {
					return err
				}
			}
		case call:
			for i := range e.args {
				if err := track(e.args[i], false); err != nil {
					return err
				}
			}
		case parameter:
			if err := track(e.expr, false); err != nil {
				return err
			}
		case array:
			for i := range e.list {
				if err := track(e.list[i], false); err != nil {
					return err
				}
			}
		case dict:
			for k := range e.list {
				if err := track(e.list[k], false); err != nil {
					return err
				}
			}
		case index:
			err := track(e.arr, false)
			if err != nil {
				return err
			}
			return track(e.expr, false)
		case function:
			return track(e.body, false)
		case while:
			err := track(e.cdt, false)
			if err != nil {
				return err
			}
			return track(e.body, true)
		case foreach:
			err := track(e.iter, false)
			if err != nil {
				return err
			}
			return track(e.body, true)
		case test:
			err := track(e.cdt, false)
			if err != nil {
				return err
			}
			if err = track(e.csq, inloop); err != nil {
				return err
			}
			if e.alt != nil {
				err = track(e.alt, inloop)
			}
			return err
		case breaked, continued:
			if !inloop {
				return fmt.Errorf("break/continue used outside of a loop")
			}
		default:
		}
		return nil
	}
	return expr, track(expr, false)
}

func trackVariables(expr Expression, env *Resolver) (Expression, error) {
	k := track()
	return expr, k.check(expr, env)
}

type vartracker map[string]int

func track() vartracker {
	return make(vartracker)
}

func (k vartracker) unused() error {
	var list errorsList
	for ident, count := range k {
		if count <= 1 {
			list = append(list, unusedVar(ident))
		}
	}
	if len(list) == 0 {
		return nil
	}
	return list
}

func (k vartracker) set(ident string) {
	k[ident] += 1
}

func (k vartracker) exists(ident string) bool {
	return k[ident] > 0
}

func (k vartracker) check(expr Expression, env *Resolver) error {
	var err error
	switch e := expr.(type) {
	case script:
		for i := range e.list {
			if err = k.check(e.list[i], env); err != nil {
				break
			}
		}
		if err == nil {
			err = k.unused()
		}
	case parameter:
		if err = k.check(e.expr, env); err != nil {
			break
		}
		k.set(e.ident)
	case function:
		tmp := track()
		for i := range e.params {
			if err = tmp.check(e.params[i], env); err != nil {
				break
			}
		}
		if err == nil {
			err = tmp.check(e.body, env)
		}
		if err == nil {
			err = tmp.unused()
		}
	case returned:
		err = k.check(e.right, env)
	case binary:
		if err = k.check(e.left, env); err != nil {
			break
		}
		err = k.check(e.right, env)
	case unary:
		err = k.check(e.right, env)
	case while:
		if err = k.check(e.cdt, env); err != nil {
			break
		}
		err = k.check(e.body, env)
	case test:
		if err = k.check(e.cdt, env); err != nil {
			break
		}
		if err = k.check(e.csq, env); err != nil {
			break
		}
		err = k.check(e.alt, env)
	case call:
		for i := range e.args {
			arg := e.args[i]
			if p, ok := arg.(parameter); ok {
				arg = p.expr
			}
			if err = k.check(arg, env); err != nil {
				break
			}
		}
	case foreach:
		k.set(e.ident)
		if err = k.check(e.iter, env); err != nil {
			break
		}
		err = k.check(e.body, env)
	case assign:
		if v, ok := e.ident.(variable); ok {
			k.set(v.ident)
		}
		err = k.check(e.right, env)
	case path:
		err = k.check(e.right, env)
	case variable:
		if !k.exists(e.ident) {
			err = undefinedVar(e.ident)
		} else {
			k.set(e.ident)
		}
	case array:
		for i := range e.list {
			if err = k.check(e.list[i], env); err != nil {
				break
			}
		}
	case dict:
		for _, v := range e.list {
			if err = k.check(v, env); err != nil {
				break
			}
		}
	case index:
		if err = k.check(e.arr, env); err != nil {
			break
		}
		err = k.check(e.expr, env)
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

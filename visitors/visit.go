//go:build exclude

package visitors

import (
	"fmt"
)

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
			for i := range e.list {
				if err := track(e.list[i], false); err != nil {
					return err
				}
			}
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

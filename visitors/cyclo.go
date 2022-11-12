package visitors

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/token"
)

type Cyclo struct {
	count int
}

func Complexity() *Cyclo {
	return &Cyclo{
		count: 1,
	}
}

func (c *Cyclo) Count(expr ast.Expression) (int, error) {
	if expr == nil {
		return c.count, nil
	}
	switch e := expr.(type) {
	case ast.Literal:
	case ast.Double:
	case ast.Integer:
	case ast.Boolean:
	case ast.Variable:
	case ast.Array:
		for i := range e.List {
			if _, err := c.Count(e.List[i]); err != nil {
				return c.count, err
			}
		}
	case ast.Dict:
		for i := range e.List {
			if _, err := c.Count(e.List[i]); err != nil {
				return c.count, err
			}
		}
	case ast.Index:
		if _, err := c.Count(e.Arr); err != nil {
			return c.count, err
		}
		for i := range e.List {
			if _, err := c.Count(e.List[i]); err != nil {
				return c.count, err
			}
		}
	case ast.Slice:
		if _, err := c.Count(e.Start); err != nil {
			return c.count, err
		}
		if _, err := c.Count(e.End); err != nil {
			return c.count, err
		}
		if _, err := c.Count(e.Step); err != nil {
			return c.count, err
		}
	case ast.Path:
		return c.Count(e.Right)
	case ast.Call:
		for i := range e.Args {
			if _, err := c.Count(e.Args[i]); err != nil {
				return c.count, err
			}
		}
	case ast.Parameter:
		return c.Count(e.Expr)
	case ast.Assert:
		return c.Count(e.Expr)
	case ast.Let:
		return c.Count(e.Right)
	case ast.Assign:
		if _, err := c.Count(e.Ident); err != nil {
			return c.count, err
		}
		return c.Count(e.Right)
	case ast.Unary:
		return c.Count(e.Right)
	case ast.Binary:
		if _, err := c.Count(e.Left); err != nil {
			return c.count, err
		}
		if _, err := c.Count(e.Right); err != nil {
			return c.count, err
		}
		if e.Op == token.Or || e.Op == token.And {
			c.count++
		}
	case ast.ListComp:
		if _, err := c.Count(e.Body); err != nil {
			return c.count, err
		}
		for i := range e.List {
			if _, err := c.Count(e.List[i].Iter); err != nil {
				return c.count, err
			}
			for _, cdt := range e.List[i].Cdt {
				c.count++
				if _, err := c.Count(cdt); err != nil {
					return c.count, err
				}
			}
		}
	case ast.DictComp:
		if _, err := c.Count(e.Key); err != nil {
			return c.count, err
		}
		if _, err := c.Count(e.Val); err != nil {
			return c.count, err
		}
		for i := range e.List {
			if _, err := c.Count(e.List[i].Iter); err != nil {
				return c.count, err
			}
			for _, cdt := range e.List[i].Cdt {
				c.count++
				if _, err := c.Count(cdt); err != nil {
					return c.count, err
				}
			}
		}
	case ast.Test:
		c.count++
		if _, err := c.Count(e.Cdt); err != nil {
			return c.count, err
		}
		if _, err := c.Count(e.Csq); err != nil {
			return c.count, err
		}
		return c.Count(e.Alt)
	case ast.While:
		c.count++
		if _, err := c.Count(e.Cdt); err != nil {
			return c.count, err
		}
		return c.Count(e.Body)
	case ast.For:
		c.count++
		if _, err := c.Count(e.Init); err != nil {
			return c.count, err
		}
		if _, err := c.Count(e.Cdt); err != nil {
			return c.count, err
		}
		if _, err := c.Count(e.Incr); err != nil {
			return c.count, err
		}
		return c.Count(e.Body)
	case ast.ForEach:
		c.count++
		if _, err := c.Count(e.Iter); err != nil {
			return c.count, err
		}
		return c.Count(e.Body)
	case ast.Import:
	case ast.Script:
		for i := range e.List {
			if _, err := c.Count(e.List[i]); err != nil {
				return 0, err
			}
		}
	case ast.Function:
		return c.Count(e.Body)
	case ast.Return:
		return c.Count(e.Right)
	case ast.Break:
	case ast.Continue:
	default:
		return c.count, fmt.Errorf("unsupported expression type: %T", expr)
	}
	return c.count, nil
}

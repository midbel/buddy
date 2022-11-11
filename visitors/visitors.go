package visitors

import (
	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
	"golang.org/x/exp/constraints"
)

type Visitor interface {
	Visit(ast.Expression) (ast.Expression, error)
}

func Visit(expr ast.Expression, visits []Visitor) (ast.Expression, error) {
	var (
		list faults.ErrorList
		err  error
	)
	if s, ok := expr.(ast.Script); ok {
		for k, v := range s.Symbols {
			s.Symbols[k], err = Visit(v, visits)
			if err != nil {
				list.Append(err)
			}
		}
	}
	for i := range visits {
		expr, err = visits[i].Visit(expr)
		if err != nil {
			list.Append(err)
		}
	}
	if list.Size() > 0 {
		return nil, &list
	}
	return expr, nil
}

type Counter[T constraints.Ordered] struct {
	parent *Counter[T]
	data   map[T]int
}

func EmptyCounter[T constraints.Ordered]() *Counter[T] {
	return NewCounter[T](nil)
}

func NewCounter[T constraints.Ordered](parent *Counter[T]) *Counter[T] {
	return &Counter[T]{
		parent: parent,
		data:   make(map[T]int),
	}
}

func (c *Counter[T]) Zero() []T {
	var vs []T
	for k, v := range c.data {
		if v <= 1 {
			vs = append(vs, k)
		}
	}
	return vs
}

func (c *Counter[T]) Incr(ident T) {
	c.data[ident]++
}

func (c *Counter[T]) Decr(ident T) {
	c.data[ident]--
}

func (c *Counter[T]) Exists(ident T) bool {
	_, ok := c.data[ident]
	if !ok && c.parent != nil {
		return c.parent.Exists(ident)
	}
	return ok
}

func (c *Counter[T]) Wrap() *Counter[T] {
	return NewCounter(c)
}

func (c *Counter[T]) Unwrap() *Counter[T] {
	if c.parent == nil {
		return c
	}
	return c.parent
}

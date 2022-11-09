package visitors

import (
	"github.com/midbel/buddy/ast"
)

type Visitor interface {
	Visit(ast.Expression) (ast.Expression, error)
}

func Visit(expr ast.Expression, visits []Visitor) (ast.Expression, error) {
	var err error
	for i := range visits {
		expr, err = visits[i].Visit(expr)
		if err != nil {
			break
		}
	}
	return expr, err
}

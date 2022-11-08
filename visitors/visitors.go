package visitors

import (
	"github.com/midbel/buddy/ast"
)

type Visitor interface {
	Visit(ast.Expression) (ast.Expression, error)
}

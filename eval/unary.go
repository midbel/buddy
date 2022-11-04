package eval

import (
	"fmt"

	"github.com/midbel/buddy/token"
	"github.com/midbel/buddy/types"
)

func executeUnary(op rune, p types.Primitive) (types.Primitive, error) {
	switch op {
	case token.Not:
		p, ok := p.(interface {
			Not() (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return p.Not()
	case token.BinNot:
		p, ok := p.(interface{ Bnot() types.Primitive })
		if !ok {
			return nil, types.ErrOperation
		}
		return p.Bnot(), nil
	case token.Sub:
		p, ok := p.(interface {
			Rev() (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return p.Rev()
	default:
		return nil, fmt.Errorf("unary operator not recognized")
	}
}

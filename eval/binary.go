package eval

import (
	"fmt"

	"github.com/midbel/buddy/token"
	"github.com/midbel/buddy/types"
)

func executeBinary(op rune, left, right types.Primitive) (types.Primitive, error) {
	switch op {
	case token.Add:
		left, ok := left.(interface {
			Add(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Add(right)
	case token.Sub:
		left, ok := left.(interface {
			Sub(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Sub(right)
	case token.Mul:
		left, ok := left.(interface {
			Mul(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Mul(right)
	case token.Div:
		left, ok := left.(interface {
			Div(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Div(right)
	case token.Pow:
		left, ok := left.(interface {
			Pow(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Pow(right)
	case token.Mod:
		left, ok := left.(interface {
			Mod(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Mod(right)
	case token.Lshift:
		left, ok := left.(interface {
			Lshift(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Lshift(right)
	case token.Rshift:
		left, ok := left.(interface {
			Rshift(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Rshift(right)
	case token.BinAnd:
		left, ok := left.(interface {
			And(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.And(right)
	case token.BinOr:
		left, ok := left.(interface {
			Or(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Or(right)
	case token.BinXor:
		left, ok := left.(interface {
			Xor(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Xor(right)
	case token.Eq:
		left, ok := left.(interface {
			Eq(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Eq(right)
	case token.Ne:
		left, ok := left.(interface {
			Ne(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Ne(right)
	case token.Lt:
		left, ok := left.(interface {
			Lt(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Lt(right)
	case token.Le:
		left, ok := left.(interface {
			Le(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Le(right)
	case token.Gt:
		left, ok := left.(interface {
			Gt(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Gt(right)
	case token.Ge:
		left, ok := left.(interface {
			Ge(types.Primitive) (types.Primitive, error)
		})
		if !ok {
			return nil, types.ErrOperation
		}
		return left.Ge(right)
	case token.And:
		return types.And(left, right)
	case token.Or:
		return types.Or(left, right)
	default:
		return nil, fmt.Errorf("binary operator not recognized")
	}
}

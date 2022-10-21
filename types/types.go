package types

import (
	"errors"
	"fmt"
)

var (
	ErrIncompatible = errors.New("incompatible type")
	ErrOperation    = errors.New("unsupported operation")
	ErrZero         = errors.New("division by zero")
)

type Sizeable interface {
	Len() int
}

type Comparable interface {
	Eq(Primitive) (Primitive, error)
	Ne(Primitive) (Primitive, error)
	Lt(Primitive) (Primitive, error)
	Le(Primitive) (Primitive, error)
	Gt(Primitive) (Primitive, error)
	Ge(Primitive) (Primitive, error)
}

type Primitive interface {
	fmt.Stringer

	Raw() any

	Rev() (Primitive, error)
	Not() (Primitive, error)

	Add(Primitive) (Primitive, error)
	Sub(Primitive) (Primitive, error)
	Div(Primitive) (Primitive, error)
	Mod(Primitive) (Primitive, error)
	Mul(Primitive) (Primitive, error)
	Pow(Primitive) (Primitive, error)

	Comparable

	True() bool
}

func CreatePrimitive(value any) (Primitive, error) {
	switch v := value.(type) {
	case string:
		return CreateString(v), nil
	case int64:
		return CreateInt(v), nil
	case float64:
		return CreateFloat(v), nil
	default:
		return nil, fmt.Errorf("%s can not be transformed to Primitive")
	}
}

func And(left, right Primitive) (Primitive, error) {
	b := left.True() && right.True()
	return Bool{value: b}, nil
}

func Or(left, right Primitive) (Primitive, error) {
	b := left.True() || right.True()
	return Bool{value: b}, nil
}

func unsupportedOp(op string, val Primitive) error {
	return fmt.Errorf("%s: %w for type %s", op, typeName(val))
}

func incompatibleType(op string, left, right Primitive) error {
	return fmt.Errorf("%s: %w %s/%s", op, ErrIncompatible, typeName(left), typeName(right))
}

func typeName(val Primitive) string {
	switch val.(type) {
	case String:
		return "string"
	case Int:
		return "integer"
	case Float:
		return "float"
	case Bool:
		return "boolean"
	// case Array:
	// 	return "array"
	// case Dict:
	// 	return "dict"
	default:
		return "?"
	}
}

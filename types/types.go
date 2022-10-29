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

type Iterable interface {
	Iter(func(Primitive) error) error
}

type Container interface {
	Set(Primitive, Primitive) (Primitive, error)
	Get(Primitive) (Primitive, error)
}

type Comparable interface {
	Eq(Primitive) (Primitive, error)
	Ne(Primitive) (Primitive, error)
	Lt(Primitive) (Primitive, error)
	Le(Primitive) (Primitive, error)
	Gt(Primitive) (Primitive, error)
	Ge(Primitive) (Primitive, error)
}

type Calculable interface {
	Rev() (Primitive, error)
	Add(Primitive) (Primitive, error)
	Sub(Primitive) (Primitive, error)
	Div(Primitive) (Primitive, error)
	Mod(Primitive) (Primitive, error)
	Mul(Primitive) (Primitive, error)
	Pow(Primitive) (Primitive, error)
}

type BinaryCalculable interface {
	Lshift(Primitive) (Primitive, error)
	Rshift(Primitive) (Primitive, error)
	And(Primitive) (Primitive, error)
	Or(Primitive) (Primitive, error)
	Xor(Primitive) (Primitive, error)
	Bnot() Primitive
}

type Primitive interface {
	fmt.Stringer
	Raw() any

	True() bool
	Not() (Primitive, error)
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

func Type(val Primitive) (string, error) {
	name := typeName(val)
	if name == "" {
		return name, fmt.Errorf("unrecognized primitive type")
	}
	return name, nil
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
	case Array:
		return "array"
	case Dict:
		return "dict"
	default:
		return "?"
	}
}

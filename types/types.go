package types

import (
	"errors"
	"fmt"
)

var (
	ErrIncompatible = errors.New("incompatible type")
	ErrOperation    = errors.New("unsupported operation")
	ErrZero         = errors.New("division by zero")
	ErrAssert       = errors.New("assertion failed")
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

func IterationError(val Primitive) error {
	return fmt.Errorf("%w: %s is not iterable", ErrOperation, typeName(val))
}

func ContainerError(val Primitive) error {
	return fmt.Errorf("%w: %s can not be used as a container", ErrOperation, typeName(val))
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

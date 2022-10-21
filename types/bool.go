package types

import (
	"strconv"
)

type Bool struct {
	value bool
}

func CreateBool(b bool) Primitive {
	return Bool{
		value: b,
	}
}

func (b Bool) Raw() any {
	return b.value
}

func (b Bool) Rev() (Primitive, error) {
	return nil, unsupportedOp("reverse", b)
}

func (b Bool) Not() (Primitive, error) {
	b.value = !b.value
	return b, nil
}

func (b Bool) String() string {
	return strconv.FormatBool(b.value)
}

func (b Bool) Add(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("addition", b)
}

func (b Bool) Sub(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("subtraction", b)
}

func (b Bool) Div(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("division", b)
}

func (b Bool) Mul(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("multiply", b)
}

func (b Bool) Mod(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("modulo", b)
}

func (b Bool) Pow(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("power", b)
}

func (b Bool) True() bool {
	return b.value
}

func (b Bool) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(Bool)
	if !ok {
		return nil, incompatibleType("eq", b, other)
	}
	return CreateBool(b.value == x.value), nil
}

func (b Bool) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(Bool)
	if !ok {
		return nil, incompatibleType("ne", b, other)
	}
	return CreateBool(b.value != x.value), nil
}

func (b Bool) Lt(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("lt", b)
}

func (b Bool) Le(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("le", b)
}

func (b Bool) Gt(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("gt", b)
}

func (b Bool) Ge(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("ge", b)
}

package types

import (
	"math"
	"strconv"
)

type Int struct {
	value int64
}

func CreateInt(i int64) Primitive {
	return Int{
		value: i,
	}
}

func (i Int) Raw() any {
	return i.value
}

func (i Int) Rev() (Primitive, error) {
	i.value = -i.value
	return i, nil
}

func (i Int) Not() (Primitive, error) {
	return CreateBool(!i.True()), nil
}

func (i Int) String() string {
	return strconv.FormatInt(i.value, 10)
}

func (i Int) Add(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		i.value += x.value
	case Float:
		f := float64(i.value) + x.value
		return Float{value: f}, nil
	case String:
		s := i.String() + x.String()
		return String{str: s}, nil
	default:
		return nil, incompatibleType("addition", i, other)
	}
	return i, nil
}

func (i Int) Sub(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		i.value -= x.value
	case Float:
		f := float64(i.value) - x.value
		return Float{value: f}, nil
	default:
		return nil, incompatibleType("subtraction", i, other)
	}
	return i, nil
}

func (i Int) Div(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		i.value /= x.value
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f := float64(i.value) / x.value
		return Float{value: f}, nil
	default:
		return nil, incompatibleType("division", i, other)
	}
	return i, nil
}

func (i Int) Mul(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		i.value *= x.value
	case Float:
		f := float64(i.value) * x.value
		return Float{value: f}, nil
	default:
		return nil, incompatibleType("multiply", i, other)
	}
	return i, nil
}

func (i Int) Mod(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		i.value %= x.value
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f := math.Mod(float64(i.value), x.value)
		return Float{value: f}, nil
	default:
		return nil, incompatibleType("modulo", i, other)
	}
	return i, nil
}

func (i Int) Pow(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		v := math.Pow(float64(i.value), float64(x.value))
		i.value = int64(v)
	case Float:
		f := math.Pow(float64(i.value), x.value)
		return Float{value: f}, nil
	default:
		return nil, incompatibleType("power", i, other)
	}
	return i, nil
}

func (i Int) Lshift(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("left-shift", i, other)
	}
	i.value <<= x.value
	return i, nil
}

func (i Int) Rshift(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("right-shift", i, other)
	}
	i.value >>= x.value
	return i, nil
}

func (i Int) And(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("binary-and", i, other)
	}
	i.value &= x.value
	return i, nil
}

func (i Int) Or(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("binary-or", i, other)
	}
	i.value |= x.value
	return i, nil
}

func (i Int) Xor(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("binary-xor", i, other)
	}
	i.value ^= x.value
	return i, nil
}

func (i Int) Bnot() Primitive {
	i.value = ^i.value
	return i
}

func (i Int) True() bool {
	return i.value != 0
}

func (i Int) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("eq", i, other)
	}
	return CreateBool(i.value == x.value), nil
}

func (i Int) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("ne", i, other)
	}
	return CreateBool(i.value != x.value), nil
}

func (i Int) Lt(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("lt", i, other)
	}
	return CreateBool(i.value < x.value), nil
}

func (i Int) Le(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("le", i, other)
	}
	return CreateBool(i.value <= x.value), nil
}

func (i Int) Gt(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("gt", i, other)
	}
	return CreateBool(i.value > x.value), nil
}

func (i Int) Ge(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, incompatibleType("ge", i, other)
	}
	return CreateBool(i.value >= x.value), nil
}

package types

import (
	"math"
	"strconv"
)

type Float struct {
	value float64
}

func CreateFloat(f float64) Primitive {
	return Float{
		value: f,
	}
}

func (f Float) Raw() any {
	return f.value
}

func (f Float) Rev() (Primitive, error) {
	f.value = -f.value
	return f, nil
}

func (f Float) Not() (Primitive, error) {
	return CreateBool(!f.True()), nil
}

func (f Float) String() string {
	return strconv.FormatFloat(f.value, 'g', -1, 64)
}

func (f Float) Add(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value += float64(x.value)
	case Float:
		f.value += x.value
	case String:
		s := f.String() + x.String()
		return String{str: s}, nil
	default:
		return nil, incompatibleType("addition", f, other)
	}
	return f, nil
}

func (f Float) Sub(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value -= float64(x.value)
	case Float:
		f.value -= x.value
	default:
		return nil, incompatibleType("subtraction", f, other)
	}
	return f, nil
}

func (f Float) Div(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value /= float64(x.value)
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value /= x.value
	default:
		return nil, incompatibleType("division", f, other)
	}
	return f, nil
}

func (f Float) Mul(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value *= float64(x.value)
	case Float:
		f.value *= x.value
	default:
		return nil, incompatibleType("multiply", f, other)
	}
	return f, nil
}

func (f Float) Mod(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value = math.Mod(f.value, float64(x.value))
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value = math.Mod(f.value, x.value)
	default:
		return nil, incompatibleType("modulo", f, other)
	}
	return f, nil
}

func (f Float) Pow(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value = math.Pow(f.value, float64(x.value))
	case Float:
		f.value = math.Pow(f.value, x.value)
	default:
		return nil, incompatibleType("power", f, other)
	}
	return f, nil
}

func (f Float) True() bool {
	return f.value != 0
}

func (f Float) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, incompatibleType("eq", f, other)
	}
	return CreateBool(f.value == x.value), nil
}

func (f Float) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, incompatibleType("ne", f, other)
	}
	return CreateBool(f.value != x.value), nil
}

func (f Float) Lt(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, incompatibleType("lt", f, other)
	}
	return CreateBool(f.value < x.value), nil
}

func (f Float) Le(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, incompatibleType("le", f, other)
	}
	return CreateBool(f.value <= x.value), nil
}

func (f Float) Gt(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, incompatibleType("gt", f, other)
	}
	return CreateBool(f.value > x.value), nil
}

func (f Float) Ge(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, incompatibleType("ge", f, other)
	}
	return CreateBool(f.value >= x.value), nil
}

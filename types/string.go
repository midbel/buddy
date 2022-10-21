package types

import (
	"strings"
)

type String struct {
	str string
}

func CreateString(str string) Primitive {
	return String{
		str: str,
	}
}

func (s String) Len() int {
	return len(s.str)
}

func (s String) Raw() any {
	return s.str
}

func (s String) String() string {
	return s.str
}

func (s String) Rev() (Primitive, error) {
	return nil, unsupportedOp("reverse", s)
}

func (s String) Not() (Primitive, error) {
	return CreateBool(!s.True()), nil
}

func (s String) Add(other Primitive) (Primitive, error) {
	var str string
	switch x := other.(type) {
	case Int:
		str = x.String()
	case Float:
		str = x.String()
	case String:
		str = x.String()
	default:
		return nil, incompatibleType("addition", s, other)
	}
	s.str += str
	return s, nil
}

func (s String) Sub(other Primitive) (Primitive, error) {
	var part int
	switch x := other.(type) {
	case Int:
		part = int(x.value)
	case Float:
		part = int(x.value)
	default:
		return nil, incompatibleType("subtraction", s, other)
	}
	if part > len(s.str) {
		s.str = ""
		return s, nil
	}
	if part < 0 {
		s.str = s.str[-part:]
	} else {
		s.str = s.str[:part]
	}
	return s, nil
}

func (s String) Div(other Primitive) (Primitive, error) {
	var part int
	switch x := other.(type) {
	case Int:
		part = int(x.value)
	case Float:
		part = int(x.value)
	default:
		return nil, incompatibleType("division", s, other)
	}
	if part == 0 {
		return s, nil
	}
	offset := len(s.str) / part
	s.str = s.str[:offset]
	return s, nil
}

func (s String) Mul(other Primitive) (Primitive, error) {
	var count int
	switch x := other.(type) {
	case Int:
		count = int(x.value)
	case Float:
		count = int(x.value)
	default:
		return nil, incompatibleType("multiply", s, other)
	}
	s.str = strings.Repeat(s.str, count)
	return s, nil
}

func (s String) Mod(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("modulo", s)
}

func (s String) Pow(_ Primitive) (Primitive, error) {
	return nil, unsupportedOp("power", s)
}

func (s String) True() bool {
	return s.str != ""
}

func (s String) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, incompatibleType("eq", s, other)
	}
	return CreateBool(s.str == x.str), nil
}

func (s String) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, incompatibleType("ne", s, other)
	}
	return CreateBool(s.str != x.str), nil
}

func (s String) Lt(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, incompatibleType("lt", s, other)
	}
	return CreateBool(s.str < x.str), nil
}

func (s String) Le(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, incompatibleType("le", s, other)
	}
	return CreateBool(s.str <= x.str), nil
}

func (s String) Gt(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, incompatibleType("gt", s, other)
	}
	return CreateBool(s.str > x.str), nil
}

func (s String) Ge(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, incompatibleType("ge", s, other)
	}
	return CreateBool(s.str >= x.str), nil
}

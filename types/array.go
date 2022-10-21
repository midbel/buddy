package types

import (
	"fmt"
)

type Array struct {
	values []Primitive
}

func CreateArray(list []Primitive) Primitive {
	vs := make([]Primitive, len(list))
	copy(vs, list)
	return Array{
		values: list,
	}
}

func (a Array) String() string {
	return fmt.Sprintf("%s", a.values)
}

func (a Array) Raw() any {
	var list []any
	for i := range a.values {
		list = append(list, a.values[i].Raw())
	}
	return list
}

func (a Array) Len() int {
	return len(a.values)
}

func (a Array) True() bool {
	return len(a.values) > 0
}

func (a Array) Not() (Primitive, error) {
	return CreateBool(!a.True()), nil
}

func (a Array) Set(ix, value Primitive) (Primitive, error) {
	x, err := a.getIndex(ix)
	if err != nil {
		return nil, err
	}
	a.values[x] = value
	return a, nil
}

func (a Array) Get(ix Primitive) (Primitive, error) {
	x, err := a.getIndex(ix)
	if err != nil {
		return nil, err
	}
	return a.values[x], nil
}

func (a Array) getIndex(ix Primitive) (int, error) {
	var x int
	switch p := ix.(type) {
	case Int:
		x = int(p.value)
	case Float:
		x = int(p.value)
	default:
		return x, fmt.Errorf("%T can not be used as index")
	}
	if x < 0 {
		x = len(a.values) + x
	}
	if x < 0 || x >= len(a.values) {
		return x, fmt.Errorf("index out of range")
	}
	return x, nil
}

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

func (a Array) Iter(do func(Primitive) error) error {
	var err error
	for i := range a.values {
		if err = do(a.values[i]); err != nil {
			break
		}
	}
	return err
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

func (a Array) Rev() (Primitive, error) {
	return nil, unsupportedOp("reverse", a)
}

func (a Array) Add(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Array:
		a.values = append(a.values, x.values...)
	default:
		a.values = append(a.values, other)
	}
	return a, nil
}

func (a Array) Sub(other Primitive) (Primitive, error) {
	var offset int
	switch x := other.(type) {
	case Int:
		offset = int(x.value)
	case Float:
		offset = int(x.value)
	default:
		return nil, incompatibleType("multiply", a, other)
	}
	if offset > len(a.values) || -offset > len(a.values) {
		a.values = []Primitive{}
		return a, nil
	}
	if offset < 0 {
		a.values = a.values[-offset:]
	} else {
		a.values = a.values[:len(a.values)-offset]
	}
	return a, nil
}

func (a Array) Div(other Primitive) (Primitive, error) {
	var offset int
	switch x := other.(type) {
	case Int:
		offset = int(x.value)
	case Float:
		offset = int(x.value)
	default:
		return nil, incompatibleType("multiply", a, other)
	}
	if offset <= 0 {
		return nil, fmt.Errorf("array can not be divided negative values")
	}
	if offset > len(a.values) {
		return nil, fmt.Errorf("array can not be divided by %d", offset)
	}
	var (
		arr  Array
		size = len(a.values)
		step = size / offset
	)
	for i := 0; i < size && len(arr.values) < offset; i += step {
		end := i + step
		if end > size || len(arr.values) == offset-1 {
			end = size
		}
		sub := a.values[i:end]
		arr.values = append(arr.values, CreateArray(sub))
	}
	return arr, nil
}

func (a Array) Mul(other Primitive) (Primitive, error) {
	var offset int
	switch x := other.(type) {
	case Int:
		offset = int(x.value)
	case Float:
		offset = int(x.value)
	default:
		return nil, incompatibleType("multiply", a, other)
	}
	offset--

	vs := make([]Primitive, len(a.values))
	copy(vs, a.values)
	for i := 0; i < offset; i++ {
		a.values = append(a.values, vs...)
	}
	return a, nil
}

func (a Array) Pow(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("power", a)
}

func (a Array) Mod(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("modulo", a)
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
		return x, fmt.Errorf("%T can not be used as index", ix)
	}
	if x < 0 {
		x = len(a.values) + x
	}
	if x < 0 || x >= len(a.values) {
		return x, fmt.Errorf("index out of range")
	}
	return x, nil
}

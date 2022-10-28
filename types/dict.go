package types

import (
	"fmt"
	"strings"
)

type Dict struct {
	values map[Primitive]Primitive
}

func CreateDict() Primitive {
	return Dict{
		values: make(map[Primitive]Primitive),
	}
}

func (d Dict) String() string {
	var (
		str strings.Builder
		ix int
	)
	str.WriteString("{")
	for k, v := range d.values {
		if ix > 0 {
			str.WriteString(", ")
		}
		str.WriteString(k.String())
		str.WriteString(":")
		str.WriteString(v.String())
		ix++
	}
	str.WriteString("}")
	return str.String()
}

func (d Dict) Raw() any {
	n := make(map[Primitive]Primitive)
	for k, v := range d.values {
		n[k] = v
	}
	return n
}

func (d Dict) Iter(do func(Primitive) error) error {
	var err error
	for i := range d.values {
		if err = do(d.values[i]); err != nil {
			break
		}
	}
	return err
}

func (d Dict) Len() int {
	return len(d.values)
}

func (d Dict) True() bool {
	return len(d.values) > 0
}

func (d Dict) Not() (Primitive, error) {
	return CreateBool(!d.True()), nil
}

func (d Dict) Rev() (Primitive, error) {
	return nil, unsupportedOp("reverse", d)
}

func (d Dict) Add(other Primitive) (Primitive, error) {
	x, ok := other.(Dict)
	if !ok {
		return nil, incompatibleType("add", d, other)
	}
	n := make(map[Primitive]Primitive)
	for k, v := range d.values {
		n[k] = v
	}
	for k, v := range x.values {
		n[k] = v
	}
	return Dict{values: n}, nil
}

func (d Dict) Sub(other Primitive) (Primitive, error) {
	delete(d.values, other)
	return d, nil
}

func (d Dict) Mul(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("mulitply", d)
}

func (d Dict) Div(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("division", d)
}

func (d Dict) Mod(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("modulo", d)
}

func (d Dict) Pow(other Primitive) (Primitive, error) {
	return nil, unsupportedOp("power", d)
}

func (d Dict) Set(ix, value Primitive) (Primitive, error) {
	d.values[ix] = value
	return d, nil
}

func (d Dict) Get(ix Primitive) (Primitive, error) {
	p, ok := d.values[ix]
	if !ok {
		return nil, fmt.Errorf("%s: key not found", ix)
	}
	return p, nil
}

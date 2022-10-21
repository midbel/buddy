package types

type Array struct {
	values []Primitive
}

func (a Array) Len() int {
	return len(a.values)
}

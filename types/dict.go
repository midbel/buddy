package types

type Dict struct {
	values map[any]Primitive
}

func (d Dict) Len() int {
	return len(d.values)
}

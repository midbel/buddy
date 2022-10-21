package types

type Dict struct {
	values map[any]Primitive
}

func CreateDict() (Primitive, error) {
	return nil, nil
}

func (d Dict) String() string {
	return "dict"
}

func (d Dict) Raw() any {
	return nil
}

func (d Dict) Len() int {
	return len(d.values)
}

func (d Dict) Not() (Primitive, error) {
	return CreateBool(!d.True()), nil
}

func (d Dict) True() bool {
	return len(d.values) > 0
}

func (d Dict) Set(ix, value Primitive) (Primitive, error) {
	return nil, nil
}

func (d Dict) Get(ix Primitive) (Primitive, error) {
	return nil, nil
}

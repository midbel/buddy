package types

type Dict struct {
	values map[Primitive]Primitive
}

func CreateDict() Primitive {
	return Dict{
		values: make(map[Primitive]Primitive),
	}
}

func (d Dict) String() string {
	return "dict"
}

func (d Dict) Raw() any {
	return nil
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

func (d Dict) Not() (Primitive, error) {
	return CreateBool(!d.True()), nil
}

func (d Dict) True() bool {
	return len(d.values) > 0
}

func (d Dict) Set(ix, value Primitive) (Primitive, error) {
	d.values[ix] = value
	return d, nil
}

func (d Dict) Get(ix Primitive) (Primitive, error) {
	p, ok := d.values[ix]
	if !ok {

	}
	return p, nil
}

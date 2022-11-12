package types

import (
	"fmt"
)

type Context interface {
	Define(string, Primitive) error
	Assign(string, Primitive) error
	Resolve(string) (Primitive, error)

	// Call(string, string, []Argument) (Primitive, error)
}

type Argument struct {
	Name  string
	Index int
	Value Primitive
}

func PosArg(name string, index int) Argument {
	return NamedArg(name, index, nil)
}

func NamedArg(name string, index int, value Primitive) Argument {
	return Argument{
		Name:  name,
		Index: index,
		Value: value,
	}
}

type Callable interface {
	Call(Context, []Argument) (Primitive, error)
	Arity() int
}

type Module interface {
	Id() string
	Lookup(string, string) (Callable, error)
}

type Environ struct {
	parent *Environ
	values map[string]value
}

func EmptyEnv() *Environ {
	return EnclosedEnv(nil)
}

func EnclosedEnv(parent *Environ) *Environ {
	return &Environ{
		parent: parent,
		values: make(map[string]value),
	}
}

func (e *Environ) Resolve(name string) (Primitive, error) {
	v, ok := e.values[name]
	if !ok {
		if e.parent == nil {
			return nil, fmt.Errorf("%s undefined variable", name)
		}
		return e.parent.Resolve(name)
	}
	return v.value, nil
}

func (e *Environ) Assign(ident string, value Primitive) error {
	v, ok := e.values[ident]
	if !ok {
		if e.parent != nil {
			return e.parent.Assign(ident, value)
		}
		return fmt.Errorf("%s: variable not declared", ident)
	}
	if ok && v.readonly {
		return fmt.Errorf("%s readonly value", ident)
	}
	v.value = value
	e.values[ident] = v
	return nil
}

func (e *Environ) Define(ident string, value Primitive) error {
	v, ok := e.values[ident]
	if ok {
		return fmt.Errorf("%s: variable already defined", ident)
	}
	v.value = value
	e.values[ident] = v
	return nil
}

func (e *Environ) Exists(ident string) bool {
	v, ok := e.values[ident]
	return ok && v.value != nil
}

func (e *Environ) Wrap() *Environ {
	return EnclosedEnv(e)
}

func (e *Environ) Unwrap() *Environ {
	if e.parent == nil {
		return e
	}
	return e.parent
}

type value struct {
	value    Primitive
	readonly bool
}

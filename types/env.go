package types

import (
	"fmt"
)

type Context interface {
	Define(string, Primitive) error
	Resolve(string) (Primitive, error)

	Lookup(string, string) (Callable, error)
}

type Callable interface {
	Call(Context, ...Primitive) (Primitive, error)
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
		return nil, fmt.Errorf("%s undefined variable", name)
	}
	return v.value, nil
}

func (e *Environ) Define(name string, value Primitive) error {
	v, ok := e.values[name]
	if ok && v.readonly {
		return fmt.Errorf("%s readonly value", name)
	}
	v.value = value
	e.values[name] = v
	return nil
}

type value struct {
	value    Primitive
	readonly bool
}

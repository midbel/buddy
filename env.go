package buddy

import (
	"fmt"

	"github.com/midbel/buddy/builtins"
)

type Resolver interface {
	Define(string, any)
	Resolve(string) (any, error)
	Lookup(string) (Callable, error)

	getSymbols() map[string]Expression
}

type resolver struct {
	*Environ[any]
	symbols map[string]Expression
}

func (r resolver) Lookup(name string) (Callable, error) {
	b, err := builtins.Lookup(name)
	if err == nil {
		return makeCallFromFunc(b), err
	}
	e, ok := r.symbols[name]
	if !ok {
		return nil, fmt.Errorf("%s undefined function", name)
	}
	return makeCallFromExpr(e)
}

func (r resolver) getSymbols() map[string]Expression {
	return r.symbols
}

type Environ[T any] struct {
	parent *Environ[T]
	values map[string]T
}

func EmptyEnv[T any]() *Environ[T] {
	return EnclosedEnv[T](nil)
}

func EnclosedEnv[T any](parent *Environ[T]) *Environ[T] {
	return &Environ[T]{
		parent: parent,
		values: make(map[string]T),
	}
}

func (e *Environ[T]) Resolve(name string) (T, error) {
	var zero T
	v, ok := e.values[name]
	if !ok {
		return zero, fmt.Errorf("%s undefined variable")
	}
	return v, nil
}

func (e *Environ[T]) Define(name string, values T) {
	e.values[name] = values
}

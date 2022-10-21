package buddy

import (
	"fmt"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/types"
)

const LimitRecurse = 1 << 10

type Resolver struct {
	level int
	*Environ
	symbols map[string]Expression
}

func NewResolver() *Resolver {
	return ResolveEnv(EmptyEnv())
}

func ResolveEnv(env *Environ) *Resolver {
	return &Resolver{
		Environ: env,
		symbols: make(map[string]Expression),
	}
}

func (r *Resolver) Lookup(name string) (Callable, error) {
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

func (r *Resolver) enter() error {
	if r.level >= LimitRecurse {
		return fmt.Errorf("recursion limit reached!")
	}
	r.level++
	return nil
}

func (r *Resolver) leave() {
	r.level--
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

func (e *Environ) Resolve(name string) (types.Primitive, error) {
	v, ok := e.values[name]
	if !ok {
		return nil, fmt.Errorf("%s undefined variable", name)
	}
	return v.value, nil
}

func (e *Environ) Define(name string, value types.Primitive) error {
	v, ok := e.values[name]
	if ok && v.readonly {
		return fmt.Errorf("%s readonly value", name)
	}
	v.value = value
	e.values[name] = v
	return nil
}

type value struct {
	value    types.Primitive
	readonly bool
}

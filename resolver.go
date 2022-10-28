package buddy

import (
	"fmt"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/types"
)

var modPaths = []string{".", "./modules/"}

const LimitRecurse = 1 << 10

type Resolver struct {
	*types.Environ

	paths   []string
	level   int
	symbols map[string]Expression
}

func NewResolver() *Resolver {
	return ResolveEnv(types.EmptyEnv())
}

func ResolveEnv(env *types.Environ) *Resolver {
	return &Resolver{
		paths:   append([]string{}, modPaths...),
		Environ: env,
		symbols: make(map[string]Expression),
	}
}

func (r *Resolver) Lookup(name string) (Callable, error) {
	e, ok := r.symbols[name]
	if ok {
		return callableFromExpression(e)
	}
	b, err := builtins.LookupBuiltin(name)
	if err == nil {
		return callableFromBuiltin(b), nil
	}
	return nil, err
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

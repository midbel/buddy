package buddy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var modPaths = []string{".", "./modules/"}

const LimitRecurse = 1 << 10

type Resolver struct {
	level int
	*types.Environ
	symbols map[string]Expression
	modules map[string]*Resolver
}

func NewResolver() *Resolver {
	return ResolveEnv(types.EmptyEnv())
}

func ResolveEnv(env *types.Environ) *Resolver {
	return &Resolver{
		Environ: env,
		symbols: make(map[string]Expression),
		modules: make(map[string]*Resolver),
	}
}

func (r *Resolver) Load(name []string, alias string, symbols map[string]string) error {
	if len(name) == 0 {
		return fmt.Errorf("empty module name given")
	}
	if alias == "" {
		alias = slices.Lst(name)
	}
	return r.loadModule(name, alias, symbols)
}

func (r *Resolver) Find(name string) (*Resolver, error) {
	mod, ok := r.modules[name]
	if !ok {
		return nil, fmt.Errorf("%s: module not defined", name)
	}
	return mod, nil
}

func (r *Resolver) loadModule(name []string, alias string, symbols map[string]string) error {
	if _, ok := r.modules[alias]; ok {
		return nil
	}
	tryLoad := func(file string) (Expression, error) {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		return Parse(f)
	}

	var (
		file = filepath.Join(name...) + ".bud"
		expr Expression
		err  error
	)
	for _, dir := range modPaths {
		expr, err = tryLoad(filepath.Join(dir, file))
		if err == nil {
			break
		}
	}
	if err != nil || expr == nil {
		return fmt.Errorf("fail to load module from %s", file)
	}

	sub := ResolveEnv(types.EmptyEnv())
	sub.level = r.level
	if s, ok := expr.(script); ok {
		sub.symbols = s.symbols
	}
	_, err = execute(expr, sub)
	if err != nil {
		return err
	}
	if len(symbols) == 0 {
		r.modules[alias] = sub
		return nil
	}
	for ident, alias := range symbols {
		expr, ok := sub.symbols[ident]
		if !ok {
			return fmt.Errorf("%s: function not defined", ident)
		}
		r.symbols[alias] = expr
	}
	return nil
}

func (r *Resolver) loadBuiltin(name string) error {
	return nil
}

func (r *Resolver) Lookup(name string) (Callable, error) {
	if e, ok := r.symbols[name]; ok {
		return makeCallFromExpr(e)
	}
	b, err := builtins.Lookup(name)
	if err != nil {
		return nil, err
	}
	return makeCallFromBuiltin(b), nil
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

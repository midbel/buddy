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

type resolvedExpr struct {
	Expression
	resolv *Resolver
}

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
	r.modules[alias] = sub
	for id, a := range symbols {
		_, ok := sub.symbols[id]
		if !ok {
			return fmt.Errorf("%s: function not defined", id)
		}
		r.symbols[a] = createAlias(alias, id)
	}
	return nil
}

func (r *Resolver) Lookup(name string) (Callable, error) {
	if e, ok := r.symbols[name]; ok {
		if a, ok := e.(alias); ok {
			sub, err := r.Find(a.module)
			if err != nil {
				return nil, err
			}
			return sub.Lookup(a.ident)
		}
		return makeCallFromExpr(r, e)
	}
	b, err := builtins.Lookup(name)
	if err != nil {
		return nil, err
	}
	return makeCallFromBuiltin(r, b), nil
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

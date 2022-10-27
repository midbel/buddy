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

type Module interface {
	Lookup(string) (Callable, error)
}

type link struct {
	Expression
	Module
}

type builtinModule struct {
	module builtins.Module
}

func moduleFromBuiltin(mod builtins.Module) Module {
	return builtinModule{
		module: mod,
	}
}

func (m builtinModule) Lookup(name string) (Callable, error) {
	b, err := m.module.Lookup(name)
	if err != nil {
		return nil, err
	}
	return callableFromBuiltin(b), nil
}

type userDefinedModule map[string]Expression

func moduleFromSymbols(symbols map[string]Expression) Module {
	return userDefinedModule(symbols)
}

func emptyModule() Module {
	return make(userDefinedModule)
}

func (m userDefinedModule) Lookup(name string) (Callable, error) {
	return callableFromExpression(m[name])
}

type Resolver struct {
	*types.Environ

	paths   []string
	level   int
	current Module
	modules map[string]Module
}

func NewResolver() *Resolver {
	return ResolveEnv(types.EmptyEnv())
}

func ResolveEnv(env *types.Environ) *Resolver {
	return &Resolver{
		paths:   append([]string{}, modPaths...),
		Environ: env,
		current: emptyModule(),
		modules: make(map[string]Module),
	}
}

func (r *Resolver) Sub(env *types.Environ) *Resolver {
	sub := ResolveEnv(env)
	sub.level = r.level
	sub.current = r.current
	sub.modules = r.modules
	sub.paths = append(sub.paths[:0], r.paths...)
	return sub
}

func (r *Resolver) Lookup(name string) (Callable, error) {
	c, err := r.current.Lookup(name)
	if err == nil {
		return c, err
	}
	b, err := builtins.LookupBuiltin(name)
	if err == nil {
		return callableFromBuiltin(b), nil
	}
	return nil, err
}

func (r *Resolver) Find(name string) (*Resolver, error) {
	mod, ok := r.modules[name]
	if !ok {
		return nil, fmt.Errorf("%s: module not defined", name)
	}
	sub := NewResolver()
	sub.current = mod
	return sub, nil
}

func (r *Resolver) Load(name []string, alias string, symbols map[string]string) error {
	if len(name) == 0 {
		return fmt.Errorf("empty module name given")
	}
	mod := slices.Lst(name)
	if alias == "" {
		alias = mod
	}
	if m, err := builtins.LookupModule(mod); err == nil {
		r.modules[alias] = moduleFromBuiltin(m)
		return nil
	}
	return r.loadUserDefinedModule(name, alias, symbols)
}

func (r *Resolver) loadUserDefinedModule(name []string, alias string, symbols map[string]string) error {
	var (
		file = filepath.Join(name...) + ".bud"
		list map[string]Expression
		err  error
	)
	for i := range r.paths {
		list, err = tryLoad(filepath.Join(r.paths[i], file))
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("fail to load module from %s", file)
	}
	mod := moduleFromSymbols(list)
	if len(symbols) == 0 {
		r.modules[alias] = mod
		return nil
	}
	for n, a := range symbols {
		expr, ok := list[n]
		if !ok {
			return fmt.Errorf("%s can not be imported", n)
		}
		userdef, ok := r.current.(userDefinedModule)
		if !ok {
			return fmt.Errorf("module can not be modified")
		}
		userdef[a] = link{
			Expression: expr,
			Module: mod,
		}
		fmt.Printf("%s => %s => %T\n", n, a, expr)
	}
	return nil
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

func tryLoad(file string) (map[string]Expression, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	expr, err := Parse(f)
	if err != nil {
		return nil, err
	}
	s, ok := expr.(script)
	if !ok {
		return nil, fmt.Errorf("parse does not return a script")
	}
	return s.symbols, nil
}

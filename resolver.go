package buddy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var ImportPaths = []string{".", "./modules/"}

const LimitRecurse = 1 << 10

const (
	buddyPath    = "BUDDYPATH"
	buddyRecurse = "BUDDYRECURSELIMIT"
)

type Resolver struct {
	*types.Environ

	name    string
	paths   []string
	level   int
	symbols map[string]Expression
	modules map[string]Module
}

func NewResolver() *Resolver {
	return ResolveEnv(types.EmptyEnv())
}

func ResolveEnv(env *types.Environ) *Resolver {
	paths := filepath.SplitList(os.Getenv(buddyPath))
	if len(paths) == 0 {
		paths = append(paths, ImportPaths...)
	}
	return &Resolver{
		name:    "main",
		paths:   paths,
		Environ: env,
		symbols: make(map[string]Expression),
		modules: make(map[string]Module),
	}
}

func (r *Resolver) Load(names []string, alias string, symbols map[string]string) error {
	if mod, err := builtins.LookupModule(slices.Lst(names)); err == nil {
		if len(symbols) == 0 {
			r.modules[alias] = moduleFromBuiltin(mod)
			return nil
		}
		for ident, alias := range symbols {
			call, err := mod.Lookup(ident)
			if err != nil {
				return err
			}
			r.symbols[alias] = createLink(callableFromBuiltin(call))
		}
		return nil
	}
	mod, err := loadModule(names, r.paths)
	if err != nil {
		return err
	}
	if len(symbols) == 0 {
		r.modules[alias] = mod
		return nil
	}
	for ident, alias := range symbols {
		call, err := mod.Lookup(ident)
		if err != nil {
			return err
		}
		r.symbols[alias] = createLink(call)
	}
	return nil
}

func (r *Resolver) Lookup(name string) (Callable, error) {
	return r.LookupCallable(name, "")
}

func (r *Resolver) LookupCallable(name, module string) (Callable, error) {
	if module != "" {
		mod, ok := r.modules[module]
		if !ok {
			return nil, fmt.Errorf("%s: module not defined", module)
		}
		return mod.Lookup(name)
	}
	e, ok := r.symbols[name]
	if ok {
		return callableFromExpression(r, e)
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

func loadModule(names, paths []string) (Module, error) {
	try := func(file string) (Module, error) {
		r, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer r.Close()

		expr, err := Parse(r)
		if err != nil {
			return nil, err
		}
		res := NewResolver()
		_, err = execute(expr, res)
		if err == nil {
			res.name = slices.Lst(names)
		}
		return res, err
	}
	var (
		mod  Module
		err  error
		file = filepath.Join(names...) + ".bud"
	)
	for _, p := range paths {
		mod, err = try(filepath.Join(p, file))
		if err == nil {
			return mod, nil
		}
	}
	if mod == nil {
		err = fmt.Errorf("no module loaded from %s", strings.Join(names, "."))
	}
	return nil, err
}

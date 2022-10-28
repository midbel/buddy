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
		paths:   paths,
		Environ: env,
		symbols: make(map[string]Expression),
		modules: make(map[string]Module),
	}
}

func (r *Resolver) Load(names []string, alias string) error {
	if mod, err := builtins.LookupModule(slices.Lst(names)); err == nil {
		r.modules[alias] = moduleFromBuiltin(mod)
		return err
	}
	mod, err := loadModule(names, r.paths)
	if err != nil {
		return err
	}
	r.modules[alias] = mod
	return nil
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

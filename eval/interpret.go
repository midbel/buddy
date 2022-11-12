package eval

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/parse"
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

const LimitDepth = 1024

type CallFunc func(call types.Callable) (types.Primitive, error)

type Interpreter struct {
	Stdout io.Writer
	Stderr io.Writer
	Stdin  io.Reader

	ImportPats []string
	MaxDepth   int
	currDepth  int

	stack   *slices.Stack[types.Module]
	modules []types.Module
	*types.Environ
}

func Default() *Interpreter {
	return New(types.EmptyEnv())
}

func New(env *types.Environ) *Interpreter {
	i := Interpreter{
		Stdout:   os.Stdout,
		Stderr:   os.Stderr,
		Stdin:    os.Stdin,
		Environ:  env,
		MaxDepth: LimitDepth,
		stack:    slices.New[types.Module](),
	}
	mod := emptyModule("main")
	i.stack.Push(mod)
	return &i
}

func (i *Interpreter) Eval(r io.Reader) (types.Primitive, error) {
	expr, err := parse.New(r).Parse()
	if err != nil {
		return nil, err
	}
	return eval(expr, i)
}

func (i *Interpreter) EvalString(str string) (types.Primitive, error) {
	return i.Eval(strings.NewReader(str))
}

func (i *Interpreter) Load(ident []string, alias string) error {
	if mod, err := builtins.LookupModule(slices.Lst(ident)); err == nil {
		tmp := i.stack.Top()
		if reg, ok := tmp.(mutableModule); !ok {
			return fmt.Errorf("builtin module can not be imported")
		} else {
			reg.Register(alias, mod)
		}
		return nil
	}
	r, err := os.Open(filepath.Join(ident...) + ".bud")
	if err != nil {
		return err
	}
	defer r.Close()

	expr, err := parse.New(r).Parse()
	if err != nil {
		return err
	}

	mod := emptyModule(slices.Lst(ident))
	i.stack.Push(mod)

	if _, err := eval(expr, i); err != nil {
		return err
	}
	i.stack.Pop()

	s, ok := expr.(ast.Script)
	if !ok {
		return fmt.Errorf("fail to load module from %s", strings.Join(ident, "."))
	}
	for ident, expr := range s.Symbols {
		call, err := callableFromExpression(expr)
		if err != nil {
			return err
		}
		mod.Append(ident, call)
	}

	tmp := i.stack.Top()
	if reg, ok := tmp.(mutableModule); !ok {
		return fmt.Errorf("user module can not be imported")
	} else {
		reg.Register(alias, mod)
	}
	return nil
}

func (i *Interpreter) Call(mod, ident string, call CallFunc) (types.Primitive, error) {
	if err := i.enter(); err != nil {
		return nil, err
	}
	defer i.leave()

	var (
		m   types.Module
		err error
	)
	if mod == "" {
		m = i.stack.Top()
	} else {
		m, err = i.lookupModule(mod)
		if err != nil {
			return nil, err
		}
		i.stack.Push(m)
		defer i.stack.Pop()
	}
	fn, err := m.Lookup("", ident)
	if err == nil {
		return call(fn)
	}
	fn, err = builtins.LookupBuiltin(ident)
	if err != nil {
		return nil, err
	}
	return call(fn)
}

func (i *Interpreter) Lookup(mod, ident string) (types.Callable, error) {
	if mod == "" {
		return i.stack.Top().Lookup("", ident)
	}
	m, err := i.lookupModule(mod)
	if err != nil {
		return nil, err
	}
	return m.Lookup("", ident)
}

func (i *Interpreter) lookupModule(ident string) (types.Module, error) {
	var (
		curr    = i.stack.Top()
		get, ok = curr.(mutableModule)
	)
	if !ok {
		return nil, fmt.Errorf("%s: fail to find module from %s", ident, curr.Id())
	}
	return get.Get(ident)
}

func (i *Interpreter) enterScope() {
	i.Environ = i.Environ.Wrap()
}

func (i *Interpreter) leaveScope() {
	i.Environ = i.Environ.Unwrap()
}

func (i *Interpreter) enter() error {
	if i.currDepth >= i.MaxDepth {
		return fmt.Errorf("max call stacked reached!")
	}
	i.currDepth++
	return nil
}

func (i *Interpreter) leave() {
	i.currDepth--
}

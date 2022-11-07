package builtins

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var ErrType = errors.New("type error")

type Module struct {
	Name     string
	Builtins map[string]Builtin
}

func (m Module) Id() string {
	return m.Name
}

func (m Module) Filter(names map[string]string) (Module, error) {
	if len(names) == 0 {
		return m, nil
	}
	bs := make(map[string]Builtin)
	for n, a := range names {
		b, ok := m.Builtins[n]
		if !ok {
			return m, fmt.Errorf("%s: undefined function", n)
		}
		bs[a] = b
	}
	mod := Module{
		Name:     m.Name,
		Builtins: bs,
	}
	return mod, nil
}

func (m Module) Lookup(mod, name string) (types.Callable, error) {
	if mod != "" {
		return nil, fmt.Errorf("%s: no sub module defined", name)
	}
	b, ok := m.Builtins[name]
	if !ok {
		return b, fmt.Errorf("%s: function not defined", name)
	}
	return b, nil
}

type BuiltinFunc func(...types.Primitive) (types.Primitive, error)

type Parameter struct {
	Name  string
	Value types.Primitive
}

func createPositional(name string) Parameter {
	return Parameter{
		Name: name,
	}
}

func createNamed(name string, value types.Primitive) Parameter {
	return Parameter{
		Name:  name,
		Value: value,
	}
}

type Builtin struct {
	Name     string
	Variadic bool
	Params   []Parameter
	Run      BuiltinFunc
}

func (b Builtin) Arity() int {
	return len(b.Params)
}

func (b Builtin) Call(_ types.Context, args ...types.Primitive) (types.Primitive, error) {
	if b.Call == nil {
		return nil, fmt.Errorf("%s can not be called", b.Name)
	}
	if len(args) != len(b.Params) {
		if b.Variadic && len(args) < len(b.Params) {
			return nil, fmt.Errorf("%s: not enough argument given", b.Name)
		}
	}
	res, err := b.Run(args...)
	if err != nil {
		err = fmt.Errorf("%s: %w", b.Name, err)
	}
	return res, err
}

var Modules = []Module{
	iomod,
	strmod,
	defmod,
	arrmod,
	timemod,
}

var defmod = Module{
	Name: "builtin",
	Builtins: map[string]Builtin{
		"int": {
			Name: "int",
			Params: []Parameter{
				createPositional("value"),
			},
			Run: runInt,
		},
		"float": {
			Name: "runFloat",
			Params: []Parameter{
				createPositional("value"),
			},
			Run: runFloat,
		},
		"string": {
			Name: "string",
			Params: []Parameter{
				createPositional("value"),
			},
			Run: runString,
		},
		"bool": {
			Name: "string",
			Params: []Parameter{
				createPositional("value"),
			},
			Run: runBool,
		},
		"len": {
			Name: "len",
			Params: []Parameter{
				createPositional("value"),
			},
			Run: runLen,
		},
		"exit": {
			Name: "exit",
			Params: []Parameter{
				createPositional("code"),
			},
			Run: runExit,
		},
		"all": {
			Name:     "all",
			Variadic: true,
			Run:      runAll,
		},
		"any": {
			Name:     "any",
			Variadic: true,
			Run:      runAny,
		},
		"typeof": {
			Name: "typeof",
			Params: []Parameter{
				createPositional("value"),
			},
			Run: runTypeof,
		},
		"dir": {
			Name:     "dir",
			Variadic: true,
			Run:      nil,
		},
	},
}

func LookupModule(name string) (Module, error) {
	sort.Slice(Modules, func(i, j int) bool {
		return Modules[i].Name <= Modules[j].Name
	})
	i := sort.Search(len(Modules), func(i int) bool {
		return Modules[i].Name >= name
	})
	if i < len(Modules) && Modules[i].Name == name {
		return Modules[i], nil
	}
	return Module{}, fmt.Errorf("%s: undefined module", name)
}

func LookupBuiltin(ident string) (types.Callable, error) {
	return defmod.Lookup("", ident)
}

func runTypeof(args ...types.Primitive) (types.Primitive, error) {
	name, err := types.Type(slices.Fst(args))
	if err != nil {
		return nil, err
	}
	return types.CreateString(name), nil
}

func runAll(args ...types.Primitive) (types.Primitive, error) {
	var ok bool
	for _, a := range args {
		ok = a.True()
		if !ok {
			break
		}
	}
	return types.CreateBool(ok), nil
}

func runAny(args ...types.Primitive) (types.Primitive, error) {
	var ok bool
	for _, a := range args {
		ok = a.True()
		if ok {
			break
		}
	}
	return types.CreateBool(ok), nil
}

func runInt(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}
	var (
		raw = slices.Fst(args).Raw()
		val int64
	)
	switch v := raw.(type) {
	case int64:
		val = v
	case float64:
		val = int64(v)
	case string:
		x, err := strconv.ParseInt(v, 0, 64)
		if err != nil {
			return nil, err
		}
		val = x
	case bool:
		val = 0
		if v {
			val = 1
		}
	default:
		return nil, fmt.Errorf("%T can not be casted to integer", val)
	}
	return types.CreateInt(val), nil
}

func runFloat(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}
	var (
		raw = slices.Fst(args).Raw()
		val float64
	)
	switch v := raw.(type) {
	case int64:
		val = float64(v)
	case float64:
		val = v
	case string:
		x, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		val = x
	case bool:
		val = 0
		if v {
			val = 1
		}
	default:
		return nil, fmt.Errorf("%T can not be casted to float", val)
	}
	return types.CreateFloat(val), nil
}

func runString(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}
	str := slices.Fst(args).String()
	return types.CreateString(str), nil
}

func runBool(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}
	return types.CreateBool(slices.Fst(args).True()), nil
}

func runLen(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len: no enough argument given")
	}
	siz, ok := slices.Fst(args).(types.Sizeable)
	if !ok {
		return nil, fmt.Errorf("can not get length from %s", siz)
	}
	return types.CreateInt(int64(siz.Len())), nil
}

func typeError(got, want types.Primitive) error {
	src, err := types.Type(got)
	if err != nil {
		src = "unknown"
	}
	dst, err := types.Type(want)
	if err != nil {
		dst = "unknown"
	}
	return fmt.Errorf("%w: expected %s, got %s", ErrType, dst, src)
}

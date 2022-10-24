package builtins

import (
	"fmt"
	"strconv"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

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
	Call     BuiltinFunc
}

func (b Builtin) Run(args ...types.Primitive) (types.Primitive, error) {
	res, err := b.Call(args...)
	if err != nil {
		err = fmt.Errorf("%s: %w", b.Name, err)
	}
	return res, err
}

var Builtins = map[string]Builtin{
	"int": {
		Name: "int",
		Params: []Parameter{
			createPositional("value"),
		},
		Call: runInt,
	},
	"float": {
		Name: "runFloat",
		Params: []Parameter{
			createPositional("value"),
		},
		Call: runFloat,
	},
	"string": {
		Name: "string",
		Params: []Parameter{
			createPositional("value"),
		},
		Call: runString,
	},
	"bool": {
		Name: "string",
		Params: []Parameter{
			createPositional("value"),
		},
		Call: runBool,
	},
	"len": {
		Name: "len",
		Params: []Parameter{
			createPositional("value"),
		},
		Call: runLen,
	},
	"exit": {
		Name: "exit",
		Params: []Parameter{
			createPositional("code"),
		},
		Call: runExit,
	},
	"lower": {
		Name: "lower",
		Params: []Parameter{
			createPositional("str"),
		},
		Call: runLower,
	},
	"upper": {
		Name: "upper",
		Params: []Parameter{
			createPositional("str"),
		},
		Call: runUpper,
	},
	"print": {
		Name:     "print",
		Variadic: true,
		Call:     runPrint,
	},
	"printf": {
		Name:     "printf",
		Variadic: true,
		Params: []Parameter{
			createPositional("format"),
		},
		Call: runPrintf,
	},
}

func Lookup(name string) (Builtin, error) {
	b, ok := Builtins[name]
	if !ok {
		return b, fmt.Errorf("%s: builtin not defined", name)
	}
	return b, nil
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

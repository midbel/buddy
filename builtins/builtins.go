package builtins

import (
	"fmt"
	"strconv"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

type BuiltinFunc func(...types.Primitive) (types.Primitive, error)

var Builtins = map[string]BuiltinFunc{
	"len":    Len,
	"upper":  Upper,
	"lower":  Lower,
	"printf": Printf,
	"print":  Print,
	"lshift": Lshift,
	"rshift": Rshift,
	"incr":   Incr,
	"decr":   Decr,
	"exit":   Exit,
	"int":    Int,
	"float":  Float,
	"string": String,
	"bool":   Bool,
}

func Lookup(name string) (BuiltinFunc, error) {
	b, ok := Builtins[name]
	if !ok {
		return nil, fmt.Errorf("%s: builtin not defined", name)
	}
	return b, nil
}

func Int(args ...types.Primitive) (types.Primitive, error) {
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

func Float(args ...types.Primitive) (types.Primitive, error) {
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

func String(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}
	str := slices.Fst(args).String()
	return types.CreateString(str), nil
}

func Bool(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("invalid number of arguments")
	}
	return types.CreateBool(slices.Fst(args).True()), nil
}

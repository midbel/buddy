package builtins

import (
	"fmt"
)

type BuiltinFunc func(...any) (any, error)

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
}

func Lookup(name string) (BuiltinFunc, error) {
	b, ok := Builtins[name]
	if !ok {
		return nil, fmt.Errorf("%s: builtin not defined", name)
	}
	return b, nil
}

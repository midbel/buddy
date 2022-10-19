package builtins

import (
	"fmt"
)

type Builtin func(...any) (any, error)

var Builtins = map[string]Builtin{
	"len":    Len,
	"upper":  Upper,
	"lower":  Lower,
	"printf": Printf,
	"print":  Print,
}

func Lookup(name string) (Builtin, error) {
	b, ok := Builtins[name]
	if !ok {
		return nil, fmt.Errorf("%s: builtin not defined", name)
	}
	return b, nil
}

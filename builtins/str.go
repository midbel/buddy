package builtins

import (
	"fmt"
	"strings"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var strmod = Module{
	Name: "strings",
	Builtins: map[string]Builtin{
		"upper": {
			Name: "upper",
			Params: []Parameter{
				createPositional("str"),
			},
			Call: runUpper,
		},
		"lower": {
			Name: "lower",
			Params: []Parameter{
				createPositional("str"),
			},
			Call: runLower,
		},
		"index": {
			Name: "index",
			Params: []Parameter{
				createPositional("str"),
				createPositional("search"),
			},
			Call: nil,
		},
		"substr": {
			Name: "substr",
			Params: []Parameter{
				createPositional("str"),
				createPositional("index"),
			},
			Call: nil,
		},
	},
}

func runLower(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	str, ok := slices.Fst(args).Raw().(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	str = strings.ToLower(str)
	return types.CreateString(str), nil
}

func runUpper(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	str, ok := slices.Fst(args).Raw().(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	str = strings.ToUpper(str)
	return types.CreateString(str), nil
}

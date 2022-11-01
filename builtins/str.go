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
		"format": {
			Name: "format",
			Variadic: true,
			Params: []Parameter{
				createPositional("pattern"),
			},
			Call: runFormat,
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

func runFormat(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no enough argument given")
	}
	str, ok := slices.Fst(args).Raw().(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	var list []any
	for _, a := range slices.Rest(args) {
		list = append(list, a)
	}
	str = fmt.Sprintf(str, list...)
	return types.CreateString(str), nil
}

func runLower(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no enough argument given")
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
		return nil, fmt.Errorf("no enough argument given")
	}
	str, ok := slices.Fst(args).Raw().(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	str = strings.ToUpper(str)
	return types.CreateString(str), nil
}

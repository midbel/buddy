package builtins

import (
	"fmt"
	"os"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var iomod = Module{
	Name: "io",
	Builtins: map[string]Builtin{
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
	},
}

func runPrintf(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no enough argument given")
	}
	pattern, ok := slices.Fst(args).Raw().(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	var list []any
	for _, a := range slices.Rest(args) {
		list = append(list, a.Raw())
	}
	fmt.Fprintf(os.Stdout, pattern, list...)
	return nil, nil
}

func runPrint(args ...types.Primitive) (types.Primitive, error) {
	var list []any
	for i := range args {
		list = append(list, args[i].Raw())
	}
	fmt.Fprintln(os.Stdout, list...)
	return nil, nil
}

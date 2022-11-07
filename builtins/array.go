package builtins

import (
	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var arrmod = Module{
	Name: "array",
	Builtins: map[string]Builtin{
		"first": {
			Name: "first",
			Params: []types.Argument{
				types.PosArg("array", 1),
			},
			Run: runFirst,
		},
		"last": {
			Name: "last",
			Params: []types.Argument{
				types.PosArg("array", 1),
			},
			Run: runLast,
		},
	},
}

func runFirst(args ...types.Primitive) (types.Primitive, error) {
	arr, ok := slices.Fst(args).(types.Array)
	if !ok {
		return nil, typeError(slices.Fst(args), arr)
	}
	return arr.Get(types.CreateInt(0))
}

func runLast(args ...types.Primitive) (types.Primitive, error) {
	arr, ok := slices.Fst(args).(types.Array)
	if !ok {
		return nil, typeError(slices.Fst(args), arr)
	}
	x := arr.Len() - 1
	return arr.Get(types.CreateInt(int64(x)))
}

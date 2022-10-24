package builtins

import (
	"fmt"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

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

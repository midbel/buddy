package builtins

import (
	"fmt"

	// "github.com/midbel/slices"
	"github.com/midbel/buddy/types"
)

func Len(args ...types.Primitive) (types.Primitive, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("len: no enough argument given")
	}
	// str, ok := slices.Fst(args).(string)
	// if !ok {
	// 	return nil, fmt.Errorf("incompatible type: string expected")
	// }
	// return float64(len(str)), nil
	return types.CreateInt(0), nil
}

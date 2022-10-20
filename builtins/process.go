package builtins

import (
	"errors"
	// "fmt"

	// "github.com/midbel/slices"
	"github.com/midbel/buddy/types"
)

var ErrExit = errors.New("exit")

func IsExit(err error) bool {
	return errors.Is(err, ErrExit)
}

func Exit(args ...types.Primitive) (types.Primitive, error) {
	// if len(args) == 0 {
	// 	return 0, ErrExit
	// }
	// if len(args) != 1 {
	// 	return nil, fmt.Errorf("exit: not enough argument give")
	// }
	// code, ok := slices.Fst(args).(float64)
	// if !ok {
	// 	return nil, fmt.Errorf("number expected, got %T", slices.Fst(args))
	// }
	// return code, ErrExit
	return types.CreateInt(0), ErrExit
}

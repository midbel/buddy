package builtins

import (
	"errors"
	"fmt"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

var ErrExit = errors.New("exit")

func IsExit(err error) bool {
	return errors.Is(err, ErrExit)
}

func Exit(args ...types.Primitive) (types.Primitive, error) {
	if len(args) == 0 {
		return types.CreateInt(0), ErrExit
	}
	if len(args) != 1 {
		return nil, fmt.Errorf("exit: not enough argument give")
	}
	var code int64
	switch x := slices.Fst(args).Raw().(type) {
	case int64:
		code = x
	case float64:
		code = int64(x)
	default:
		return nil, fmt.Errorf("number expected, got %T", slices.Fst(args))
	}
	return types.CreateInt(code), ErrExit
}

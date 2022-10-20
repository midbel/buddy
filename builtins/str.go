package builtins

import (
	"fmt"
	"strings"

	"github.com/midbel/buddy/types"
	"github.com/midbel/slices"
)

func Lower(args ...types.Primitive) (types.Primitive, error) {
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

func Upper(args ...types.Primitive) (types.Primitive, error) {
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

func Printf(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	pattern, ok := slices.Fst(args).Raw().(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	var list []any
	for _, a := range slices.Rest(args) {
		list = append(list, a.Raw())
	}
	str := fmt.Sprintf(pattern, list...)
	return types.CreateString(str), nil
}

func Print(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	var list []any
	for i := range args {
		list = append(list, args[i])
	}
	str := fmt.Sprint(list...)
	return types.CreateString(str), nil
}

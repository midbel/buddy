package builtins

import (
	"fmt"
	// "strings"

	// "github.com/midbel/slices"
	"github.com/midbel/buddy/types"
)

func Lower(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	// str, ok := slices.Fst(args).(string)
	// if !ok {
	// 	return nil, fmt.Errorf("incompatible type: string expected")
	// }
	// return strings.ToLower(str), nil
	return nil, nil
}

func Upper(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	// str, ok := slices.Fst(args).(string)
	// if !ok {
	// 	return nil, fmt.Errorf("incompatible type: string expected")
	// }
	// return strings.ToUpper(str), nil
	return nil, nil
}

func Printf(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	// pattern, ok := slices.Fst(args).(string)
	// if !ok {
	// 	return nil, fmt.Errorf("incompatible type: string expected")
	// }
	// return fmt.Sprintf(pattern, slices.Rest(args)...), nil
	return nil, nil
}

func Print(args ...types.Primitive) (types.Primitive, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	// return fmt.Sprint(args...), nil
	return nil, nil
}

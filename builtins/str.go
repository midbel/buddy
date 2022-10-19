package builtins

import (
	"fmt"
	"strings"

	"github.com/midbel/slices"
)

func Lower(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	str, ok := slices.Fst(args).(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	return strings.ToLower(str), nil
}

func Upper(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	str, ok := slices.Fst(args).(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	return strings.ToUpper(str), nil
}

func Printf(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	pattern, ok := slices.Fst(args).(string)
	if !ok {
		return nil, fmt.Errorf("incompatible type: string expected")
	}
	return fmt.Sprintf(pattern, slices.Rest(args)...), nil
}

func Print(args ...any) (any, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("printf: no enough argument given")
	}
	return fmt.Sprint(args...), nil
}

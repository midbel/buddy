package builtins

import (
	"time"

	"github.com/midbel/buddy/types"
)

var timemod = Module{
	Name: "time",
	Builtins: map[string]Builtin{
		"now": {
			Name: "now",
			Run:  runNow,
		},
		"unix": {
			Name: "unix",
			Run:  runUnix,
		},
	},
}

func runUnix(args ...types.Primitive) (types.Primitive, error) {
	var (
		now  = time.Now()
		unix = now.Unix()
	)
	return types.CreateInt(unix), nil
}

func runNow(args ...types.Primitive) (types.Primitive, error) {
	var (
		now = time.Now()
		str = now.Format(time.RFC3339)
	)
	return types.CreateString(str), nil
}

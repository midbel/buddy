package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/parse"
	"github.com/midbel/buddy/visitors"
)

func main() {
	flag.Parse()

	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer r.Close()

	expr, err := parse.Parse(r)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	base := filepath.Base(flag.Arg(0))
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if s, ok := expr.(ast.Script); ok {
		for k := range s.Symbols {
			compute(fmt.Sprintf("%s.%s", base, k), s.Symbols[k])
		}
	}
	compute(base, expr)
}

func compute(ident string, expr ast.Expression) {
	cmp := visitors.Complexity()
	count, err := cmp.Count(expr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", ident, err)
		fmt.Fprintln(os.Stderr)
		return
	}
	fmt.Fprintf(os.Stdout, "%s: %3d", ident, count)
	fmt.Fprintln(os.Stdout)
}

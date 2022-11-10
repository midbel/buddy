package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
	"github.com/midbel/buddy/parse"
	"github.com/midbel/buddy/visitors"
)

func main() {
	lint := flag.Bool("l", false, "lint code")
	flag.Parse()
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	err = Debug(os.Stdout, r, *lint)
	if err != nil {
		faults.PrintError(os.Stderr, err)
		os.Exit(1)
	}
}

func Debug(w io.Writer, r io.Reader, lint bool) error {
	expr, err := parse.New(r).Parse()
	if err != nil {
		return err
	}
	if lint {
		all := []visitors.Visitor{
			visitors.Value(),
			visitors.Variable(),
			visitors.Import(),
		}
		expr, err = visitors.Visit(expr, all)
	}
	if err == nil {
		ast.Debug(w, expr)
	}
	return err
}

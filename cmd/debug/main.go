package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/midbel/buddy"
)

func main() {
	visit := flag.Bool("v", false, "modify generated AST")
	flag.Parse()
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	defer r.Close()

	err = buddy.Debug(os.Stdout, r, *visit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

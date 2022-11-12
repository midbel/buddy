package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/parse"
	"github.com/midbel/buddy/visitors"
	"github.com/midbel/slices"
)

func main() {
	var (
		avg = flag.Bool("avg", false, "show overall average")
		top = flag.Int("top", 0, "show the top N most complex functions")
	)
	flag.Parse()

	var list []Result
	for _, a := range flag.Args() {
		scores, err := processFile(a)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		list = append(list, scores...)
	}
	switch {
	case *avg:
		fmt.Fprintf(os.Stdout, "average complexty: %.0f", getAverage(list))
		fmt.Fprintln(os.Stdout)
	case *top > 0:
		list = getTop(list, *top)
		printResult(list)
	default:
		printResult(list)
	}
}

const (
	Low  = "low"
	Mod  = "moderate"
	High = "high"
	Risk = "unreasonable"
)

type Result struct {
	File  string
	Func  string
	Score float64
}

func (r Result) Label() string {
	switch {
	case r.Score < 10:
		return Low
	case r.Score >= 10 && r.Score < 20:
		return Mod
	case r.Score >= 20 && r.Score < 50:
		return High
	default:
		return Risk
	}
}

func Make(file, ident string, score int) Result {
	return Result{
		File:  file,
		Func:  ident,
		Score: float64(score),
	}
}

func printResult(list []Result) {
	for _, r := range list {
		fmt.Fprintf(os.Stdout, "%s: complexity of %s is %s (score: %.0f)", r.File, r.Func, r.Label(), r.Score)
		fmt.Fprintln(os.Stdout)
	}
}

func getTop(list []Result, n int) []Result {
	sort.Slice(list, func(i, j int) bool {
		return list[i].Score < list[j].Score
	})
	x := len(list)
	if n < x {
		list = list[x-n:]
	}
	return slices.Reverse(list)
}

func getAverage(list []Result) float64 {
	var (
		total = len(list)
		sum   float64
	)
	if total == 0 {
		return sum
	}
	for i := range list {
		sum += list[i].Score
	}
	return sum / float64(total)
}

func processFile(file string) ([]Result, error) {
	r, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	expr, err := parse.Parse(r)
	if err != nil {
		return nil, err
	}
	var (
		list []Result
		base = filepath.Base(file)
	)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if s, ok := expr.(ast.Script); ok {
		for k := range s.Symbols {
			score, err := compute(s.Symbols[k])
			if err != nil {
				return nil, err
			}
			list = append(list, Make(base, k, score))
		}
	}
	score, err := compute(expr)
	if err != nil {
		return nil, err
	}
	return append(list, Make(base, "main", score)), nil
}

func compute(expr ast.Expression) (int, error) {
	cmp := visitors.Complexity()
	return cmp.Count(expr)
}

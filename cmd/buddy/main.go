package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/midbel/buddy/builtins"
	"github.com/midbel/buddy/eval"
	"github.com/midbel/buddy/faults"
)

func main() {
	flag.Parse()
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		interactive(os.Stdin)
		return
	}
	defer r.Close()
	if err := execute(r); err != nil {
		os.Exit(1)
	}
}

func execute(r io.Reader) error {
	res, err := eval.Eval(r)
	if err != nil {
		faults.PrintError(os.Stderr, err)
		return err
	}
	if res != nil {
		fmt.Printf("%+v", res)
		fmt.Println()
	}
	return nil
}

const (
	in  = "\x1b[1;97min [%3d]:\x1b[0m "
	ok  = "\x1b[1;92mout[%3d]:\x1b[0m %v"
	nok = "\x1b[1;91mout[%3d]:\x1b[0m %s"
)

func interactive(r io.Reader) {
	var (
		cmd  int
		scan = bufio.NewScanner(r)
		env  = eval.Default()
	)
	cmd++
	io.WriteString(os.Stdout, fmt.Sprintf(in, cmd))
	for scan.Scan() {
		line := scan.Text()
		if strings.TrimSpace(line) == "" {
			cmd++
			io.WriteString(os.Stdout, fmt.Sprintf(in, cmd))
			continue
		}
		res, err := env.EvalString(line)
		if err != nil {
			if builtins.IsExit(err) {
				var code int
				switch c := res.Raw().(type) {
				case int64:
					code = int(c)
				case float64:
					code = int(c)
				default:
				}
				os.Exit(code)
			}
			fmt.Fprintf(os.Stderr, fmt.Sprintf(nok, cmd, err))
			fmt.Fprintln(os.Stderr)
		} else {
			fmt.Fprintf(os.Stdout, ok, cmd, res)
			fmt.Fprintln(os.Stdout)
			env.Define("_", res)
		}
		cmd++
		io.WriteString(os.Stdout, fmt.Sprintf(in, cmd))
	}
	fmt.Fprintln(os.Stdout)
}

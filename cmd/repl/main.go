package main

import (
	"fmt"
	"io"
	"bufio"
	"os"
	"strings"

	"github.com/midbel/buddy"
)

const (
	in  = "\x1b[1;97min [%3d]:\x1b[0m "
	ok  = "\x1b[1;92mout[%3d]:\x1b[0m %v"
	nok = "\x1b[1;91mout[%3d]:\x1b[0m %s"
)

func main() {
	repl()
}

func repl() {
	var (
		cmd int
		scan = bufio.NewScanner(os.Stdin)
		env = buddy.EmptyEnv[any]()
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
		res, err := buddy.EvalWithEnv(strings.NewReader(line), env)
		if err != nil {
			fmt.Fprintf(os.Stderr, fmt.Sprintf(nok, cmd, err))
			fmt.Fprintln(os.Stderr)
		} else {
			fmt.Fprintf(os.Stdout, ok, cmd, res)
			fmt.Fprintln(os.Stdout)
		}
		cmd++
		io.WriteString(os.Stdout, fmt.Sprintf(in, cmd))
	}
	fmt.Fprintln(os.Stdout)
}
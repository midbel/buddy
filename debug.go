package buddy

import (
	"fmt"
	"io"
	"strings"
)

func Debug(w io.Writer, r io.Reader) error {
	expr, err := Parse(r)
	if err == nil {
		printAST(w, expr, 0)
	}
	return err
}

func printAST(w io.Writer, e Expression, level int) {
	prefix := strings.Repeat(" ", level*2)
	if level == 0 {
		prefix = ""
	}
	switch e := e.(type) {
	case script:
		fmt.Fprintln(w, prefix+"script")
		for i := range e.list {
			printAST(w, e.list[i], level+1)
		}
	case call:
		fmt.Fprintln(w, fmt.Sprintf("%scall(%s)", prefix, e.ident))
		for i := range e.args {
			printAST(w, e.args[i], level+1)
		}
	case binary:
		fmt.Fprintln(w, prefix+"binary")
		printAST(w, e.left, level+1)
		printAST(w, e.right, level+1)
	case unary:
		fmt.Fprintln(w, prefix+"unary")
		printAST(w, e.right, level+1)
	case assign:
		fmt.Fprintln(w, fmt.Sprintf("%sassign(%s)", prefix, e.ident))
		printAST(w, e.right, level+1)
	case test:
		fmt.Fprintln(w, prefix+"if")
		printAST(w, e.cdt, level+1)
		printAST(w, e.csq, level+1)
		if e.alt != nil {
			printAST(w, e.alt, level+1)
		}
	case while:
		fmt.Fprintln(w, prefix+"while")
		printAST(w, e.cdt, level+1)
		printAST(w, e.body, level+1)
	case returned:
		fmt.Fprintln(w, prefix+"return")
		if e.right != nil {
			printAST(w, e.right, level+1)
		}
	case breaked:
		fmt.Fprintln(w, prefix+"break")
	case continued:
		fmt.Fprintln(w, prefix+"continue")
	case function:
		fmt.Fprintln(w, fmt.Sprintf("%sfunction(%s)", prefix, e.ident))
		for i := range e.params {
			printAST(w, e.params[i], level+1)
		}
		printAST(w, e.body, level+1)
	case parameter:
		fmt.Fprintln(w, fmt.Sprintf("%parameter(%s)", prefix, e.ident))
		if e.expr != nil {
			printAST(w, e.expr, level+1)
		}
	case boolean:
		fmt.Fprintln(w, fmt.Sprintf("%sboolean(%t)", prefix, e.value))
	case number:
		fmt.Fprintln(w, fmt.Sprintf("%snumber(%.5f)", prefix, e.value))
	case literal:
		fmt.Fprintln(w, fmt.Sprintf("%sliteral(%.5f)", prefix, e.str))
	}
}

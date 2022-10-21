package buddy

import (
	"fmt"
	"io"
	"math"
	"strings"
)

func Debug(w io.Writer, r io.Reader, visit bool) error {
	expr, err := Parse(r)
	if err != nil {
		return err
	}
	if visit {
		resolv := NewResolver()
		if s, ok := expr.(script); ok {
			resolv.symbols = s.symbols
		}
		visitors := []visitFunc{
			checkVariables,
			replaceFunctionArgs,
			inlineFunctionCall,
			replaceValue,
		}
		if expr, err = traverse(expr, resolv, visitors); err != nil {
			return err
		}
		for k, e := range resolv.symbols {
			resolv.symbols[k], err = traverse(e, resolv, visitors)
			if err != nil {
				return err
			}
		}
	}
	printAST(w, expr, 0)
	return nil
}

func printAST(w io.Writer, e Expression, level int) {
	prefix := strings.Repeat(" ", level*2)
	if level == 0 {
		prefix = ""
	}
	switch e := e.(type) {
	case script:
		fmt.Fprintln(w, prefix+"block")
		for i := range e.list {
			printAST(w, e.list[i], level+1)
		}
		for _, f := range e.symbols {
			printAST(w, f, level)
		}
	case call:
		fmt.Fprintln(w, fmt.Sprintf("%scall(%s)", prefix, e.ident))
		for i := range e.args {
			printAST(w, e.args[i], level+1)
		}
	case binary:
		fmt.Fprintln(w, fmt.Sprintf("%sbinary(%s)", prefix, binaryOp(e.op)))
		printAST(w, e.left, level+1)
		printAST(w, e.right, level+1)
	case unary:
		fmt.Fprintln(w, fmt.Sprintf("%sunary(%s)", prefix, unaryOp(e.op)))
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
		fmt.Fprintln(w, fmt.Sprintf("%sparameter(%s)", prefix, e.ident))
		if e.expr != nil {
			printAST(w, e.expr, level+1)
		}
	case boolean:
		fmt.Fprintln(w, fmt.Sprintf("%sboolean(%t)", prefix, e.value))
	case number:
		if math.Round(e.value) == e.value {
			fmt.Fprintln(w, fmt.Sprintf("%snumber(%d)", prefix, int(e.value)))
		} else {
			fmt.Fprintln(w, fmt.Sprintf("%snumber(%f)", prefix, e.value))
		}
	case literal:
		fmt.Fprintln(w, fmt.Sprintf("%sliteral(%s)", prefix, e.str))
	case variable:
		fmt.Fprintln(w, fmt.Sprintf("%svariable(%s)", prefix, e.ident))
	case dict:
		fmt.Fprintln(w, fmt.Sprintf("%sdict(%d)", prefix, len(e.list)))
		for k, v := range e.list {
			printAST(w, k, level+1)
			printAST(w, v, level+2)
		}
	case array:
		fmt.Fprintln(w, fmt.Sprintf("%sarray(%d)", prefix, len(e.list)))
		for i := range e.list {
			printAST(w, e.list[i], level+1)
		}
	case index:
		fmt.Fprintln(w, prefix+"index")
		printAST(w, e.expr, level+1)
	}
}

func unaryOp(op rune) string {
	switch op {
	case Sub:
		return "rev"
	case Not:
		return "not"
	}
	return "?"
}

func binaryOp(op rune) string {
	switch op {
	case Add:
		return "add"
	case Sub:
		return "sub"
	case Div:
		return "div"
	case Mod:
		return "mod"
	case Pow:
		return "pow"
	case And:
		return "and"
	case Or:
		return "or"
	case Eq:
		return "eq"
	case Ne:
		return "ne"
	case Lt:
		return "lt"
	case Le:
		return "le"
	case Gt:
		return "gt"
	case Ge:
		return "ge"
	}
	return "?"
}

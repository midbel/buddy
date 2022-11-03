package buddy

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/midbel/buddy/token"
)

func Debug(w io.Writer, r io.Reader, visit bool) error {
	expr, err := Parse(r)
	if err != nil {
		return err
	}
	if visit {
		expr, err = optimize(expr)
	}
	printAST(w, expr, 0)
	return nil
}

func optimize(expr Expression) (Expression, error) {
	s, ok := expr.(script)
	if !ok {
		return expr, nil
	}
	var (
		res = NewResolver()
		err error
	)
	for k, e := range s.symbols {
		s.symbols[k], err = traverse(e, res, visitors)
		if err != nil {
			return nil, err
		}
	}
	res.symbols = s.symbols
	return traverse(expr, res, visitors)
}

func printAST(w io.Writer, e Expression, level int) {
	prefix := strings.Repeat(" ", level*2)
	if level == 0 {
		prefix = ""
	}
	switch e := e.(type) {
	case script:
		fmt.Fprintf(w, "%sblock", prefix)
		fmt.Fprintln(w)
		for i := range e.list {
			printAST(w, e.list[i], level+1)
		}
		for _, f := range e.symbols {
			printAST(w, f, level)
		}
	case call:
		fmt.Fprintf(w, "%scall(%s)", prefix, e.ident)
		fmt.Fprintln(w)
		for i := range e.args {
			printAST(w, e.args[i], level+1)
		}
	case assert:
		fmt.Fprintf(w, "%sassert", prefix)
		fmt.Fprintln(w)
		printAST(w, e.expr, level+1)
	case binary:
		fmt.Fprintf(w, "%sbinary(%s)", prefix, binaryOp(e.op))
		fmt.Fprintln(w)
		printAST(w, e.left, level+1)
		printAST(w, e.right, level+1)
	case unary:
		fmt.Fprintf(w, "%sunary(%s)", prefix, unaryOp(e.op))
		fmt.Fprintln(w)
		printAST(w, e.right, level+1)
	case assign:
		fmt.Fprintf(w, "%sassign", prefix)
		fmt.Fprintln(w)
		printAST(w, e.ident, level+1)
		printAST(w, e.right, level+2)
	case test:
		fmt.Fprintf(w, "%sif", prefix)
		fmt.Fprintln(w)
		printAST(w, e.cdt, level+1)
		printAST(w, e.csq, level+1)
		if e.alt != nil {
			printAST(w, e.alt, level+1)
		}
	case listcomp:
		fmt.Fprintf(w, "%slistcomp", prefix)
		fmt.Fprintln(w)
		printAST(w, e.body, level+1)
		for i := range e.list {
			printAST(w, e.list[i], level+1)
		}
	case compitem:
		fmt.Fprintf(w, "%scompitem(%s)", prefix, e.ident)
		fmt.Fprintln(w)
		printAST(w, e.iter, level+1)
		for i := range e.cdt {
			printAST(w, e.cdt[i], level+1)
		}
	case foreach:
		fmt.Fprintf(w, "%sforeach(%s)", prefix, e.ident)
		fmt.Fprintln(w)
		printAST(w, e.iter, level+1)
		printAST(w, e.body, level+1)
	case while:
		fmt.Fprintf(w, "%swhile", prefix)
		fmt.Fprintln(w)
		printAST(w, e.cdt, level+1)
		printAST(w, e.body, level+1)
	case forloop:
		fmt.Fprintf(w, "%sfor", prefix)
		fmt.Fprintln(w)
		if e.init != nil {
			printAST(w, e.init, level+1)
		}
		if e.cdt != nil {
			printAST(w, e.cdt, level+1)
		}
		if e.incr != nil {
			printAST(w, e.incr, level+1)
		}
		printAST(w, e.body, level+1)
	case returned:
		fmt.Fprintf(w, "%sreturn", prefix)
		fmt.Fprintln(w)
		if e.right != nil {
			printAST(w, e.right, level+1)
		}
	case breaked:
		fmt.Fprintf(w, "%sbreak", prefix)
		fmt.Fprintln(w)
	case continued:
		fmt.Fprintf(w, "%scontinue", prefix)
		fmt.Fprintln(w)
	case function:
		fmt.Fprintf(w, "%sfunction(%s)", prefix, e.ident)
		fmt.Fprintln(w)
		for i := range e.params {
			printAST(w, e.params[i], level+1)
		}
		printAST(w, e.body, level+1)
	case parameter:
		fmt.Fprintf(w, "%sparameter(%s)", prefix, e.ident)
		fmt.Fprintln(w)
		if e.expr != nil {
			printAST(w, e.expr, level+1)
		}
	case module:
		fmt.Fprintf(w, "%simport(%s", prefix, strings.Join(e.ident, "."))
		if e.alias != "" {
			fmt.Fprintf(w, ":%s", e.alias)
		}
		fmt.Fprintln(w, ")")
		for i := range e.symbols {
			printAST(w, e.symbols[i], level+1)
		}
	case symbol:
		fmt.Fprintf(w, "%ssymbol(%s", prefix, e.ident)
		if e.alias != "" {
			fmt.Fprintf(w, ":%s", e.alias)
		}
		fmt.Fprintln(w, ")")
	case path:
		fmt.Fprintf(w, "%spath(%s)", prefix, e.ident)
		fmt.Fprintln(w)
		printAST(w, e.right, level+1)
	case boolean:
		fmt.Fprintf(w, "%sboolean(%t)", prefix, e.value)
		fmt.Fprintln(w)
	case double:
		if math.Round(e.value) == e.value {
			fmt.Fprintf(w, "%sdouble(%d)", prefix, int(e.value))
		} else {
			fmt.Fprintf(w, "%sdouble(%f)", prefix, e.value)
		}
		fmt.Fprintln(w)
	case integer:
		fmt.Fprintf(w, "%sinteger(%d)", prefix, e.value)
		fmt.Fprintln(w)
	case literal:
		fmt.Fprintf(w, "%sliteral(%s)", prefix, e.str)
		fmt.Fprintln(w)
	case variable:
		fmt.Fprintf(w, "%svariable(%s)", prefix, e.ident)
		fmt.Fprintln(w)
	case dict:
		fmt.Fprintf(w, "%sdict(%d)", prefix, len(e.list))
		fmt.Fprintln(w)
		for k, v := range e.list {
			printAST(w, k, level+1)
			printAST(w, v, level+2)
		}
	case array:
		fmt.Fprintf(w, "%sarray(%d)", prefix, len(e.list))
		fmt.Fprintln(w)
		for i := range e.list {
			printAST(w, e.list[i], level+1)
		}
	case index:
		fmt.Fprintf(w, "%sindex", prefix)
		fmt.Fprintln(w)
		printAST(w, e.arr, level+1)
		for i := range e.list {
			printAST(w, e.list[i], level+2)
		}
	}
}

func unaryOp(op rune) string {
	switch op {
	case token.Sub:
		return "rev"
	case token.Not:
		return "not"
	case token.BinNot:
		return "binary-not"
	}
	return "?"
}

func binaryOp(op rune) string {
	switch op {
	case token.Add:
		return "add"
	case token.Sub:
		return "sub"
	case token.Div:
		return "div"
	case token.Mod:
		return "mod"
	case token.Pow:
		return "pow"
	case token.And:
		return "and"
	case token.Or:
		return "or"
	case token.Eq:
		return "eq"
	case token.Ne:
		return "ne"
	case token.Lt:
		return "lt"
	case token.Le:
		return "le"
	case token.Gt:
		return "gt"
	case token.Ge:
		return "ge"
	case token.BinOr:
		return "binary-or"
	case token.BinAnd:
		return "binary-and"
	case token.BinXor:
		return "binary-xor"
	case token.BinNot:
		return "binary-not"
	case token.Lshift:
		return "left-shift"
	case token.Rshift:
		return "right-shift"
	}
	return "?"
}

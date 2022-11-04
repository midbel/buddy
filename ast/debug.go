package ast

import (
	"fmt"
	"io"
	"strings"

	"github.com/midbel/buddy/token"
)

func Debug(w io.Writer, expr Expression) {
	printAST(w, expr, 0)
}

func printAST(w io.Writer, e Expression, level int) {
	prefix := strings.Repeat(" ", level*2)
	if level == 0 {
		prefix = ""
	}
	switch e := e.(type) {
	case Script:
		fmt.Fprintf(w, "%sblock", prefix)
		fmt.Fprintln(w)
		for i := range e.List {
			printAST(w, e.List[i], level+1)
		}
		for _, f := range e.Symbols {
			printAST(w, f, level)
		}
	case Call:
		fmt.Fprintf(w, "%scall(%s)", prefix, e.Ident)
		fmt.Fprintln(w)
		for i := range e.Args {
			printAST(w, e.Args[i], level+1)
		}
	case Assert:
		fmt.Fprintf(w, "%sassert", prefix)
		fmt.Fprintln(w)
		printAST(w, e.Expr, level+1)
	case Binary:
		fmt.Fprintf(w, "%sbinary(%s)", prefix, binaryOp(e.Op))
		fmt.Fprintln(w)
		printAST(w, e.Left, level+1)
		printAST(w, e.Right, level+1)
	case Unary:
		fmt.Fprintf(w, "%sunary(%s)", prefix, unaryOp(e.Op))
		fmt.Fprintln(w)
		printAST(w, e.Right, level+1)
	case Assign:
		fmt.Fprintf(w, "%sassign", prefix)
		fmt.Fprintln(w)
		printAST(w, e.Ident, level+1)
		printAST(w, e.Right, level+2)
	case Test:
		fmt.Fprintf(w, "%sif", prefix)
		fmt.Fprintln(w)
		printAST(w, e.Cdt, level+1)
		printAST(w, e.Csq, level+1)
		if e.Alt != nil {
			printAST(w, e.Alt, level+1)
		}
	case ListComp:
		fmt.Fprintf(w, "%slistcomp", prefix)
		fmt.Fprintln(w)
		printAST(w, e.Body, level+1)
		for i := range e.List {
			printAST(w, e.List[i], level+1)
		}
	case DictComp:
		fmt.Fprintf(w, "%sdictcomp", prefix)
		fmt.Fprintln(w)
		printAST(w, e.Key, level+1)
		printAST(w, e.Val, level+1)
		for i := range e.List {
			printAST(w, e.List[i], level+1)
		}
	case CompItem:
		fmt.Fprintf(w, "%scompitem(%s)", prefix, e.Ident)
		fmt.Fprintln(w)
		printAST(w, e.Iter, level+1)
		for i := range e.Cdt {
			printAST(w, e.Cdt[i], level+1)
		}
	case ForEach:
		fmt.Fprintf(w, "%sforeach(%s)", prefix, e.Ident)
		fmt.Fprintln(w)
		printAST(w, e.Iter, level+1)
		printAST(w, e.Body, level+1)
	case While:
		fmt.Fprintf(w, "%swhile", prefix)
		fmt.Fprintln(w)
		printAST(w, e.Cdt, level+1)
		printAST(w, e.Body, level+1)
	case For:
		fmt.Fprintf(w, "%sfor", prefix)
		fmt.Fprintln(w)
		if e.Init != nil {
			printAST(w, e.Init, level+1)
		}
		if e.Cdt != nil {
			printAST(w, e.Cdt, level+1)
		}
		if e.Incr != nil {
			printAST(w, e.Incr, level+1)
		}
		printAST(w, e.Body, level+1)
	case Return:
		fmt.Fprintf(w, "%sreturn", prefix)
		fmt.Fprintln(w)
		if e.Right != nil {
			printAST(w, e.Right, level+1)
		}
	case Break:
		fmt.Fprintf(w, "%sbreak", prefix)
		fmt.Fprintln(w)
	case Continue:
		fmt.Fprintf(w, "%scontinue", prefix)
		fmt.Fprintln(w)
	case Function:
		fmt.Fprintf(w, "%sfunction(%s)", prefix, e.Ident)
		fmt.Fprintln(w)
		for i := range e.Params {
			printAST(w, e.Params[i], level+1)
		}
		printAST(w, e.Body, level+1)
	case Parameter:
		fmt.Fprintf(w, "%sparameter(%s)", prefix, e.Ident)
		fmt.Fprintln(w)
		if e.Expr != nil {
			printAST(w, e.Expr, level+1)
		}
	case Import:
		fmt.Fprintf(w, "%simport(%s", prefix, strings.Join(e.Ident, "."))
		if e.Alias != "" {
			fmt.Fprintf(w, ":%s", e.Alias)
		}
		fmt.Fprintln(w, ")")
		for i := range e.Symbols {
			printAST(w, e.Symbols[i], level+1)
		}
	case Symbol:
		fmt.Fprintf(w, "%ssymbol(%s", prefix, e.Ident)
		if e.Alias != "" {
			fmt.Fprintf(w, ":%s", e.Alias)
		}
		fmt.Fprintln(w, ")")
	case Path:
		fmt.Fprintf(w, "%spath(%s)", prefix, e.Ident)
		fmt.Fprintln(w)
		printAST(w, e.Right, level+1)
	case Boolean:
		fmt.Fprintf(w, "%sboolean(%t)", prefix, e.Value)
		fmt.Fprintln(w)
	case Double:
		fmt.Fprintf(w, "%sdouble(%f)", prefix, e.Value)
		fmt.Fprintln(w)
	case Integer:
		fmt.Fprintf(w, "%sinteger(%d)", prefix, e.Value)
		fmt.Fprintln(w)
	case Literal:
		fmt.Fprintf(w, "%sliteral(%s)", prefix, e.Str)
		fmt.Fprintln(w)
	case Variable:
		fmt.Fprintf(w, "%svariable(%s)", prefix, e.Ident)
		fmt.Fprintln(w)
	case Dict:
		fmt.Fprintf(w, "%sdict(%d)", prefix, len(e.List))
		fmt.Fprintln(w)
		for k, v := range e.List {
			printAST(w, k, level+1)
			printAST(w, v, level+2)
		}
	case Array:
		fmt.Fprintf(w, "%sarray(%d)", prefix, len(e.List))
		fmt.Fprintln(w)
		for i := range e.List {
			printAST(w, e.List[i], level+1)
		}
	case Index:
		fmt.Fprintf(w, "%sindex", prefix)
		fmt.Fprintln(w)
		printAST(w, e.Arr, level+1)
		for i := range e.List {
			printAST(w, e.List[i], level+1)
		}
	case Slice:
		fmt.Fprintf(w, "%sslice", prefix)
		fmt.Fprintln(w)
		if e.Start != nil {
			printAST(w, e.Start, level+1)
		}
		if e.End != nil {
			printAST(w, e.End, level+1)
		}
		if e.Step != nil {
			printAST(w, e.Step, level+1)
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

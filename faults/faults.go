package faults

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/midbel/buddy/parse"
)

func PrintError(w io.Writer, err error) {
	if err == nil {
		return
	}
	var perr parse.ParseError
	if errors.As(err, &perr) {
		printParseError(w, perr)
	} else {
		fmt.Fprintln(w, err)
	}
}

const MaxErrorCount = 15

type ErrorList []error

func (es *ErrorList) Size() int {
	return len((*es))
}

func (es *ErrorList) Append(err error) {
	*es = append(*es, err)
}

func (es *ErrorList) Error() string {
	var str strings.Builder
	for i := range *es {
		if i > 0 {
			str.WriteString("\n")
		}
		str.WriteString((*es)[i].Error())
	}
	return str.String()
}

func printParseError(w io.Writer, err parse.ParseError) {
	var (
		space = strings.Repeat(" ", err.Token.Position.Column-1)
		tilde = "^"
	)
	if err.Token.Literal != "" {
		tilde += strings.Repeat("~", len(err.Token.Literal))
	}

	line := strings.ReplaceAll(err.Line, "\t", " ")

	fmt.Fprintf(w, "\x1b[1;91m%s: parsing error at %s:\x1b[0m", err.File, err.Position)
	fmt.Fprintln(w)
	fmt.Fprintln(w, line)
	fmt.Fprintf(w, "%s%s \x1b[1;91m%s\x1b[0m", space, tilde, err.Message)
	fmt.Fprintln(w)
}

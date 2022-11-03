package buddy

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func PrintError(w io.Writer, err error) {
	if err == nil {
		return
	}
	var perr ParseError
	if errors.As(err, &perr) {
		printParseError(w, perr)
	} else {
		fmt.Fprintln(w, err)
	}
}

type ParseError struct {
	Token
	File    string
	Line    string
	Message string
}

func (e ParseError) Error() string {
	if e.File == "" {
		e.File = "<input>"
	} else {
		e.File = filepath.Base(e.File)
	}
	return fmt.Sprintf("%s %s: %s", e.File, e.Position, e.Message)
}

type errorsList []error

func (e errorsList) Error() string {
	var str strings.Builder
	for i := range e {
		if i > 0 {
			str.WriteString("\n")
		}
		str.WriteString(e[i].Error())
	}
	return str.String()
}

func printParseError(w io.Writer, err ParseError) {
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

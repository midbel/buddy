package buddy

import (
	"fmt"
)

const (
	kwIf       = "if"
	kwElse     = "else"
	kwWhile    = "while"
	kwReturn   = "return"
	kwDef      = "def"
	kwTrue     = "true"
	kwFalse    = "false"
	kwBreak    = "break"
	kwContinue = "continue"
	kwImport   = "import"
)

func isKeyword(str string) bool {
	switch str {
	case kwDef:
	case kwIf:
	case kwElse:
	case kwWhile:
	case kwBreak:
	case kwContinue:
	case kwReturn:
	case kwImport:
	default:
		return false
	}
	return true
}

const (
	Invalid rune = -(iota + 1)
	Keyword
	Literal
	Ident
	Boolean
	Variable
	Number
	Command
	Comment
	Comma
	Lparen
	Rparen
	Lcurly
	Rcurly
	Expr
	Sum
	Range
	RangeSum
	Add
	AddAssign
	Sub
	SubAssign
	Mul
	MulAssign
	Pow
	Div
	DivAssign
	Mod
	ModAssign
	Lt
	Le
	Gt
	Ge
	Eq
	Ne
	Assign
	Ternary
	Alt
	Not
	And
	Or
	EOL
	EOF
)

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Literal string
	Type    rune
	Position
}

func (t Token) String() string {
	var prefix string
	switch t.Type {
	default:
		prefix = "unknown"
	case Invalid:
		prefix = "invalid"
	case Literal:
		prefix = "literal"
	case Number:
		prefix = "number"
	case Comment:
		prefix = "comment"
	case Keyword:
		prefix = "keyword"
	case Boolean:
		prefix = "boolean"
	case Ident:
		prefix = "identifier"
	case Comma:
		return "<comma>"
	case EOL:
		return "<eol>"
	case EOF:
		return "<eof>"
	case Lparen:
		return "<lparen>"
	case Rparen:
		return "<rparen>"
	case Lcurly:
		return "<lcurly>"
	case Rcurly:
		return "<rcurly>"
	case Add:
		return "<add>"
	case AddAssign:
		return "<add-assign>"
	case Sub:
		return "<subtract>"
	case SubAssign:
		return "<subtract-assign>"
	case Mul:
		return "<multiply>"
	case MulAssign:
		return "<multiply-assign>"
	case Div:
		return "<divide>"
	case DivAssign:
		return "<divide-assign>"
	case Mod:
		return "<modulo>"
	case ModAssign:
		return "<modulo-assign>"
	case Pow:
		return "<power>"
	case Lt:
		return "<lt>"
	case Le:
		return "<le>"
	case Gt:
		return "<gt>"
	case Ge:
		return "<ge>"
	case Eq:
		return "<eq>"
	case Ne:
		return "<ne>"
	case And:
		return "<and>"
	case Or:
		return "<or>"
	case Assign:
		return "<assign>"
	case Ternary:
		return "<ternary>"
	case Alt:
		return "<alter>"
	case Not:
		return "<not>"
	}
	return fmt.Sprintf("%s(%s)", prefix, t.Literal)
}
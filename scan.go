package buddy

import (
	"bytes"
	"io"
	"unicode/utf8"
)

type Scanner struct {
	input []byte

	curr int
	next int
	char rune

	Position
	seen int
}

func Scan(r io.Reader) *Scanner {
	in, _ := io.ReadAll(r)
	x := Scanner{
		input: bytes.ReplaceAll(in, []byte{cr, nl}, []byte{nl}),
	}
	x.Line++
	return &x
}

func (s *Scanner) getLine(pos Position) string {
	var start int
	for i := 0; i < pos.Line-1; i++ {
		x := bytes.IndexByte(s.input[start:], nl)
		start += x + 1
	}
	end := bytes.IndexByte(s.input[start:], nl)
	if end < 0 {
		end = len(s.input)
	} else {
		end += start
	}
	return string(s.input[start:end])
}

func (s *Scanner) Scan() Token {
	s.read()
	if isBlank(s.char) {
		s.skipBlank()
		s.read()
	}
	var tok Token
	tok.Position = s.Position
	if s.done() {
		tok.Type = EOF
		return tok
	}
	switch {
	case isDigit(s.char):
		s.scanNumber(&tok)
	case isOperator(s.char):
		s.scanOperator(&tok)
	case isLetter(s.char):
		s.scanIdent(&tok)
	case isQuote(s.char):
		s.scanLiteral(&tok)
	case isNL(s.char):
		s.skipNL()
		tok.Type = EOL
	default:
		tok.Type = Invalid
	}
	return tok
}

func (s *Scanner) scanLiteral(tok *Token) {
	quote := s.char
	s.read()
	pos := s.curr
	for s.char != quote && !s.done() {
		s.read()
	}
	tok.Type = Literal
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanIdent(tok *Token) {
	defer s.unread()
	pos := s.curr
	for isAlpha(s.char) {
		s.read()
	}
	tok.Type = Ident
	tok.Literal = string(s.input[pos:s.curr])
	if isKeyword(tok.Literal) {
		tok.Type = Keyword
	}
	if tok.Literal == kwTrue || tok.Literal == kwFalse {
		tok.Type = Boolean
	}
}

func (s *Scanner) scanNumber(tok *Token) {
	defer s.unread()

	pos := s.curr
	for isDigit(s.char) {
		s.read()
	}
	if s.char == dot {
		s.read()
		for isDigit(s.char) {
			s.read()
		}
	}
	tok.Type = Number
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanOperator(tok *Token) {
	switch s.char {
	case dot:
		tok.Type = Dot
	case caret:
		tok.Type = BinXor
		if s.peek() == equal {
			tok.Type = BinXorAssign
			s.read()
		}
	case tilde:
		tok.Type = BinNot
	case ampersand:
		tok.Type = BinAnd
		if k := s.peek(); k == equal {
			tok.Type = BinAndAssign
			s.read()
		} else if k == ampersand {
			tok.Type = And
			s.read()
		}
	case pipe:
		tok.Type = BinOr
		if k := s.peek(); k == equal {
			tok.Type = BinOrAssign
			s.read()
		} else if k == pipe {
			tok.Type = Or
			s.read()
		}
	case bang:
		tok.Type = Not
		if s.peek() == equal {
			tok.Type = Ne
			s.read()
		}
	case equal:
		tok.Type = Assign
		if s.peek() == equal {
			tok.Type = Eq
			s.read()
		}
	case langle:
		tok.Type = Lt
		if k := s.peek(); k == equal {
			tok.Type = Le
			s.read()
		} else if k == langle {
			tok.Type = Lshift
			s.read()
			if s.char == equal {
				tok.Type = LshiftAssign
			}
		}
	case rangle:
		tok.Type = Gt
		if k := s.peek(); k == equal {
			tok.Type = Ge
			s.read()
		} else if k == rangle {
			tok.Type = Rshift
			s.read()
			if s.char == equal {
				tok.Type == RshiftAssign
			}
		}
	case comma:
		tok.Type = Comma
	case lparen:
		tok.Type = Lparen
	case rparen:
		tok.Type = Rparen
	case lsquare:
		tok.Type = Lsquare
		if isNL(s.peek()) {
			s.read()
			s.skipNL()
		}
	case rsquare:
		tok.Type = Rsquare
	case lcurly:
		tok.Type = Lcurly
		if isNL(s.peek()) {
			s.read()
			s.skipNL()
		}
	case rcurly:
		tok.Type = Rcurly
	case plus:
		tok.Type = Add
		if s.peek() == equal {
			tok.Type = AddAssign
			s.read()
		}
	case minus:
		tok.Type = Sub
		if s.peek() == equal {
			tok.Type = SubAssign
			s.read()
		}
	case star:
		tok.Type = Mul
		if s.peek() == star {
			tok.Type = Pow
			s.read()
		} else if s.peek() == equal {
			tok.Type = MulAssign
			s.read()
		}
	case slash:
		tok.Type = Div
		if s.peek() == equal {
			tok.Type = DivAssign
			s.read()
		}
	case percent:
		tok.Type = Mod
		if s.peek() == equal {
			tok.Type = ModAssign
			s.read()
		}
	case semicolon:
		tok.Type = EOL
	case question:
		tok.Type = Ternary
	case colon:
		tok.Type = Colon
	default:
		tok.Type = Invalid
	}
}

func (s *Scanner) done() bool {
	return s.char == utf8.RuneError
}

func (s *Scanner) peek() rune {
	r, _ := utf8.DecodeRune(s.input[s.next:])
	return r
}

func (s *Scanner) read() {
	if s.curr >= len(s.input) || s.char == utf8.RuneError {
		return
	}
	if s.char == nl {
		s.seen = s.Column
		s.Line++
		s.Column = 0
	}
	s.Column++

	r, size := utf8.DecodeRune(s.input[s.next:])
	s.curr = s.next
	s.next += size
	s.char = r
}

func (s *Scanner) unread() {
	var size int
	if s.char == nl {
		s.Line--
		s.Column = s.seen
	}
	s.Column--
	s.char, size = utf8.DecodeRune(s.input[s.curr:])
	s.next = s.curr
	s.curr -= size
}

func (s *Scanner) skipBlank() {
	s.skip(isBlank)
}

func (s *Scanner) skipNL() {
	s.skip(isNL)
}

func (s *Scanner) skip(accept func(rune) bool) {
	defer s.unread()
	for accept(s.char) && !s.done() {
		s.read()
	}
}

const (
	space      rune = ' '
	tab             = '\t'
	cr              = '\r'
	nl              = '\n'
	colon           = ':'
	plus            = '+'
	minus           = '-'
	slash           = '/'
	star            = '*'
	percent         = '%'
	semicolon       = ';'
	equal           = '='
	lparen          = '('
	rparen          = ')'
	lcurly          = '{'
	rcurly          = '}'
	lsquare         = '['
	rsquare         = ']'
	comma           = ','
	hash            = '#'
	dollar          = '$'
	dot             = '.'
	squote          = '\''
	dquote          = '"'
	underscore      = '_'
	question        = '?'
	bang            = '!'
	langle          = '<'
	rangle          = '>'
	ampersand       = '&'
	pipe            = '|'
	caret           = '^'
	tilde           = '~'
)

func isOperator(r rune) bool {
	switch r {
	case caret:
	case tilde:
	case plus:
	case minus:
	case star:
	case percent:
	case slash:
	case semicolon:
	case dot:
	case lparen:
	case rparen:
	case equal:
	case comma:
	case question:
	case colon:
	case bang:
	case langle:
	case rangle:
	case ampersand:
	case pipe:
	case lcurly:
	case rcurly:
	case lsquare:
	case rsquare:
	default:
		return false
	}
	return true
}

func isChar(r rune) bool {
	return isLower(r) || isUpper(r)
}

func isLetter(r rune) bool {
	return isChar(r) || r == underscore
}

func isAlpha(r rune) bool {
	return isLetter(r) || isDigit(r)
}

func isQuote(r rune) bool {
	return r == squote || r == dquote
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isBlank(r rune) bool {
	return r == space || r == tab
}

func isNL(r rune) bool {
	return r == nl
}

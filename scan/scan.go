package scan

import (
	"bytes"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/midbel/buddy/token"
)

type Scanner struct {
	input []byte

	curr int
	next int
	char rune

	token.Position
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

func (s *Scanner) CurrentLine(pos token.Position) string {
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

func (s *Scanner) Scan() token.Token {
	s.read()
	if isBlank(s.char) {
		s.skipBlank()
		s.read()
	}
	var tok token.Token
	tok.Position = s.Position
	if s.done() {
		tok.Type = token.EOF
		return tok
	}
	if pk := s.peek(); s.char == slash && (pk == slash || pk == star) {
		s.scanComment(&tok)
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
		tok.Type = token.EOL
	default:
		tok.Type = token.Invalid
	}
	return tok
}

func (s *Scanner) scanComment(tok *token.Token) {
	var (
		accept = func() bool { return isNL(s.char) }
		long   bool
	)
	if long = s.peek() == star; long {
		accept = func() bool { return s.char == star && s.peek() == slash }
	}
	s.read()
	s.read()
	s.skipBlank()
	pos := s.curr
	for !accept() {
		s.read()
	}
	tok.Type = token.Comment
	tok.Literal = string(s.input[pos:s.curr])
	if long {
		s.read()
		s.read()
	}
	s.skipNL()
}

func (s *Scanner) scanLiteral(tok *token.Token) {
	quote := s.char
	s.read()
	pos := s.curr
	for s.char != quote && !s.done() {
		s.read()
	}
	tok.Type = token.Literal
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanIdent(tok *token.Token) {
	defer s.unread()
	pos := s.curr
	for isAlpha(s.char) {
		s.read()
	}
	tok.Type = token.Ident
	tok.Literal = string(s.input[pos:s.curr])
	if token.IsKeyword(tok.Literal) {
		tok.Type = token.Keyword
	}
	if tok.Literal == token.KwTrue || tok.Literal == token.KwFalse {
		tok.Type = token.Boolean
	}
}

func (s *Scanner) scanNumber(tok *token.Token) {
	scan := func(accept func(rune) bool) {
		for accept(s.char) {
			s.read()
			if s.char == underscore && accept(s.char) {
				s.read()
			}
		}
	}
	defer s.unread()
	pos := s.curr
	tok.Type = token.Integer
	if s.char == '0' {
		var accept func(rune) bool
		switch s.peek() {
		case 'b':
			accept = isBin
		case 'o':
			accept = isOctal
		case 'x':
			accept = isHex
		default:
		}
		if accept != nil {
			s.read()
			s.read()
			scan(accept)
			tok.Literal = string(s.input[pos:s.curr])
			return
		}
	}
	scan(isDigit)
	if s.char == dot {
		s.read()
		scan(isDigit)
		tok.Type = token.Double
	}
	tok.Literal = string(s.input[pos:s.curr])
}

func (s *Scanner) scanOperator(tok *token.Token) {
	switch s.char {
	case dot:
		tok.Type = token.Dot
	case caret:
		tok.Type = token.BinXor
		if s.peek() == equal {
			tok.Type = token.BinXorAssign
			s.read()
		}
	case tilde:
		tok.Type = token.BinNot
	case ampersand:
		tok.Type = token.BinAnd
		if k := s.peek(); k == equal {
			tok.Type = token.BinAndAssign
			s.read()
		} else if k == ampersand {
			tok.Type = token.And
			s.read()
		}
	case pipe:
		tok.Type = token.BinOr
		if k := s.peek(); k == equal {
			tok.Type = token.BinOrAssign
			s.read()
		} else if k == pipe {
			tok.Type = token.Or
			s.read()
		}
	case bang:
		tok.Type = token.Not
		if s.peek() == equal {
			tok.Type = token.Ne
			s.read()
		}
	case equal:
		tok.Type = token.Assign
		if s.peek() == equal {
			tok.Type = token.Eq
			s.read()
		}
	case langle:
		tok.Type = token.Lt
		if k := s.peek(); k == equal {
			tok.Type = token.Le
			s.read()
		} else if k == langle {
			tok.Type = token.Lshift
			s.read()
			if s.char == equal {
				tok.Type = token.LshiftAssign
			}
		}
	case rangle:
		tok.Type = token.Gt
		if k := s.peek(); k == equal {
			tok.Type = token.Ge
			s.read()
		} else if k == rangle {
			tok.Type = token.Rshift
			s.read()
			if s.char == equal {
				tok.Type = token.RshiftAssign
			}
		}
	case comma:
		tok.Type = token.Comma
	case lparen:
		tok.Type = token.Lparen
	case rparen:
		tok.Type = token.Rparen
	case lsquare:
		tok.Type = token.Lsquare
		if isNL(s.peek()) {
			s.read()
			s.skipNL()
		}
	case rsquare:
		tok.Type = token.Rsquare
	case lcurly:
		tok.Type = token.Lcurly
		if isNL(s.peek()) {
			s.read()
			s.skipNL()
		}
	case rcurly:
		tok.Type = token.Rcurly
	case plus:
		tok.Type = token.Add
		if s.peek() == equal {
			tok.Type = token.AddAssign
			s.read()
		}
	case minus:
		tok.Type = token.Sub
		if s.peek() == equal {
			tok.Type = token.SubAssign
			s.read()
		}
	case star:
		tok.Type = token.Mul
		if s.peek() == star {
			tok.Type = token.Pow
			s.read()
		} else if s.peek() == equal {
			tok.Type = token.MulAssign
			s.read()
		}
	case slash:
		tok.Type = token.Div
		if s.peek() == equal {
			tok.Type = token.DivAssign
			s.read()
		}
	case percent:
		tok.Type = token.Mod
		if s.peek() == equal {
			tok.Type = token.ModAssign
			s.read()
		}
	case semicolon:
		tok.Type = token.EOL
	case question:
		tok.Type = token.Ternary
	case colon:
		tok.Type = token.Colon
		if s.peek() == equal {
			tok.Type = token.Walrus
			s.read()
		}
	default:
		tok.Type = token.Invalid
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

func isBin(r rune) bool {
	return r == '0' || r == '1'
}

func isOctal(r rune) bool {
	return r >= '0' && r <= '7'
}

func isHex(r rune) bool {
	return isDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

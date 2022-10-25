package buddy

import (
	"fmt"
	"io"
	"strconv"
)

const MaxArity = 255

const (
	powLowest   = iota
	powAssign   // =
	powTernary  // ?:
	powRelation // &&, ||
	powEqual    // ==, !=
	powCompare  // <, <=, >, >=
	powAdd      // +, -
	powMul      // /, *, **, %
	powIndex
	powPrefix
	powCall // ()
	powDot
)

type powerMap map[rune]int

func (p powerMap) Get(r rune) int {
	v, ok := p[r]
	if !ok {
		return powLowest
	}
	return v
}

var powers = powerMap{
	Add:       powAdd,
	Sub:       powAdd,
	Mul:       powMul,
	Div:       powMul,
	Mod:       powMul,
	Pow:       powMul,
	Assign:    powAssign,
	AddAssign: powAssign,
	SubAssign: powAssign,
	MulAssign: powAssign,
	DivAssign: powAssign,
	ModAssign: powAssign,
	Lparen:    powCall,
	Ternary:   powTernary,
	And:       powRelation,
	Or:        powRelation,
	Eq:        powEqual,
	Ne:        powEqual,
	Lt:        powCompare,
	Le:        powCompare,
	Gt:        powCompare,
	Ge:        powCompare,
	Lsquare:   powIndex,
	Dot:       powDot,
}

type program struct {
	Expression
	symbols map[string]Expression
}

type parser struct {
	scan *Scanner
	curr Token
	peek Token

	prefix map[rune]func() (Expression, error)
	infix  map[rune]func(Expression) (Expression, error)
}

func Parse(r io.Reader) (Expression, error) {
	p := parser{
		scan: Scan(r),
	}
	p.prefix = map[rune]func() (Expression, error){
		Sub:     p.parsePrefix,
		Not:     p.parsePrefix,
		Number:  p.parsePrefix,
		Boolean: p.parsePrefix,
		Literal: p.parsePrefix,
		Ident:   p.parsePrefix,
		Lparen:  p.parseGroup,
		Lsquare: p.parseArray,
		Lcurly:  p.parseDict,
		Keyword: p.parseKeyword,
	}
	p.infix = map[rune]func(Expression) (Expression, error){
		Add:       p.parseInfix,
		Sub:       p.parseInfix,
		Mul:       p.parseInfix,
		Div:       p.parseInfix,
		Mod:       p.parseInfix,
		Pow:       p.parseInfix,
		Assign:    p.parseAssign,
		Dot:       p.parsePath,
		AddAssign: p.parseAssign,
		SubAssign: p.parseAssign,
		DivAssign: p.parseAssign,
		MulAssign: p.parseAssign,
		ModAssign: p.parseAssign,
		Lparen:    p.parseCall,
		Lsquare:   p.parseIndex,
		Ternary:   p.parseTernary,
		Eq:        p.parseInfix,
		Ne:        p.parseInfix,
		Lt:        p.parseInfix,
		Le:        p.parseInfix,
		Gt:        p.parseInfix,
		Ge:        p.parseInfix,
		And:       p.parseInfix,
		Or:        p.parseInfix,
	}
	p.next()
	p.next()
	return p.Parse()
}

func (p *parser) Parse() (Expression, error) {
	var s script
	s.symbols = make(map[string]Expression)
	for !p.done() {
		if ok, err := p.parseSpecial(&s); ok {
			if err != nil {
				return nil, err
			}
			continue
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		s.list = append(s.list, e)
		if err := p.eol(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (p *parser) parse(pow int) (Expression, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, fmt.Errorf("prefix: %s can not be parsed", p.curr)
	}
	left, err := fn()
	if err != nil {
		return nil, err
	}
	for (p.curr.Type != EOL || p.curr.Type != EOF) && pow < powers.Get(p.curr.Type) {
		fn, ok := p.infix[p.curr.Type]
		if !ok {
			return nil, fmt.Errorf("infix: %s can not be parsed", p.curr)
		}
		left, err = fn(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *parser) parseSpecial(s *script) (bool, error) {
	if p.curr.Type != Keyword {
		return false, nil
	}
	switch p.curr.Literal {
	default:
		return false, nil
	case kwDef:
		var (
			ident   = p.peek.Literal
			fn, err = p.parseFunction()
		)
		if err == nil {
			s.symbols[ident] = fn
			err = p.eol()
		}
		return true, err
	}
}

func (p *parser) parseKeyword() (Expression, error) {
	switch p.curr.Literal {
	case kwIf:
		return p.parseIf()
	case kwWhile:
		return p.parseWhile()
	case kwBreak:
		return p.parseBreak()
	case kwContinue:
		return p.parseContinue()
	case kwReturn:
		return p.parseReturn()
	case kwImport:
		return p.parseImport()
	default:
		return nil, fmt.Errorf("%s: keyword not implemented", p.curr.Literal)
	}
}

func (p *parser) parseImport() (Expression, error) {
	p.next()
	if p.curr.Type != Literal {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	m := module{
		ident: p.curr.Literal,
	}
	p.next()
	return m, nil
}

func (p *parser) parseParameters() ([]Expression, error) {
	if p.curr.Type != Lparen {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	p.next()

	var list []Expression
	for p.curr.Type != Rparen && !p.done() {
		if p.peek.Type == Assign {
			break
		}
		if p.curr.Type != Ident {
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
		a := createParameter(p.curr.Literal)
		list = append(list, a)
		p.next()
		switch p.curr.Type {
		case Comma:
			p.next()
		case Rparen:
		default:
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
	}
	for p.curr.Type != Rparen && !p.done() {
		if p.curr.Type != Ident {
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
		a := createParameter(p.curr.Literal)
		p.next()
		if p.curr.Type != Assign {
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
		p.next()
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		a.expr = expr
		list = append(list, a)
		switch p.curr.Type {
		case Comma:
			p.next()
		case Rparen:
		default:
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
	}
	if len(list) > MaxArity {
		return nil, fmt.Errorf("too many parameters given to function")
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	p.next()
	return list, nil
}

func (p *parser) parseFunction() (Expression, error) {
	p.next()
	fn := function{
		ident: p.curr.Literal,
	}
	p.next()
	args, err := p.parseParameters()
	if err != nil {
		return nil, err
	}
	fn.params = args

	fn.body, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	return fn, nil
}

func (p *parser) parseBlock() (Expression, error) {
	if p.curr.Type != Lcurly {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	var list []Expression
	for p.curr.Type != Rcurly && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
		if p.curr.Type != EOL {
			return nil, fmt.Errorf("syntax error! missing eol")
		}
		p.next()
	}
	if p.curr.Type != Rcurly {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	switch len(list) {
	case 1:
		return list[0], nil
	default:
		return script{list: list}, nil
	}
}

func (p *parser) parseIf() (Expression, error) {
	p.next()
	if p.curr.Type != Lparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()

	var (
		expr test
		err  error
	)
	expr.cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	expr.csq, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.curr.Type == Keyword && p.curr.Literal == kwElse {
		p.next()
		switch p.curr.Type {
		case Lcurly:
			expr.alt, err = p.parseBlock()
		case Keyword:
			expr.alt, err = p.parseKeyword()
		default:
		}
	}
	if p.curr.Type != EOL && p.curr.Type != EOF {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	return expr, nil
}

func (p *parser) parseWhile() (Expression, error) {
	p.next()
	if p.curr.Type != Lparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()

	var (
		expr while
		err  error
	)
	expr.cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("unexpected token %s", p.curr)
	}
	p.next()
	expr.body, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if p.curr.Type != EOL && p.curr.Type != EOF {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	return expr, nil
}

func (p *parser) parseReturn() (Expression, error) {
	p.next()
	if p.curr.Type == EOL || p.curr.Type == EOF {
		return returned{}, nil
	}
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr := returned{
		right: right,
	}
	return expr, nil
}

func (p *parser) parseBreak() (Expression, error) {
	p.next()
	return breaked{}, nil
}

func (p *parser) parseContinue() (Expression, error) {
	p.next()
	return continued{}, nil
}

func (p *parser) parseTernary(left Expression) (Expression, error) {
	var err error
	expr := test{
		cdt: left,
	}
	p.next()
	if expr.csq, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	if p.curr.Type != Colon {
		return nil, fmt.Errorf("syntax error! missing colon")
	}
	p.next()

	if expr.alt, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *parser) parsePath(left Expression) (Expression, error) {
	v, ok := left.(variable)
	if !ok {
		return nil, fmt.Errorf("can not used %T as path identifier")
	}
	p.next()
	a := path{
		ident: v.ident,
	}
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	a.right = right
	return a, nil
}

func (p *parser) parseAssign(left Expression) (Expression, error) {
	switch left.(type) {
	case variable, index:
	default:
		return nil, fmt.Errorf("can not assign to %T", left)
	}
	op := p.curr.Type
	p.next()
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr := assign{
		ident: left,
		right: right,
	}
	if op != Assign {
		switch op {
		case AddAssign:
			op = Add
		case SubAssign:
			op = Sub
		case MulAssign:
			op = Mul
		case DivAssign:
			op = Div
		case ModAssign:
			op = Mod
		default:
			return nil, fmt.Errorf("invalid compound assignment operator")
		}
		expr.right = binary{
			op:    op,
			left:  left,
			right: right,
		}
	}
	return expr, nil
}

func (p *parser) parseInfix(left Expression) (Expression, error) {
	expr := binary{
		op:   p.curr.Type,
		left: left,
	}
	pow := powers.Get(p.curr.Type)
	p.next()
	right, err := p.parse(pow)
	if err != nil {
		return nil, err
	}
	expr.right = right
	return expr, nil
}

func (p *parser) parseIndex(left Expression) (Expression, error) {
	switch left.(type) {
	case array, dict, index, variable:
	default:
		return nil, fmt.Errorf("%T can not be indexed", left)
	}
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != Rsquare {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	p.next()
	ix := index{
		arr:  left,
		expr: expr,
	}
	return ix, nil
}

func (p *parser) parseArray() (Expression, error) {
	p.next()
	var arr array
	for p.curr.Type != Rsquare && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		arr.list = append(arr.list, e)
		switch p.curr.Type {
		case Comma:
			p.next()
			p.skip(EOL)
		case Rsquare:
		default:
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
	}
	if p.curr.Type != Rsquare {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	p.next()
	return arr, nil
}

func (p *parser) parseDict() (Expression, error) {
	p.next()
	var d dict
	d.list = make(map[Expression]Expression)
	for p.curr.Type != Rcurly && !p.done() {
		k, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if p.curr.Type != Colon {
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
		p.next()
		v, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		d.list[k] = v
		switch p.curr.Type {
		case Comma:
			p.next()
			p.skip(EOL)
		case Rcurly:
		default:
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
	}
	if p.curr.Type != Rcurly {
		return nil, fmt.Errorf("unexpected token: %s", p.curr)
	}
	p.next()
	return d, nil
}

func (p *parser) parsePrefix() (Expression, error) {
	var expr Expression
	switch p.curr.Type {
	case Sub, Not:
		op := p.curr.Type
		p.next()

		right, err := p.parse(powPrefix)
		if err != nil {
			return nil, err
		}
		expr = unary{
			op:    op,
			right: right,
		}
	case Literal:
		expr = createLiteral(p.curr.Literal)
		p.next()
	case Number:
		n, err := strconv.ParseFloat(p.curr.Literal, 64)
		if err != nil {
			return nil, err
		}
		expr = createNumber(n)
		p.next()
	case Ident:
		expr = createVariable(p.curr.Literal)
		p.next()
	case Boolean:
		b, err := strconv.ParseBool(p.curr.Literal)
		if err != nil {
			return nil, err
		}
		expr = createBoolean(b)
		p.next()
	default:
		return nil, fmt.Errorf("unuspported token: %s", p.curr)
	}
	return expr, nil
}

func (p *parser) parseCall(left Expression) (Expression, error) {
	v, ok := left.(variable)
	if !ok {
		return nil, fmt.Errorf("syntax error! try to call non function")
	}
	p.next()
	expr := call{
		ident: v.ident,
	}
	for p.curr.Type != Rparen && !p.done() {
		if p.peek.Type == Assign {
			break
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		expr.args = append(expr.args, e)
		switch p.curr.Type {
		case Comma:
			if p.peek.Type == Rparen {
				return nil, fmt.Errorf("unexpected comma before closing paren")
			}
			p.next()
		case Rparen:
		default:
			return nil, fmt.Errorf("syntax error! missing comma")
		}
	}
	for p.curr.Type != Rparen && !p.done() {
		if p.curr.Type != Ident {
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
		a := createParameter(p.curr.Literal)
		p.next()
		if p.curr.Type != Assign {
			return nil, fmt.Errorf("unexpected token: %s", p.curr)
		}
		p.next()
		val, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		a.expr = val
		expr.args = append(expr.args, a)
		switch p.curr.Type {
		case Comma:
			if p.peek.Type == Rparen {
				return nil, fmt.Errorf("unexpected comma before closing paren")
			}
			p.next()
		case Rparen:
		default:
			return nil, fmt.Errorf("syntax error! missing comma")
		}
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("syntax error! missing closing )")
	}
	p.next()
	return expr, nil
}

func (p *parser) parseGroup() (Expression, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	if p.curr.Type != Rparen {
		return nil, fmt.Errorf("syntax error: missing closing )")
	}
	p.next()
	return expr, nil
}

func (p *parser) eol() error {
	switch p.curr.Type {
	case EOL:
		p.next()
	case EOF:
	default:
		return fmt.Errorf("syntax error! missing eol")
	}
	return nil
}

func (p *parser) skip(r rune) {
	for p.curr.Type == r {
		p.next()
	}
}

func (p *parser) done() bool {
	return p.curr.Type == EOF
}

func (p *parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

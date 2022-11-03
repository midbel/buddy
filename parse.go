package buddy

import (
	"fmt"
	"io"
	"strconv"

	"github.com/midbel/slices"
)

const MaxArity = 255

const (
	powLowest   = iota
	powAssign   // =
	powTernary  // ?:
	powBinary   // &, |, ^, ~
	powRelation // &&, ||
	powShift    // <<, >>
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
	Lshift:       powShift,
	Rshift:       powShift,
	BinAnd:       powBinary,
	BinOr:        powBinary,
	BinXor:       powBinary,
	Add:          powAdd,
	Sub:          powAdd,
	Mul:          powMul,
	Div:          powMul,
	Mod:          powMul,
	Pow:          powMul,
	Walrus:       powAssign,
	Assign:       powAssign,
	AddAssign:    powAssign,
	SubAssign:    powAssign,
	MulAssign:    powAssign,
	DivAssign:    powAssign,
	ModAssign:    powAssign,
	RshiftAssign: powAssign,
	LshiftAssign: powAssign,
	BinAndAssign: powAssign,
	BinOrAssign:  powAssign,
	BinXorAssign: powAssign,
	Lparen:       powCall,
	Ternary:      powTernary,
	And:          powRelation,
	Or:           powRelation,
	Eq:           powEqual,
	Ne:           powEqual,
	Lt:           powCompare,
	Le:           powCompare,
	Gt:           powCompare,
	Ge:           powCompare,
	Lsquare:      powIndex,
	Dot:          powDot,
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
		BinNot:  p.parsePrefix,
		Sub:     p.parsePrefix,
		Not:     p.parsePrefix,
		Double:  p.parsePrefix,
		Integer: p.parsePrefix,
		Boolean: p.parsePrefix,
		Literal: p.parsePrefix,
		Ident:   p.parsePrefix,
		Lparen:  p.parseGroup,
		Lsquare: p.parseArray,
		Lcurly:  p.parseDict,
		Keyword: p.parseKeyword,
	}
	p.infix = map[rune]func(Expression) (Expression, error){
		Add:          p.parseInfix,
		Sub:          p.parseInfix,
		Mul:          p.parseInfix,
		Div:          p.parseInfix,
		Mod:          p.parseInfix,
		Pow:          p.parseInfix,
		Lshift:       p.parseInfix,
		Rshift:       p.parseInfix,
		BinAnd:       p.parseInfix,
		BinOr:        p.parseInfix,
		BinXor:       p.parseInfix,
		Walrus:       p.parseWalrus,
		Assign:       p.parseAssign,
		Dot:          p.parsePath,
		AddAssign:    p.parseAssign,
		SubAssign:    p.parseAssign,
		DivAssign:    p.parseAssign,
		MulAssign:    p.parseAssign,
		ModAssign:    p.parseAssign,
		BinAndAssign: p.parseAssign,
		BinOrAssign:  p.parseAssign,
		BinXorAssign: p.parseAssign,
		LshiftAssign: p.parseAssign,
		RshiftAssign: p.parseAssign,
		Lparen:       p.parseCall,
		Lsquare:      p.parseIndex,
		Ternary:      p.parseTernary,
		Eq:           p.parseInfix,
		Ne:           p.parseInfix,
		Lt:           p.parseInfix,
		Le:           p.parseInfix,
		Gt:           p.parseInfix,
		Ge:           p.parseInfix,
		And:          p.parseInfix,
		Or:           p.parseInfix,
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
	left, err := p.getPrefixExpr()
	if err != nil {
		return nil, err
	}
	for (!p.is(EOL) || !p.is(EOF)) && pow < powers.Get(p.curr.Type) {
		left, err = p.getInfixExpr(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *parser) parseSpecial(s *script) (bool, error) {
	if err := p.expectKW(kwDef, ""); err != nil {
		return false, nil
	}
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
	case kwFrom:
		return p.parseFrom()
	case kwFor:
		return p.parseFor()
	case kwAssert:
		return p.parseAssert()
	default:
		return nil, p.parseError("keyword not recognized")
	}
}

func (p *parser) parseAssert() (Expression, error) {
	p.next()
	var (
		ass assert
		err error
	)
	ass.expr, err = p.parse(powLowest)
	return ass, err
}

func (p *parser) parseFor() (Expression, error) {
	p.next()

	var (
		loop forloop
		err  error
	)
	if !p.is(EOL) {
		loop.init, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		switch e := loop.init.(type) {
		case variable:
			return p.parseForeach(e.ident)
		case assign:
		case walrus:
		default:
			return nil, p.parseError("illegal expression! assignment expected")
		}
	}
	if err := p.expect(EOL, "expected ';'"); err != nil {
		return nil, err
	}
	p.next()
	if !p.is(EOL) {
		loop.cdt, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
	}
	if err := p.expect(EOL, "expected ';'"); err != nil {
		return nil, err
	}
	p.next()
	if !p.is(Lcurly) {
		loop.incr, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
	}
	loop.body, err = p.parseBlock()
	if !p.is(EOL) && !p.is(EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return loop, nil
}

func (p *parser) parseForeach(ident string) (Expression, error) {
	var (
		expr foreach
		err  error
	)
	expr.ident = ident
	if err := p.expectKW(kwIn, "expected 'in' keyword"); err != nil {
		return nil, err
	}
	p.next()
	if expr.iter, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	expr.body, err = p.parseBlock()
	if !p.is(EOL) && !p.is(EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *parser) parseFrom() (Expression, error) {
	p.next()
	var mod module
	for p.is(Ident) {
		mod.ident = append(mod.ident, p.curr.Literal)
		p.next()
		switch p.curr.Type {
		case Dot:
			p.next()
		case Keyword:
		default:
			return nil, p.parseError("expected keyword, newline, '.' or ';'")
		}
	}
	if len(mod.ident) == 0 {
		return nil, p.parseError("no identifier given for import")
	}
	if err := p.expectKW(kwImport, "expected 'import' keyword"); err != nil {
		return nil, err
	}
	p.next()
	for !p.is(EOL) && !p.done() {
		if err := p.expect(Ident, "expected identifier"); err != nil {
			return nil, err
		}
		s := createSymbol(p.curr.Literal)
		p.next()
		if err := p.expectKW(kwAs, ""); err == nil {
			p.next()
			if err := p.expect(Ident, "expected identifier"); err != nil {
				return nil, err
			}
			s.alias = p.curr.Literal
			p.next()
		}
		mod.symbols = append(mod.symbols, s)
		switch p.curr.Type {
		case Comma:
			if p.peekIs(EOL) || p.peekIs(EOF) {
				return nil, p.parseError("unexpected ',' before end of line")
			}
			p.next()
		case EOL, EOF:
		default:
			return nil, p.parseError("expected newline, ',' or ;'")
		}
	}
	return mod, nil
}

func (p *parser) parseImport() (Expression, error) {
	p.next()
	var mod module
	for p.is(Ident) {
		mod.ident = append(mod.ident, p.curr.Literal)
		p.next()
		switch p.curr.Type {
		case Dot:
			p.next()
		case Keyword, EOL, EOF:
		default:
			return nil, p.parseError("expected keyword, newline, '.' or ';'")
		}
	}
	if len(mod.ident) == 0 {
		return nil, p.parseError("no identifier given for import")
	}
	mod.alias = slices.Lst(mod.ident)
	if err := p.expectKW(kwAs, ""); err == nil {
		p.next()
		if err := p.expect(Ident, "expected identifier"); err != nil {
			return nil, err
		}
		mod.alias = p.curr.Literal
		p.next()
	}
	return mod, nil
}

func (p *parser) parseParameters() ([]Expression, error) {
	if err := p.expect(Lparen, "expected ')'"); err != nil {
		return nil, err
	}
	p.next()

	var list []Expression
	for !p.is(Rparen) && !p.done() {
		if p.peekIs(Assign) {
			break
		}
		if err := p.expect(Ident, "expected identifier"); err != nil {
			return nil, err
		}
		a := createParameter(p.curr.Literal)
		list = append(list, a)
		p.next()
		switch p.curr.Type {
		case Comma:
			if p.peekIs(Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case Rparen:
		default:
			return nil, p.parseError("expected ')' or ','")
		}
	}
	for !p.is(Rparen) && !p.done() {
		if err := p.expect(Ident, "expected identifier"); err != nil {
			return nil, err
		}
		a := createParameter(p.curr.Literal)
		p.next()
		if err := p.expect(Assign, "expected '='"); err != nil {
			return nil, err
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
			if p.peekIs(Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case Rparen:
		default:
			return nil, p.parseError("expected ')' or ','")
		}
	}
	if len(list) > MaxArity {
		return nil, p.parseError("too many parameters given to function")
	}
	if err := p.expect(Rparen, "expected ')"); err != nil {
		return nil, err
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
	if err := p.expect(Lcurly, "expected '{"); err != nil {
		return nil, err
	}
	p.next()
	var list []Expression
	for !p.is(Rcurly) && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
		if err := p.expect(EOL, "expected newline or ';'"); err != nil {
			return nil, err
		}
		p.next()
	}
	if err := p.expect(Rcurly, "expected '}"); err != nil {
		return nil, err
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

	var (
		expr test
		err  error
	)
	expr.cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr.csq, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if err := p.expectKW(kwElse, ""); err == nil {
		p.next()
		switch p.curr.Type {
		case Lcurly:
			expr.alt, err = p.parseBlock()
		case Keyword:
			expr.alt, err = p.parseKeyword()
		default:
		}
	}
	if !p.is(EOL) && !p.is(EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *parser) parseWhile() (Expression, error) {
	p.next()

	var (
		expr while
		err  error
	)
	expr.cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr.body, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if !p.is(EOL) && !p.is(EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *parser) parseReturn() (Expression, error) {
	p.next()
	if p.is(EOL) || p.is(EOF) {
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
	if err := p.expect(Colon, "expected ':'"); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("unexpected path operator")
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

func (p *parser) parseWalrus(left Expression) (Expression, error) {
	switch left.(type) {
	case variable, index:
	default:
		return nil, fmt.Errorf("unexpected assignment operator")
	}
	p.next()
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr := walrus{
		assign: createAssign(left, right),
	}
	return expr, nil
}

func (p *parser) parseAssign(left Expression) (Expression, error) {
	switch left.(type) {
	case variable, index:
	default:
		return nil, fmt.Errorf("unexpected assignment operator")
	}
	op := p.curr.Type
	p.next()
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr := createAssign(left, right)
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
		case BinAndAssign:
			op = BinAnd
		case BinOrAssign:
			op = BinOr
		case BinXorAssign:
			op = BinXor
		case LshiftAssign:
			op = Lshift
		case RshiftAssign:
			op = Rshift
		default:
			return nil, p.parseError("compound assignment operator not recognized")
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
	case array, dict, index, variable, literal:
	default:
		return nil, p.parseError("unexpected index operator")
	}
	ix := index{
		arr: left,
	}
	p.next()
	for !p.is(Rsquare) && !p.done() {
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		ix.list = append(ix.list, expr)
		switch p.curr.Type {
		case Comma:
			if p.peekIs(Rsquare) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case Rsquare:
		default:
			return nil, p.parseError("expected ',' or ']")
		}
	}
	if err := p.expect(Rsquare, "expected ']'"); err != nil {
		return nil, err
	}
	if len(ix.list) == 0 {
		return nil, p.parseError("empty index")
	}
	p.next()
	return ix, nil
}

func (p *parser) parseCompitem(until rune) ([]compitem, error) {
	var list []compitem
	p.next()
	for !p.is(until) && !p.done() {
		var item compitem
		if err := p.expect(Ident, "expected identifier"); err != nil {
			return nil, err
		}
		item.ident = p.curr.Literal
		p.next()
		if err := p.expectKW(kwIn, "expected 'in' keyword"); err != nil {
			return nil, err
		}
		p.next()
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		item.iter = expr
		for p.curr.Type == Keyword && p.curr.Literal == kwIf {
			p.next()
			expr, err := p.parse(powLowest)
			if err != nil {
				return nil, err
			}
			item.cdt = append(item.cdt, expr)
		}
		switch p.curr.Type {
		case Keyword:
			if p.curr.Literal != kwFor {
				return nil, fmt.Errorf("expected 'for' keyword")
			}
			p.next()
		case until:
		default:
			return nil, fmt.Errorf("expected ']' or 'for' keyword")
		}
		list = append(list, item)
	}
	if err := p.expect(until, "unexpected token"); err != nil {
		return nil, err
	}
	p.next()
	return list, nil
}

func (p *parser) parseListcomp(left Expression) (Expression, error) {
	cmp := listcomp{
		body: left,
	}
	list, err := p.parseCompitem(Rsquare)
	if err == nil {
		cmp.list = list
	}
	return cmp, err
}

func (p *parser) parseArray() (Expression, error) {
	p.next()
	var arr array
	for !p.is(Rsquare) && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if err := p.expectKW(kwFor, ""); len(arr.list) == 0 && err == nil {
			return p.parseListcomp(e)
		}
		arr.list = append(arr.list, e)
		switch p.curr.Type {
		case Comma:
			p.next()
			p.skip(EOL)
		case Rsquare:
		default:
			return nil, p.parseError("expected ',' or ']")
		}
	}
	if err := p.expect(Rsquare, "expected ']'"); err != nil {
		return nil, err
	}
	p.next()
	return arr, nil
}

func (p *parser) parseDictcomp(key, val Expression) (Expression, error) {
	cmp := dictcomp{
		key: key,
		val: val,
	}
	list, err := p.parseCompitem(Rcurly)
	if err == nil {
		cmp.list = list
	}
	return cmp, err
}

func (p *parser) parseDict() (Expression, error) {
	p.next()
	var d dict
	d.list = make(map[Expression]Expression)
	for !p.is(Rcurly) && !p.done() {
		k, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if err := p.expect(Colon, "expected ':'"); err != nil {
			return nil, err
		}
		p.next()
		v, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if err := p.expectKW(kwFor, ""); len(d.list) == 0 && err == nil {
			return p.parseDictcomp(k, v)
		}
		d.list[k] = v
		switch p.curr.Type {
		case Comma:
			p.next()
			p.skip(EOL)
		case Rcurly:
		default:
			return nil, p.parseError("expected ',' or '}")
		}
	}
	if err := p.expect(Rcurly, "expected '}'"); err != nil {
		return nil, err
	}
	p.next()
	return d, nil
}

func (p *parser) parsePrefix() (Expression, error) {
	var expr Expression
	switch p.curr.Type {
	case Sub, Not, BinNot:
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
	case Double:
		n, err := strconv.ParseFloat(p.curr.Literal, 64)
		if err != nil {
			return nil, err
		}
		expr = createDouble(n)
		p.next()
	case Integer:
		n, err := strconv.ParseInt(p.curr.Literal, 0, 64)
		if err != nil {
			return nil, err
		}
		expr = createInteger(n)
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
		return nil, p.parseError("prefix operator not recognized")
	}
	return expr, nil
}

func (p *parser) parseCall(left Expression) (Expression, error) {
	v, ok := left.(variable)
	if !ok {
		return nil, p.parseError("unexpected call operator")
	}
	p.next()
	expr := call{
		ident: v.ident,
	}
	for !p.is(Rparen) && !p.done() {
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
			if p.peekIs(Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case Rparen:
		default:
			return nil, p.parseError("expected ','")
		}
	}
	for !p.is(Rparen) && !p.done() {
		if err := p.expect(Ident, "expected identifier"); err != nil {
			return nil, err
		}
		a := createParameter(p.curr.Literal)
		p.next()
		if err := p.expect(Assign, "expected '='"); err != nil {
			return nil, err
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
			if p.peekIs(Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case Rparen:
		default:
			return nil, p.parseError("expected ','")
		}
	}
	if err := p.expect(Rparen, "expected ')'"); err != nil {
		return nil, err
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
	if err := p.expect(Rparen, "expected ')'"); err != nil {
		return nil, err
	}
	p.next()
	return expr, nil
}

func (p *parser) getPrefixExpr() (Expression, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, p.parseError("unary operator not recognized")
	}
	return fn()
}

func (p *parser) getInfixExpr(left Expression) (Expression, error) {
	fn, ok := p.infix[p.curr.Type]
	if !ok {
		return nil, p.parseError("binary operator not recognized")
	}
	return fn(left)
}

func (p *parser) peekIs(r rune) bool {
	return p.peek.Type == r
}

func (p *parser) is(r rune) bool {
	return p.curr.Type == r
}

func (p *parser) expect(r rune, msg string) error {
	var err error
	if !p.is(r) {
		err = p.parseError(msg)
	}
	return err
}

func (p *parser) expectKW(kw, msg string) error {
	var err error
	if err = p.expect(Keyword, msg); err != nil {
		return err
	}
	if p.curr.Literal != kw {
		err = p.parseError(msg)
	}
	return err
}

func (p *parser) eol() error {
	switch p.curr.Type {
	case EOL:
		p.next()
	case EOF:
	default:
		return p.parseError("expected newline or ';'")
	}
	return nil
}

func (p *parser) skip(r rune) {
	for p.is(r) {
		p.next()
	}
}

func (p *parser) done() bool {
	return p.is(EOF)
}

func (p *parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *parser) parseError(message string) error {
	return ParseError{
		Token:   p.curr,
		Line:    p.scan.getLine(p.curr.Position),
		Message: message,
	}
}

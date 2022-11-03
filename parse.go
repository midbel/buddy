package buddy

import (
	"fmt"
	"io"
	"strconv"

	"github.com/midbel/buddy/scan"
	"github.com/midbel/buddy/token"
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
	token.Lshift:       powShift,
	token.Rshift:       powShift,
	token.BinAnd:       powBinary,
	token.BinOr:        powBinary,
	token.BinXor:       powBinary,
	token.Add:          powAdd,
	token.Sub:          powAdd,
	token.Mul:          powMul,
	token.Div:          powMul,
	token.Mod:          powMul,
	token.Pow:          powMul,
	token.Walrus:       powAssign,
	token.Assign:       powAssign,
	token.AddAssign:    powAssign,
	token.SubAssign:    powAssign,
	token.MulAssign:    powAssign,
	token.DivAssign:    powAssign,
	token.ModAssign:    powAssign,
	token.RshiftAssign: powAssign,
	token.LshiftAssign: powAssign,
	token.BinAndAssign: powAssign,
	token.BinOrAssign:  powAssign,
	token.BinXorAssign: powAssign,
	token.Lparen:       powCall,
	token.Ternary:      powTernary,
	token.And:          powRelation,
	token.Or:           powRelation,
	token.Eq:           powEqual,
	token.Ne:           powEqual,
	token.Lt:           powCompare,
	token.Le:           powCompare,
	token.Gt:           powCompare,
	token.Ge:           powCompare,
	token.Lsquare:      powIndex,
	token.Dot:          powDot,
}

type parser struct {
	file string
	scan *scan.Scanner
	curr token.Token
	peek token.Token

	prefix map[rune]func() (Expression, error)
	infix  map[rune]func(Expression) (Expression, error)
}

func Parse(r io.Reader) (Expression, error) {
	p := parser{
		scan: scan.Scan(r),
	}
	if n, ok := r.(interface{ Name() string }); ok {
		p.file = n.Name()
	}
	p.prefix = map[rune]func() (Expression, error){
		token.BinNot:  p.parsePrefix,
		token.Sub:     p.parsePrefix,
		token.Not:     p.parsePrefix,
		token.Double:  p.parsePrefix,
		token.Integer: p.parsePrefix,
		token.Boolean: p.parsePrefix,
		token.Literal: p.parsePrefix,
		token.Ident:   p.parsePrefix,
		token.Lparen:  p.parseGroup,
		token.Lsquare: p.parseArray,
		token.Lcurly:  p.parseDict,
		token.Keyword: p.parseKeyword,
	}
	p.infix = map[rune]func(Expression) (Expression, error){
		token.Add:          p.parseInfix,
		token.Sub:          p.parseInfix,
		token.Mul:          p.parseInfix,
		token.Div:          p.parseInfix,
		token.Mod:          p.parseInfix,
		token.Pow:          p.parseInfix,
		token.Lshift:       p.parseInfix,
		token.Rshift:       p.parseInfix,
		token.BinAnd:       p.parseInfix,
		token.BinOr:        p.parseInfix,
		token.BinXor:       p.parseInfix,
		token.Walrus:       p.parseWalrus,
		token.Assign:       p.parseAssign,
		token.Dot:          p.parsePath,
		token.AddAssign:    p.parseAssign,
		token.SubAssign:    p.parseAssign,
		token.DivAssign:    p.parseAssign,
		token.MulAssign:    p.parseAssign,
		token.ModAssign:    p.parseAssign,
		token.BinAndAssign: p.parseAssign,
		token.BinOrAssign:  p.parseAssign,
		token.BinXorAssign: p.parseAssign,
		token.LshiftAssign: p.parseAssign,
		token.RshiftAssign: p.parseAssign,
		token.Lparen:       p.parseCall,
		token.Lsquare:      p.parseIndex,
		token.Ternary:      p.parseTernary,
		token.Eq:           p.parseInfix,
		token.Ne:           p.parseInfix,
		token.Lt:           p.parseInfix,
		token.Le:           p.parseInfix,
		token.Gt:           p.parseInfix,
		token.Ge:           p.parseInfix,
		token.And:          p.parseInfix,
		token.Or:           p.parseInfix,
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
	for (!p.is(token.EOL) || !p.is(token.EOF)) && pow < powers.Get(p.curr.Type) {
		left, err = p.getInfixExpr(left)
		if err != nil {
			return nil, err
		}
	}
	return left, nil
}

func (p *parser) parseSpecial(s *script) (bool, error) {
	if err := p.expectKW(token.KwDef, ""); err != nil {
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
	case token.KwIf:
		return p.parseIf()
	case token.KwWhile:
		return p.parseWhile()
	case token.KwBreak:
		return p.parseBreak()
	case token.KwContinue:
		return p.parseContinue()
	case token.KwReturn:
		return p.parseReturn()
	case token.KwImport:
		return p.parseImport()
	case token.KwFrom:
		return p.parseFrom()
	case token.KwFor:
		return p.parseFor()
	case token.KwAssert:
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
	if !p.is(token.EOL) {
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
	if err := p.expect(token.EOL, "expected ';'"); err != nil {
		return nil, err
	}
	p.next()
	if !p.is(token.EOL) {
		loop.cdt, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
	}
	if err := p.expect(token.EOL, "expected ';'"); err != nil {
		return nil, err
	}
	p.next()
	if !p.is(token.Lcurly) {
		loop.incr, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
	}
	loop.body, err = p.parseBlock()
	if !p.is(token.EOL) && !p.is(token.EOF) {
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
	if err := p.expectKW(token.KwIn, "expected 'in' keyword"); err != nil {
		return nil, err
	}
	p.next()
	if expr.iter, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	expr.body, err = p.parseBlock()
	if !p.is(token.EOL) && !p.is(token.EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *parser) parseFrom() (Expression, error) {
	p.next()
	var mod module
	for p.is(token.Ident) {
		mod.ident = append(mod.ident, p.curr.Literal)
		p.next()
		switch p.curr.Type {
		case token.Dot:
			p.next()
		case token.Keyword:
		default:
			return nil, p.parseError("expected keyword, newline, '.' or ';'")
		}
	}
	if len(mod.ident) == 0 {
		return nil, p.parseError("no identifier given for import")
	}
	if err := p.expectKW(token.KwImport, "expected 'import' keyword"); err != nil {
		return nil, err
	}
	p.next()
	for !p.is(token.EOL) && !p.done() {
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		s := createSymbol(p.curr.Literal)
		p.next()
		if err := p.expectKW(token.KwAs, ""); err == nil {
			p.next()
			if err := p.expect(token.Ident, "expected identifier"); err != nil {
				return nil, err
			}
			s.alias = p.curr.Literal
			p.next()
		}
		mod.symbols = append(mod.symbols, s)
		switch p.curr.Type {
		case token.Comma:
			if p.peekIs(token.EOL) || p.peekIs(token.EOF) {
				return nil, p.parseError("unexpected ',' before end of line")
			}
			p.next()
		case token.EOL, token.EOF:
		default:
			return nil, p.parseError("expected newline, ',' or ;'")
		}
	}
	return mod, nil
}

func (p *parser) parseImport() (Expression, error) {
	p.next()
	var mod module
	for p.is(token.Ident) {
		mod.ident = append(mod.ident, p.curr.Literal)
		p.next()
		switch p.curr.Type {
		case token.Dot:
			p.next()
		case token.Keyword, token.EOL, token.EOF:
		default:
			return nil, p.parseError("expected keyword, newline, '.' or ';'")
		}
	}
	if len(mod.ident) == 0 {
		return nil, p.parseError("no identifier given for import")
	}
	mod.alias = slices.Lst(mod.ident)
	if err := p.expectKW(token.KwAs, ""); err == nil {
		p.next()
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		mod.alias = p.curr.Literal
		p.next()
	}
	return mod, nil
}

func (p *parser) parseParameters() ([]Expression, error) {
	if err := p.expect(token.Lparen, "expected ')'"); err != nil {
		return nil, err
	}
	p.next()

	var list []Expression
	for !p.is(token.Rparen) && !p.done() {
		if p.peekIs(token.Assign) {
			break
		}
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		a := createParameter(p.curr.Literal)
		list = append(list, a)
		p.next()
		switch p.curr.Type {
		case token.Comma:
			if p.peekIs(token.Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case token.Rparen:
		default:
			return nil, p.parseError("expected ')' or ','")
		}
	}
	for !p.is(token.Rparen) && !p.done() {
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		a := createParameter(p.curr.Literal)
		p.next()
		if err := p.expect(token.Assign, "expected '='"); err != nil {
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
		case token.Comma:
			if p.peekIs(token.Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case token.Rparen:
		default:
			return nil, p.parseError("expected ')' or ','")
		}
	}
	if len(list) > MaxArity {
		return nil, p.parseError("too many parameters given to function")
	}
	if err := p.expect(token.Rparen, "expected ')"); err != nil {
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
	if err := p.expect(token.Lcurly, "expected '{"); err != nil {
		return nil, err
	}
	p.next()
	var list []Expression
	for !p.is(token.Rcurly) && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		list = append(list, e)
		if err := p.expect(token.EOL, "expected newline or ';'"); err != nil {
			return nil, err
		}
		p.next()
	}
	if err := p.expect(token.Rcurly, "expected '}"); err != nil {
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
	if err := p.expectKW(token.KwElse, ""); err == nil {
		p.next()
		switch p.curr.Type {
		case token.Lcurly:
			expr.alt, err = p.parseBlock()
		case token.Keyword:
			expr.alt, err = p.parseKeyword()
		default:
		}
	}
	if !p.is(token.EOL) && !p.is(token.EOF) {
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
	if !p.is(token.EOL) && !p.is(token.EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *parser) parseReturn() (Expression, error) {
	p.next()
	if p.is(token.EOL) || p.is(token.EOF) {
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
	if err := p.expect(token.Colon, "expected ':'"); err != nil {
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
	if op != token.Assign {
		switch op {
		case token.AddAssign:
			op = token.Add
		case token.SubAssign:
			op = token.Sub
		case token.MulAssign:
			op = token.Mul
		case token.DivAssign:
			op = token.Div
		case token.ModAssign:
			op = token.Mod
		case token.BinAndAssign:
			op = token.BinAnd
		case token.BinOrAssign:
			op = token.BinOr
		case token.BinXorAssign:
			op = token.BinXor
		case token.LshiftAssign:
			op = token.Lshift
		case token.RshiftAssign:
			op = token.Rshift
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
	for !p.is(token.Rsquare) && !p.done() {
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		ix.list = append(ix.list, expr)
		switch p.curr.Type {
		case token.Comma:
			if p.peekIs(token.Rsquare) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case token.Rsquare:
		default:
			return nil, p.parseError("expected ',' or ']")
		}
	}
	if err := p.expect(token.Rsquare, "expected ']'"); err != nil {
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
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		item.ident = p.curr.Literal
		p.next()
		if err := p.expectKW(token.KwIn, "expected 'in' keyword"); err != nil {
			return nil, err
		}
		p.next()
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		item.iter = expr
		for p.is(token.Keyword) && p.curr.Literal == token.KwIf {
			p.next()
			expr, err := p.parse(powLowest)
			if err != nil {
				return nil, err
			}
			item.cdt = append(item.cdt, expr)
		}
		switch p.curr.Type {
		case token.Keyword:
			if p.curr.Literal != token.KwFor {
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
	list, err := p.parseCompitem(token.Rsquare)
	if err == nil {
		cmp.list = list
	}
	return cmp, err
}

func (p *parser) parseArray() (Expression, error) {
	p.next()
	var arr array
	for !p.is(token.Rsquare) && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if err := p.expectKW(token.KwFor, ""); len(arr.list) == 0 && err == nil {
			return p.parseListcomp(e)
		}
		arr.list = append(arr.list, e)
		switch p.curr.Type {
		case token.Comma:
			p.next()
			p.skip(token.EOL)
		case token.Rsquare:
		default:
			return nil, p.parseError("expected ',' or ']")
		}
	}
	if err := p.expect(token.Rsquare, "expected ']'"); err != nil {
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
	list, err := p.parseCompitem(token.Rcurly)
	if err == nil {
		cmp.list = list
	}
	return cmp, err
}

func (p *parser) parseDict() (Expression, error) {
	p.next()
	var d dict
	d.list = make(map[Expression]Expression)
	for !p.is(token.Rcurly) && !p.done() {
		k, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if err := p.expect(token.Colon, "expected ':'"); err != nil {
			return nil, err
		}
		p.next()
		v, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if err := p.expectKW(token.KwFor, ""); len(d.list) == 0 && err == nil {
			return p.parseDictcomp(k, v)
		}
		d.list[k] = v
		switch p.curr.Type {
		case token.Comma:
			p.next()
			p.skip(token.EOL)
		case token.Rcurly:
		default:
			return nil, p.parseError("expected ',' or '}")
		}
	}
	if err := p.expect(token.Rcurly, "expected '}'"); err != nil {
		return nil, err
	}
	p.next()
	return d, nil
}

func (p *parser) parsePrefix() (Expression, error) {
	var expr Expression
	switch p.curr.Type {
	case token.Sub, token.Not, token.BinNot:
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
	case token.Literal:
		expr = createLiteral(p.curr.Literal)
		p.next()
	case token.Double:
		n, err := strconv.ParseFloat(p.curr.Literal, 64)
		if err != nil {
			return nil, err
		}
		expr = createDouble(n)
		p.next()
	case token.Integer:
		n, err := strconv.ParseInt(p.curr.Literal, 0, 64)
		if err != nil {
			return nil, err
		}
		expr = createInteger(n)
		p.next()
	case token.Ident:
		expr = createVariable(p.curr.Literal)
		p.next()
	case token.Boolean:
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
	for !p.is(token.Rparen) && !p.done() {
		if p.peekIs(token.Assign) {
			break
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		expr.args = append(expr.args, e)
		switch p.curr.Type {
		case token.Comma:
			if p.peekIs(token.Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case token.Rparen:
		default:
			return nil, p.parseError("expected ','")
		}
	}
	for !p.is(token.Rparen) && !p.done() {
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		a := createParameter(p.curr.Literal)
		p.next()
		if err := p.expect(token.Assign, "expected '='"); err != nil {
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
		case token.Comma:
			if p.peekIs(token.Rparen) {
				return nil, p.parseError("unexpected ',' before ')")
			}
			p.next()
		case token.Rparen:
		default:
			return nil, p.parseError("expected ','")
		}
	}
	if err := p.expect(token.Rparen, "expected ')'"); err != nil {
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
	if err := p.expect(token.Rparen, "expected ')'"); err != nil {
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
	if err = p.expect(token.Keyword, msg); err != nil {
		return err
	}
	if p.curr.Literal != kw {
		err = p.parseError(msg)
	}
	return err
}

func (p *parser) eol() error {
	switch p.curr.Type {
	case token.EOL:
		p.next()
	case token.EOF:
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
	return p.is(token.EOF)
}

func (p *parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *parser) parseError(message string) error {
	return ParseError{
		Token:   p.curr,
		File:    p.file,
		Line:    p.scan.CurrentLine(p.curr.Position),
		Message: message,
	}
}

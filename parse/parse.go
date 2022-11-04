package parse

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/scan"
	"github.com/midbel/buddy/token"
)

const MaxArity = 255

type ParseError struct {
	token.Token
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

type Parser struct {
	file string
	scan *scan.Scanner
	curr token.Token
	peek token.Token

	prefix map[rune]func() (ast.Expression, error)
	infix  map[rune]func(ast.Expression) (ast.Expression, error)
}

func New(r io.Reader) *Parser {
	p := Parser{
		scan: scan.Scan(r),
	}
	if n, ok := r.(interface{ Name() string }); ok {
		p.file = n.Name()
	}
	p.prefix = map[rune]func() (ast.Expression, error){
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
	p.infix = map[rune]func(ast.Expression) (ast.Expression, error){
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
	return &p
}

func (p *Parser) Parse() (ast.Expression, error) {
	s := ast.CreateScript()
	for !p.done() {
		if p.is(token.Comment) {
			p.next()
			continue
		}
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
		s.List = append(s.List, e)
		if err := p.eol(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (p *Parser) parse(pow int) (ast.Expression, error) {
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

func (p *Parser) parseSpecial(s *ast.Script) (bool, error) {
	if err := p.expectKW(token.KwDef, ""); err != nil {
		return false, nil
	}
	var (
		ident   = p.peek.Literal
		fn, err = p.parseFunction()
	)
	if err == nil {
		s.Symbols[ident] = fn
		err = p.eol()
	}
	return true, err
}

func (p *Parser) parseKeyword() (ast.Expression, error) {
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

func (p *Parser) parseAssert() (ast.Expression, error) {
	p.next()
	expr, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	return ast.CreateAssert(expr), err
}

func (p *Parser) parseFor() (ast.Expression, error) {
	p.next()

	var (
		loop ast.For
		err  error
	)
	if !p.is(token.EOL) {
		loop.Init, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		switch e := loop.Init.(type) {
		case ast.Variable:
			return p.parseForeach(e.Ident)
		case ast.Assign:
		default:
			return nil, p.parseError("illegal expression! assignment expected")
		}
	}
	if err := p.expect(token.EOL, "expected ';'"); err != nil {
		return nil, err
	}
	p.next()
	if !p.is(token.EOL) {
		loop.Cdt, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
	}
	if err := p.expect(token.EOL, "expected ';'"); err != nil {
		return nil, err
	}
	p.next()
	if !p.is(token.Lcurly) {
		loop.Incr, err = p.parse(powLowest)
		if err != nil {
			return nil, err
		}
	}
	loop.Body, err = p.parseBlock()
	if !p.is(token.EOL) && !p.is(token.EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return loop, nil
}

func (p *Parser) parseForeach(ident string) (ast.Expression, error) {
	var (
		expr ast.ForEach
		err  error
	)
	expr.Ident = ident
	if err := p.expectKW(token.KwIn, "expected 'in' keyword"); err != nil {
		return nil, err
	}
	p.next()
	if expr.Iter, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	expr.Body, err = p.parseBlock()
	if !p.is(token.EOL) && !p.is(token.EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *Parser) parseFrom() (ast.Expression, error) {
	p.next()
	var imp ast.Import
	for p.is(token.Ident) {
		imp.Ident = append(imp.Ident, p.curr.Literal)
		p.next()
		switch p.curr.Type {
		case token.Dot:
			p.next()
		case token.Keyword:
		default:
			return nil, p.parseError("expected keyword, newline, '.' or ';'")
		}
	}
	if len(imp.Ident) == 0 {
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
		s := ast.CreateSymbol(p.curr.Literal)
		p.next()
		if err := p.expectKW(token.KwAs, ""); err == nil {
			p.next()
			if err := p.expect(token.Ident, "expected identifier"); err != nil {
				return nil, err
			}
			s.Alias = p.curr.Literal
			p.next()
		}
		imp.Symbols = append(imp.Symbols, s)
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
	return imp, nil
}

func (p *Parser) parseImport() (ast.Expression, error) {
	p.next()
	var imp ast.Import
	for p.is(token.Ident) {
		imp.Ident = append(imp.Ident, p.curr.Literal)
		imp.Alias = p.curr.Literal
		p.next()
		switch p.curr.Type {
		case token.Dot:
			p.next()
		case token.Keyword, token.EOL, token.EOF:
		default:
			return nil, p.parseError("expected keyword, newline, '.' or ';'")
		}
	}
	if err := p.expectKW(token.KwAs, ""); err == nil {
		p.next()
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		imp.Alias = p.curr.Literal
		p.next()
	}
	return imp, nil
}

func (p *Parser) parseParameters() ([]ast.Expression, error) {
	if err := p.expect(token.Lparen, "expected ')'"); err != nil {
		return nil, err
	}
	p.next()

	var list []ast.Expression
	for !p.is(token.Rparen) && !p.done() {
		if p.peekIs(token.Assign) {
			break
		}
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		a := ast.CreateParameter(p.curr.Literal)
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
		a := ast.CreateParameter(p.curr.Literal)
		p.next()
		if err := p.expect(token.Assign, "expected '='"); err != nil {
			return nil, err
		}
		p.next()
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		a.Expr = expr
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

func (p *Parser) parseFunction() (ast.Expression, error) {
	p.next()
	fn := ast.CreateFunction(p.curr.Literal)
	p.next()
	args, err := p.parseParameters()
	if err != nil {
		return nil, err
	}
	fn.Params = args

	if fn.Body, err = p.parseBlock(); err != nil {
		return nil, err
	}
	return fn, nil
}

func (p *Parser) parseBlock() (ast.Expression, error) {
	if err := p.expect(token.Lcurly, "expected '{"); err != nil {
		return nil, err
	}
	p.next()
	var list []ast.Expression
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
	return ast.CreateScriptFromList(list), nil
}

func (p *Parser) parseIf() (ast.Expression, error) {
	p.next()

	var (
		expr ast.Test
		err  error
	)
	expr.Cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr.Csq, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if err := p.expectKW(token.KwElse, ""); err == nil {
		p.next()
		switch p.curr.Type {
		case token.Lcurly:
			expr.Alt, err = p.parseBlock()
		case token.Keyword:
			expr.Alt, err = p.parseKeyword()
		default:
		}
	}
	if !p.is(token.EOL) && !p.is(token.EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *Parser) parseWhile() (ast.Expression, error) {
	p.next()

	var (
		expr ast.While
		err  error
	)
	expr.Cdt, err = p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr.Body, err = p.parseBlock()
	if err != nil {
		return nil, err
	}
	if !p.is(token.EOL) && !p.is(token.EOF) {
		return nil, p.parseError("expected newline or ';'")
	}
	return expr, nil
}

func (p *Parser) parseReturn() (ast.Expression, error) {
	p.next()
	if p.is(token.EOL) || p.is(token.EOF) {
		return ast.CreateReturn(nil), nil
	}
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	return ast.CreateReturn(right), nil
}

func (p *Parser) parseBreak() (ast.Expression, error) {
	p.next()
	return ast.Break{}, nil
}

func (p *Parser) parseContinue() (ast.Expression, error) {
	p.next()
	return ast.Continue{}, nil
}

func (p *Parser) parseTernary(left ast.Expression) (ast.Expression, error) {
	var err error
	expr := ast.Test{
		Cdt: left,
	}
	p.next()
	if expr.Csq, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	if err := p.expect(token.Colon, "expected ':'"); err != nil {
		return nil, err
	}
	p.next()

	if expr.Alt, err = p.parse(powLowest); err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parsePath(left ast.Expression) (ast.Expression, error) {
	v, ok := left.(ast.Variable)
	if !ok {
		return nil, p.parseError("unexpected path operator")
	}
	p.next()
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	return ast.CreatePath(v.Ident, right), nil
}

func (p *Parser) parseAssign(left ast.Expression) (ast.Expression, error) {
	switch left.(type) {
	case ast.Variable, ast.Index:
	default:
		return nil, p.parseError("unexpected assignment operator")
	}
	op := p.curr.Type
	p.next()
	right, err := p.parse(powLowest)
	if err != nil {
		return nil, err
	}
	expr := ast.CreateAssign(left, right)
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
		expr.Right = ast.Binary{
			Op:    op,
			Left:  left,
			Right: right,
		}
	}
	return expr, nil
}

func (p *Parser) parseInfix(left ast.Expression) (ast.Expression, error) {
	expr := ast.Binary{
		Op:   p.curr.Type,
		Left: left,
	}
	pow := powers.Get(p.curr.Type)
	p.next()
	right, err := p.parse(pow)
	if err != nil {
		return nil, err
	}
	expr.Right = right
	return expr, nil
}

func (p *Parser) parseIndex(left ast.Expression) (ast.Expression, error) {
	switch left.(type) {
	case ast.Array, ast.Dict, ast.Index, ast.Variable, ast.Literal:
	default:
		return nil, p.parseError("unexpected index operator")
	}
	ix := ast.Index{
		Arr: left,
	}
	p.next()
	for !p.is(token.Rsquare) && !p.done() {
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		ix.List = append(ix.List, expr)
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
	if len(ix.List) == 0 {
		return nil, p.parseError("empty index")
	}
	p.next()
	return ix, nil
}

func (p *Parser) parseCompitem(until rune) ([]ast.CompItem, error) {
	var list []ast.CompItem
	p.next()
	for !p.is(until) && !p.done() {
		var item ast.CompItem
		if err := p.expect(token.Ident, "expected identifier"); err != nil {
			return nil, err
		}
		item.Ident = p.curr.Literal
		p.next()
		if err := p.expectKW(token.KwIn, "expected 'in' keyword"); err != nil {
			return nil, err
		}
		p.next()
		expr, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		item.Iter = expr
		for p.is(token.Keyword) && p.curr.Literal == token.KwIf {
			p.next()
			expr, err := p.parse(powLowest)
			if err != nil {
				return nil, err
			}
			item.Cdt = append(item.Cdt, expr)
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

func (p *Parser) parseListcomp(left ast.Expression) (ast.Expression, error) {
	cmp := ast.ListComp{
		Body: left,
	}
	list, err := p.parseCompitem(token.Rsquare)
	if err == nil {
		cmp.List = list
	}
	return cmp, err
}

func (p *Parser) parseArray() (ast.Expression, error) {
	p.next()
	var arr ast.Array
	for !p.is(token.Rsquare) && !p.done() {
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		if err := p.expectKW(token.KwFor, ""); len(arr.List) == 0 && err == nil {
			return p.parseListcomp(e)
		}
		arr.List = append(arr.List, e)
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

func (p *Parser) parseDictcomp(key, val ast.Expression) (ast.Expression, error) {
	cmp := ast.DictComp{
		Key: key,
		Val: val,
	}
	list, err := p.parseCompitem(token.Rcurly)
	if err == nil {
		cmp.List = list
	}
	return cmp, err
}

func (p *Parser) parseDict() (ast.Expression, error) {
	p.next()
	var d ast.Dict
	d.List = make(map[ast.Expression]ast.Expression)
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
		if err := p.expectKW(token.KwFor, ""); len(d.List) == 0 && err == nil {
			return p.parseDictcomp(k, v)
		}
		d.List[k] = v
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

func (p *Parser) parsePrefix() (ast.Expression, error) {
	var expr ast.Expression
	switch p.curr.Type {
	case token.Sub, token.Not, token.BinNot:
		op := p.curr.Type
		p.next()

		right, err := p.parse(powPrefix)
		if err != nil {
			return nil, err
		}
		expr = ast.Unary{
			Op:    op,
			Right: right,
		}
	case token.Literal:
		expr = ast.CreateLiteral(p.curr.Literal)
		p.next()
	case token.Double:
		n, err := strconv.ParseFloat(p.curr.Literal, 64)
		if err != nil {
			return nil, err
		}
		expr = ast.CreateDouble(n)
		p.next()
	case token.Integer:
		n, err := strconv.ParseInt(p.curr.Literal, 0, 64)
		if err != nil {
			return nil, err
		}
		expr = ast.CreateInteger(n)
		p.next()
	case token.Ident:
		expr = ast.CreateVariable(p.curr.Literal)
		p.next()
	case token.Boolean:
		b, err := strconv.ParseBool(p.curr.Literal)
		if err != nil {
			return nil, err
		}
		expr = ast.CreateBoolean(b)
		p.next()
	default:
		return nil, p.parseError("prefix operator not recognized")
	}
	return expr, nil
}

func (p *Parser) parseCall(left ast.Expression) (ast.Expression, error) {
	v, ok := left.(ast.Variable)
	if !ok {
		return nil, p.parseError("unexpected call operator")
	}
	p.next()
	expr := ast.Call{
		Ident: v.Ident,
	}
	for !p.is(token.Rparen) && !p.done() {
		if p.peekIs(token.Assign) {
			break
		}
		e, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		expr.Args = append(expr.Args, e)
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
		a := ast.CreateParameter(p.curr.Literal)
		p.next()
		if err := p.expect(token.Assign, "expected '='"); err != nil {
			return nil, err
		}
		p.next()
		val, err := p.parse(powLowest)
		if err != nil {
			return nil, err
		}
		a.Expr = val
		expr.Args = append(expr.Args, a)
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

func (p *Parser) parseGroup() (ast.Expression, error) {
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

func (p *Parser) getPrefixExpr() (ast.Expression, error) {
	fn, ok := p.prefix[p.curr.Type]
	if !ok {
		return nil, p.parseError("unary operator not recognized")
	}
	return fn()
}

func (p *Parser) getInfixExpr(left ast.Expression) (ast.Expression, error) {
	fn, ok := p.infix[p.curr.Type]
	if !ok {
		return nil, p.parseError("binary operator not recognized")
	}
	return fn(left)
}

func (p *Parser) peekIs(r rune) bool {
	return p.peek.Type == r
}

func (p *Parser) is(r rune) bool {
	return p.curr.Type == r
}

func (p *Parser) expect(r rune, msg string) error {
	if !p.is(r) {
		return p.parseError(msg)
	}
	return nil
}

func (p *Parser) expectKW(kw, msg string) error {
	if err := p.expect(token.Keyword, msg); err != nil {
		return err
	}
	if p.curr.Literal != kw {
		return p.parseError(msg)
	}
	return nil
}

func (p *Parser) eol() error {
	switch p.curr.Type {
	case token.EOL:
		p.next()
	case token.EOF:
	default:
		return p.parseError("expected newline or ';'")
	}
	return nil
}

func (p *Parser) skip(r rune) {
	for p.is(r) {
		p.next()
	}
}

func (p *Parser) done() bool {
	return p.is(token.EOF)
}

func (p *Parser) next() {
	p.curr = p.peek
	p.peek = p.scan.Scan()
}

func (p *Parser) parseError(message string) error {
	return ParseError{
		Token:   p.curr,
		File:    p.file,
		Line:    p.scan.CurrentLine(p.curr.Position),
		Message: message,
	}
}

package ast

import (
	"fmt"
	"math"
	"strconv"

	"github.com/midbel/buddy/token"
)

type Expression interface {
	IsValue() bool
}

func CreatePrimitive(tok token.Token, res interface{}) (Expression, error) {
	switch r := res.(type) {
	case int64:
		return CreateInteger(tok, r), nil
	case float64:
		return CreateDouble(tok, r), nil
	case bool:
		return CreateBoolean(tok, r), nil
	case string:
		return CreateLiteral(tok, r), nil
	default:
		return nil, fmt.Errorf("unexpected primitive type: %T", res)
	}
}

func IsTrue(e Expression) bool {
	switch e := e.(type) {
	case Literal:
		return e.Str != ""
	case Integer:
		return e.Value != 0
	case Double:
		return e.Value != 0
	case Boolean:
		return e.Value
	case Array:
		return len(e.List) > 0
	case Dict:
		return len(e.List) > 0
	default:
		return false
	}
}

type Assert struct {
	token.Token
	Expr Expression
}

func CreateAssert(tok token.Token, expr Expression) Assert {
	return Assert{
		Token: tok,
		Expr:  expr,
	}
}

func (_ Assert) IsValue() bool {
	return false
}

type Path struct {
	token.Token
	Ident string
	Right Expression
}

func CreatePath(tok token.Token, ident string, right Expression) Path {
	return Path{
		Token: tok,
		Ident: ident,
		Right: right,
	}
}

func (_ Path) IsValue() bool {
	return false
}

type Symbol struct {
	token.Token
	Ident string
	Alias string
}

func CreateSymbol(tok token.Token, ident string) Symbol {
	return Symbol{
		Token: tok,
		Ident: ident,
		Alias: ident,
	}
}

func (_ Symbol) IsValue() bool {
	return false
}

type Import struct {
	token.Token
	Ident   []string
	Alias   string
	Symbols []Symbol
}

func (_ Import) IsValue() bool {
	return false
}

type Variable struct {
	token.Token
	Ident string
}

func CreateVariable(tok token.Token, ident string) Variable {
	return Variable{
		Token: tok,
		Ident: ident,
	}
}

func (_ Variable) IsValue() bool {
	return false
}

type Literal struct {
	token.Token
	Str string
}

func CreateLiteral(tok token.Token, str string) Literal {
	return Literal{
		Token: tok,
		Str:   str,
	}
}

func (_ Literal) IsValue() bool {
	return true
}

func (i Literal) Add(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		i.Str += strconv.FormatInt(y.Value, 10)
		return i
	case Double:
		i.Str += strconv.FormatFloat(y.Value, 'f', -1, 64)
		return i
	default:
		return nil
	}
}

func (i Literal) Eq(other Expression) Expression {
	y, ok := other.(Literal)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Str == y.Str)
}

func (i Literal) Ne(other Expression) Expression {
	y, ok := other.(Literal)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Str != y.Str)
}

func (i Literal) Lt(other Expression) Expression {
	y, ok := other.(Literal)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Str < y.Str)
}

func (i Literal) Le(other Expression) Expression {
	y, ok := other.(Literal)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Str <= y.Str)
}

func (i Literal) Gt(other Expression) Expression {
	y, ok := other.(Literal)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Str > y.Str)
}

func (i Literal) Ge(other Expression) Expression {
	y, ok := other.(Literal)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Str >= y.Str)
}

type Boolean struct {
	token.Token
	Value bool
}

func CreateBoolean(tok token.Token, b bool) Boolean {
	return Boolean{
		Token: tok,
		Value: b,
	}
}

func (_ Boolean) IsValue() bool {
	return true
}

func (b Boolean) Eq(other Expression) Expression {
	y, ok := other.(Boolean)
	if !ok {
		return nil
	}
	return CreateBoolean(b.Token, b.Value == y.Value)
}

func (b Boolean) Ne(other Expression) Expression {
	y, ok := other.(Boolean)
	if !ok {
		return nil
	}
	return CreateBoolean(b.Token, b.Value != y.Value)
}

type Integer struct {
	token.Token
	Value int64
}

func CreateInteger(tok token.Token, n int64) Integer {
	return Integer{
		Token: tok,
		Value: n,
	}
}

func (_ Integer) IsValue() bool {
	return true
}

func (i Integer) Add(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		i.Value += y.Value
		return i
	case Double:
		y.Value += float64(i.Value)
		return y
	case Literal:
		y.Str = strconv.FormatInt(i.Value, 10) + y.Str
		return y
	default:
		return nil
	}
}

func (i Integer) Sub(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		i.Value -= y.Value
		return i
	case Double:
		y.Value = float64(i.Value) - y.Value
		return y
	default:
		return nil
	}
}

func (i Integer) Mul(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		i.Value *= y.Value
		return i
	case Double:
		y.Value = float64(i.Value) * y.Value
		return y
	default:
		return nil
	}
}

func (i Integer) Div(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		if y.Value == 0 {
			return nil
		}
		i.Value /= y.Value
		return i
	case Double:
		if y.Value == 0 {
			return nil
		}
		y.Value = float64(i.Value) / y.Value
		return y
	default:
		return nil
	}
}

func (i Integer) Pow(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		x := math.Pow(float64(i.Value), float64(y.Value))
		i.Value = int64(x)
		return i
	case Double:
		y.Value = math.Pow(float64(i.Value), y.Value)
		return y
	default:
		return nil
	}
}

func (i Integer) Mod(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		if y.Value == 0 {
			return nil
		}
		i.Value %= y.Value
		return i
	case Double:
		if y.Value == 0 {
			return nil
		}
		y.Value = math.Mod(float64(i.Value), y.Value)
		return y
	default:
		return nil
	}
}

func (i Integer) Lshift(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	i.Value = i.Value << y.Value
	return i
}

func (i Integer) Rshift(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	i.Value = i.Value >> y.Value
	return i
}

func (i Integer) And(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	i.Value = i.Value & y.Value
	return i
}

func (i Integer) Or(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	i.Value = i.Value | y.Value
	return i
}

func (i Integer) Eq(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Value == y.Value)
}

func (i Integer) Ne(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Value != y.Value)
}

func (i Integer) Lt(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Value < y.Value)
}

func (i Integer) Le(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Value <= y.Value)
}

func (i Integer) Gt(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Value > y.Value)
}

func (i Integer) Ge(other Expression) Expression {
	y, ok := other.(Integer)
	if !ok {
		return nil
	}
	return CreateBoolean(i.Token, i.Value >= y.Value)
}

type Double struct {
	token.Token
	Value float64
}

func CreateDouble(tok token.Token, f float64) Double {
	return Double{
		Token: tok,
		Value: f,
	}
}

func (_ Double) IsValue() bool {
	return true
}

func (d Double) Add(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		d.Value += float64(y.Value)
		return d
	case Double:
		d.Value += y.Value
		return d
	case Literal:
		y.Str = strconv.FormatFloat(d.Value, 'f', -1, 64) + y.Str
		return y
	default:
		return nil
	}
}

func (d Double) Sub(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		d.Value -= float64(y.Value)
		return d
	case Double:
		d.Value -= y.Value
		return d
	default:
		return nil
	}
}

func (d Double) Mul(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		d.Value *= float64(y.Value)
		return d
	case Double:
		d.Value *= d.Value
		return d
	default:
		return nil
	}
	return d
}

func (d Double) Div(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		if y.Value == 0 {
			return nil
		}
		d.Value /= float64(y.Value)
	case Double:
		if y.Value == 0 {
			return nil
		}
		d.Value /= y.Value
	default:
		return nil
	}
	return d
}

func (d Double) Mod(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		if y.Value == 0 {
			return nil
		}
		d.Value = math.Mod(d.Value, float64(y.Value))
		return d
	case Double:
		if y.Value == 0 {
			return nil
		}
		d.Value = math.Mod(d.Value, y.Value)
		return d
	default:
		return nil
	}
	return d
}

func (d Double) Pow(other Expression) Expression {
	switch y := other.(type) {
	case Integer:
		d.Value = math.Pow(d.Value, float64(y.Value))
		return d
	case Double:
		d.Value = math.Pow(d.Value, y.Value)
		return d
	default:
		return nil
	}
	return d
}

func (d Double) Eq(other Expression) Expression {
	y, ok := other.(Double)
	if !ok {
		return nil
	}
	return CreateBoolean(d.Token, d.Value == y.Value)
}

func (d Double) Ne(other Expression) Expression {
	y, ok := other.(Double)
	if !ok {
		return nil
	}
	return CreateBoolean(d.Token, d.Value != y.Value)
}

func (d Double) Lt(other Expression) Expression {
	y, ok := other.(Double)
	if !ok {
		return nil
	}
	return CreateBoolean(d.Token, d.Value < y.Value)
}

func (d Double) Le(other Expression) Expression {
	y, ok := other.(Double)
	if !ok {
		return nil
	}
	return CreateBoolean(d.Token, d.Value <= y.Value)
}

func (d Double) Gt(other Expression) Expression {
	y, ok := other.(Double)
	if !ok {
		return nil
	}
	return CreateBoolean(d.Token, d.Value > y.Value)
}

func (d Double) Ge(other Expression) Expression {
	y, ok := other.(Double)
	if !ok {
		return nil
	}
	return CreateBoolean(d.Token, d.Value >= y.Value)
}

type Array struct {
	token.Token
	List []Expression
}

func (_ Array) IsValue() bool {
	return false
}

type Dict struct {
	token.Token
	List map[Expression]Expression
}

func (_ Dict) IsValue() bool {
	return false
}

type Slice struct {
	token.Token
	Start Expression
	End   Expression
	Step  Expression
}

func CreateSlice(tok token.Token, start, end Expression) Slice {
	return Slice{
		Token: tok,
		Start: start,
		End:   end,
	}
}

func (_ Slice) IsValue() bool {
	return false
}

type Index struct {
	token.Token
	Arr  Expression
	List []Expression
}

func CreateIndex(tok token.Token, arr Expression) Index {
	return Index{
		Token: tok,
		Arr:   arr,
	}
}

func (_ Index) IsValue() bool {
	return false
}

type Parameter struct {
	token.Token
	Ident string
	Expr  Expression
}

func CreateParameter(tok token.Token, ident string) Parameter {
	return Parameter{
		Token: tok,
		Ident: ident,
	}
}

func (_ Parameter) IsValue() bool {
	return false
}

type Function struct {
	token.Token
	Ident  string
	Params []Expression
	Body   Expression
}

func CreateFunction(tok token.Token, ident string) Function {
	return Function{
		Token: tok,
		Ident: ident,
	}
}

func (_ Function) IsValue() bool {
	return false
}

type Assign struct {
	token.Token
	Ident Expression
	Right Expression
}

func CreateAssign(tok token.Token, ident, expr Expression) Assign {
	return Assign{
		Token: tok,
		Ident: ident,
		Right: expr,
	}
}

func (_ Assign) IsValue() bool {
	return false
}

type Call struct {
	token.Token
	Ident string
	Args  []Expression
}

func (_ Call) IsValue() bool {
	return false
}

type Return struct {
	token.Token
	Right Expression
}

func CreateReturn(tok token.Token, right Expression) Return {
	return Return{
		Token: tok,
		Right: right,
	}
}

func (_ Return) IsValue() bool {
	return false
}

type Unary struct {
	token.Token
	Op    rune
	Right Expression
}

func (_ Unary) IsValue() bool {
	return false
}

type Binary struct {
	token.Token
	Op    rune
	Left  Expression
	Right Expression
}

func (_ Binary) IsValue() bool {
	return false
}

type Script struct {
	token.Token
	List    []Expression
	Symbols map[string]Expression
}

func CreateScript(tok token.Token) Script {
	return Script{
		Token:   tok,
		Symbols: make(map[string]Expression),
	}
}

func CreateScriptFromList(tok token.Token, list []Expression) Script {
	return Script{
		Token: tok,
		List:  list,
	}
}

func (_ Script) IsValue() bool {
	return false
}

type CompItem struct {
	token.Token
	Ident string
	Iter  Expression
	Cdt   []Expression
}

func (_ CompItem) IsValue() bool {
	return false
}

type DictComp struct {
	token.Token
	Key  Expression
	Val  Expression
	List []CompItem
}

func (_ DictComp) IsValue() bool {
	return false
}

type ListComp struct {
	token.Token
	Body Expression
	List []CompItem
}

func (c ListComp) IsValue() bool {
	return false
}

type For struct {
	token.Token
	Init Expression
	Incr Expression
	While
}

func (f For) IsValue() bool {
	return false
}

type ForEach struct {
	token.Token
	Ident string
	Iter  Expression
	Body  Expression
}

func (_ ForEach) IsValue() bool {
	return false
}

type While struct {
	token.Token
	Cdt  Expression
	Body Expression
}

func (_ While) IsValue() bool {
	return false
}

type Break struct {
	token.Token
}

func (_ Break) IsValue() bool {
	return false
}

type Continue struct {
	token.Token
}

func (_ Continue) IsValue() bool {
	return false
}

type Test struct {
	token.Token
	Cdt Expression
	Csq Expression
	Alt Expression
}

func (_ Test) IsValue() bool {
	return false
}

package ast

import (
	"fmt"
)

type Expression interface {
	isPrimitive() bool
}

func CreatePrimitive(res interface{}) (Expression, error) {
	switch r := res.(type) {
	case int64:
		return CreateInteger(r), nil
	case float64:
		return CreateDouble(r), nil
	case bool:
		return CreateBoolean(r), nil
	case string:
		return CreateLiteral(r), nil
	default:
		return nil, fmt.Errorf("unexpected primitive type: %T", res)
	}
}

type Assert struct {
	Expr Expression
}

func (_ Assert) isPrimitive() bool {
	return false
}

type Link struct {
	Callable
}

func CreateLink(call Callable) Expression {
	return Link{
		Callable: call,
	}
}

func (_ Link) isPrimitive() bool {
	return false
}

type Path struct {
	Ident string
	Right Expression
}

func CreatePath(ident string, right Expression) Path {
	return Path{
		Ident: ident,
		Right: right,
	}
}

func (_ Path) isPrimitive() bool {
	return false
}

type Symbol struct {
	Ident string
	Alias string
}

func CreateSymbol(ident string) symbol {
	return symbol{
		Ident: ident,
		Alias: ident,
	}
}

func (_ Symbol) isPrimitive() bool {
	return false
}

type Module struct {
	Ident   []string
	Alias   string
	Symbols []Symbol
}

func (_ Module) isPrimitive() bool {
	return false
}

type Variable struct {
	Ident string
}

func CreateVariable(ident string) variable {
	return Variable{
		Ident: ident,
	}
}

func (_ Variable) isPrimitive() bool {
	return false
}

type Literal struct {
	Str string
}

func CreateLiteral(str string) literal {
	return Literal{
		Str: str,
	}
}

func (_ Literal) isPrimitive() bool {
	return true
}

type Boolean struct {
	Value bool
}

func CreateBoolean(b bool) boolean {
	return Boolean{
		Value: b,
	}
}

func (_ Boolean) isPrimitive() bool {
	return true
}

type Integer struct {
	Value int64
}

func CreateInteger(n int64) integer {
	return Integer{
		Value: n,
	}
}

func (_ Integer) isPrimitive() bool {
	return true
}

type Double struct {
	Value float64
}

func CreateDouble(f float64) double {
	return Double{
		Value: f,
	}
}

func (_ Double) isPrimitive() bool {
	return true
}

type Array struct {
	List []Expression
}

func (_ Array) isPrimitive() bool {
	return false
}

type Dict struct {
	List map[Expression]Expression
}

func (_ Dict) isPrimitive() bool {
	return false
}

type Slice struct {
	Start Expression
	End   Expression
}

func CreateSlice(start, end Expression) Slice {
	return Slice{
		Start: start,
		End:   end,
	}
}

func (_ Slice) isPrimitive() bool {
	return false
}

type Index struct {
	Arr  Expression
	List []Expression
}

func (_ Index) isPrimitive() bool {
	return false
}

type Parameter struct {
	Ident string
	Expr  Expression
}

func CreateParameter(ident string) parameter {
	return parameter{
		Ident: ident,
	}
}

func (_ Parameter) isPrimitive() bool {
	return false
}

type Function struct {
	Ident  string
	Params []Expression
	Body   Expression
}

func (_ Function) isPrimitive() bool {
	return false
}

type Walrus struct {
	assign
}

func (w Walrus) isPrimitive() bool {
	return false
}

type Assign struct {
	Ident Expression
	Right Expression
}

func CreateAssign(ident, expr Expression) assign {
	return Assign{
		Ident: ident,
		Right: expr,
	}
}

func (_ Assign) isPrimitive() bool {
	return false
}

type Call struct {
	Ident string
	Args  []Expression
}

func (_ Call) isPrimitive() bool {
	return false
}

type Return struct {
	Right Expression
}

func (_ Returned) isPrimitive() bool {
	return false
}

type Unary struct {
	Op    rune
	Right Expression
}

func (_ Unary) isPrimitive() bool {
	return false
}

type Binary struct {
	Op    rune
	Left  Expression
	Right Expression
}

func (_ Binary) isPrimitive() bool {
	return false
}

type Script struct {
	List    []Expression
	Symbols map[string]Expression
}

func (_ Script) isPrimitive() bool {
	return false
}

type CompItem struct {
	Ident string
	Iter  Expression
	Cdt   []Expression
}

func (_ CompItem) isPrimitive() bool {
	return false
}

type DictComp struct {
	Key  Expression
	Val  Expression
	List []CompItem
}

func (_ DictComp) isPrimitive() bool {
	return false
}

type ListComp struct {
	Body Expression
	List []CompItem
}

func (c ListComp) isPrimitive() bool {
	return false
}

type For struct {
	Init Expression
	Incr Expression
	While
}

func (f For) isPrimitive() bool {
	return false
}

type ForEach struct {
	Ident string
	Iter  Expression
	Body  Expression
}

func (_ ForEach) isPrimitive() bool {
	return false
}

type While struct {
	Cdt  Expression
	Body Expression
}

func (_ While) isPrimitive() bool {
	return false
}

type Break struct{}

func (_ Break) isPrimitive() bool {
	return false
}

type Continue struct{}

func (_ Continue) isPrimitive() bool {
	return false
}

type Test struct {
	Cdt Expression
	Csq Expression
	Alt Expression
}

func (_ Test) isPrimitive() bool {
	return false
}

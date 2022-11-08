package ast

import (
	"fmt"
)

type Expression interface {
	IsValue() bool
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

func CreateAssert(expr Expression) Assert {
	return Assert{
		Expr: expr,
	}
}

func (_ Assert) IsValue() bool {
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

func (_ Path) IsValue() bool {
	return false
}

type Symbol struct {
	Ident string
	Alias string
}

func CreateSymbol(ident string) Symbol {
	return Symbol{
		Ident: ident,
		Alias: ident,
	}
}

func (_ Symbol) IsValue() bool {
	return false
}

type Import struct {
	Ident   []string
	Alias   string
	Symbols []Symbol
}

func (_ Import) IsValue() bool {
	return false
}

type Variable struct {
	Ident string
}

func CreateVariable(ident string) Variable {
	return Variable{
		Ident: ident,
	}
}

func (_ Variable) IsValue() bool {
	return false
}

type Literal struct {
	Str string
}

func CreateLiteral(str string) Literal {
	return Literal{
		Str: str,
	}
}

func (_ Literal) IsValue() bool {
	return true
}

type Boolean struct {
	Value bool
}

func CreateBoolean(b bool) Boolean {
	return Boolean{
		Value: b,
	}
}

func (_ Boolean) IsValue() bool {
	return true
}

type Integer struct {
	Value int64
}

func CreateInteger(n int64) Integer {
	return Integer{
		Value: n,
	}
}

func (_ Integer) IsValue() bool {
	return true
}

type Double struct {
	Value float64
}

func CreateDouble(f float64) Double {
	return Double{
		Value: f,
	}
}

func (_ Double) IsValue() bool {
	return true
}

type Array struct {
	List []Expression
}

func (_ Array) IsValue() bool {
	return false
}

type Dict struct {
	List map[Expression]Expression
}

func (_ Dict) IsValue() bool {
	return false
}

type Slice struct {
	Start Expression
	End   Expression
	Step  Expression
}

func CreateSlice(start, end Expression) Slice {
	return Slice{
		Start: start,
		End:   end,
	}
}

func (_ Slice) IsValue() bool {
	return false
}

type Index struct {
	Arr  Expression
	List []Expression
}

func CreateIndex(arr Expression) Index {
	return Index{
		Arr: arr,
	}
}

func (_ Index) IsValue() bool {
	return false
}

type Parameter struct {
	Ident string
	Expr  Expression
}

func CreateParameter(ident string) Parameter {
	return Parameter{
		Ident: ident,
	}
}

func (_ Parameter) IsValue() bool {
	return false
}

type Function struct {
	Ident  string
	Params []Expression
	Body   Expression
}

func CreateFunction(ident string) Function {
	return Function{
		Ident: ident,
	}
}

func (_ Function) IsValue() bool {
	return false
}

type Assign struct {
	Ident Expression
	Right Expression
}

func CreateAssign(ident, expr Expression) Assign {
	return Assign{
		Ident: ident,
		Right: expr,
	}
}

func (_ Assign) IsValue() bool {
	return false
}

type Call struct {
	Ident string
	Args  []Expression
}

func (_ Call) IsValue() bool {
	return false
}

type Return struct {
	Right Expression
}

func CreateReturn(right Expression) Return {
	return Return{
		Right: right,
	}
}

func (_ Return) IsValue() bool {
	return false
}

type Unary struct {
	Op    rune
	Right Expression
}

func (_ Unary) IsValue() bool {
	return false
}

type Binary struct {
	Op    rune
	Left  Expression
	Right Expression
}

func (_ Binary) IsValue() bool {
	return false
}

type Script struct {
	List    []Expression
	Symbols map[string]Expression
}

func CreateScript() Script {
	return Script{
		Symbols: make(map[string]Expression),
	}
}

func CreateScriptFromList(list []Expression) Script {
	return Script{
		List: list,
	}
}

func (_ Script) IsValue() bool {
	return false
}

type CompItem struct {
	Ident string
	Iter  Expression
	Cdt   []Expression
}

func (_ CompItem) IsValue() bool {
	return false
}

type DictComp struct {
	Key  Expression
	Val  Expression
	List []CompItem
}

func (_ DictComp) IsValue() bool {
	return false
}

type ListComp struct {
	Body Expression
	List []CompItem
}

func (c ListComp) IsValue() bool {
	return false
}

type For struct {
	Init Expression
	Incr Expression
	While
}

func (f For) IsValue() bool {
	return false
}

type ForEach struct {
	Ident string
	Iter  Expression
	Body  Expression
}

func (_ ForEach) IsValue() bool {
	return false
}

type While struct {
	Cdt  Expression
	Body Expression
}

func (_ While) IsValue() bool {
	return false
}

type Break struct{}

func (_ Break) IsValue() bool {
	return false
}

type Continue struct{}

func (_ Continue) IsValue() bool {
	return false
}

type Test struct {
	Cdt Expression
	Csq Expression
	Alt Expression
}

func (_ Test) IsValue() bool {
	return false
}

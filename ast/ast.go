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

func CreateAssert(expr Expression) Assert {
	return Assert{
		Expr: expr,
	}
}

func (_ Assert) isPrimitive() bool {
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

func CreateSymbol(ident string) Symbol {
	return Symbol{
		Ident: ident,
		Alias: ident,
	}
}

func (_ Symbol) isPrimitive() bool {
	return false
}

type Import struct {
	Ident   []string
	Alias   string
	Symbols []Symbol
}

func (_ Import) isPrimitive() bool {
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

func (_ Variable) isPrimitive() bool {
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

func (_ Literal) isPrimitive() bool {
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

func (_ Boolean) isPrimitive() bool {
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

func (_ Integer) isPrimitive() bool {
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

func CreateParameter(ident string) Parameter {
	return Parameter{
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

func CreateFunction(ident string) Function {
	return Function{
		Ident: ident,
	}
}

func (_ Function) isPrimitive() bool {
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

func CreateReturn(right Expression) Return {
	return Return{
		Right: right,
	}
}

func (_ Return) isPrimitive() bool {
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

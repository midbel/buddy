package buddy

import (
	"fmt"
)

type Expression interface {
	isPrimitive() bool
}

func createPrimitive(res interface{}) (Expression, error) {
	switch r := res.(type) {
	case int64:
		return createNumber(float64(r)), nil
	case float64:
		return createNumber(r), nil
	case bool:
		return createBoolean(r), nil
	case string:
		return createLiteral(r), nil
	default:
		return nil, fmt.Errorf("unexpected primitive type: %T", res)
	}
}

type path struct {
	ident string
	right Expression
}

func (_ path) isPrimitive() bool {
	return false
}

type symbol struct {
	ident string
	alias string
}

func (_ symbol) isPrimitive() bool {
	return false
}

type module struct {
	ident   []string
	alias   string
	symbols []symbol
}

func (_ module) isPrimitive() bool {
	return false
}

type variable struct {
	ident string
}

func createVariable(ident string) variable {
	return variable{
		ident: ident,
	}
}

func (_ variable) isPrimitive() bool {
	return false
}

type literal struct {
	str string
}

func createLiteral(str string) literal {
	return literal{
		str: str,
	}
}

func (_ literal) isPrimitive() bool {
	return true
}

type boolean struct {
	value bool
}

func createBoolean(b bool) boolean {
	return boolean{
		value: b,
	}
}

func (_ boolean) isPrimitive() bool {
	return true
}

type number struct {
	value float64
}

func createNumber(f float64) number {
	return number{
		value: f,
	}
}

func (_ number) isPrimitive() bool {
	return true
}

type array struct {
	list []Expression
}

func (_ array) isPrimitive() bool {
	return false
}

type dict struct {
	list map[Expression]Expression
}

func (_ dict) isPrimitive() bool {
	return false
}

type slice struct {
	start Expression
	end   Expression
}

func (_ slice) isPrimitive() bool {
	return false
}

type index struct {
	arr  Expression
	list []Expression
}

func (_ index) isPrimitive() bool {
	return false
}

type parameter struct {
	ident string
	expr  Expression
}

func createParameter(ident string) parameter {
	return parameter{
		ident: ident,
	}
}

func (_ parameter) isPrimitive() bool {
	return false
}

type function struct {
	ident  string
	params []Expression
	body   Expression
}

func (_ function) isPrimitive() bool {
	return false
}

type assign struct {
	ident Expression
	right Expression
}

func (_ assign) isPrimitive() bool {
	return false
}

type call struct {
	ident string
	args  []Expression
}

func (_ call) isPrimitive() bool {
	return false
}

type returned struct {
	right Expression
}

func (_ returned) isPrimitive() bool {
	return false
}

type unary struct {
	op    rune
	right Expression
}

func (_ unary) isPrimitive() bool {
	return false
}

type binary struct {
	op    rune
	left  Expression
	right Expression
}

func (_ binary) isPrimitive() bool {
	return false
}

type script struct {
	list    []Expression
	symbols map[string]Expression
}

func (_ script) isPrimitive() bool {
	return false
}

type foreach struct {
	ident string
	iter  Expression
	body  Expression
}

func (_ foreach) isPrimitive() bool {
	return false
}

type while struct {
	cdt  Expression
	body Expression
}

func (_ while) isPrimitive() bool {
	return false
}

type breaked struct{}

func (_ breaked) isPrimitive() bool {
	return false
}

type continued struct{}

func (_ continued) isPrimitive() bool {
	return false
}

type test struct {
	cdt Expression
	csq Expression
	alt Expression
}

func (_ test) isPrimitive() bool {
	return false
}

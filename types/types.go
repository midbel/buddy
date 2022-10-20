package types

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var (
	ErrIncompatible = errors.New("incompatible type")
	ErrOperation    = errors.New("unsupported operation")
	ErrZero         = errors.New("division by zero")
)

type Sizeable interface {
	Len() int
}

type Primitive interface {
	fmt.Stringer

	Raw() any

	Rev() (Primitive, error)
	Not() (Primitive, error)

	Add(Primitive) (Primitive, error)
	Sub(Primitive) (Primitive, error)
	Div(Primitive) (Primitive, error)
	Mod(Primitive) (Primitive, error)
	Mul(Primitive) (Primitive, error)
	Pow(Primitive) (Primitive, error)

	Eq(Primitive) (Primitive, error)
	Ne(Primitive) (Primitive, error)
	Lt(Primitive) (Primitive, error)
	Le(Primitive) (Primitive, error)
	Gt(Primitive) (Primitive, error)
	Ge(Primitive) (Primitive, error)

	True() bool
}

func And(left, right Primitive) (Primitive, error) {
	b := left.True() && right.True()
	return Bool{value: b}, nil
}

func Or(left, right Primitive) (Primitive, error) {
	b := left.True() || right.True()
	return Bool{value: b}, nil
}

type String struct {
	str string
}

func CreateString(str string) Primitive {
	return String{
		str: str,
	}
}

func (s String) Len() int {
	return len(s.str)
}

func (s String) Raw() any {
	return s.str
}

func (s String) String() string {
	return s.str
}

func (_ String) Rev() (Primitive, error) {
	return nil, ErrOperation
}

func (_ String) Not() (Primitive, error) {
	return nil, ErrOperation
}

func (s String) Add(other Primitive) (Primitive, error) {
	var str string
	switch x := other.(type) {
	case Int:
		str = x.String()
	case Float:
		str = x.String()
	case String:
		str = x.String()
	default:
		return nil, ErrOperation
	}
	s.str += str
	return s, nil
}

func (s String) Sub(other Primitive) (Primitive, error) {
	var part int
	switch x := other.(type) {
	case Int:
		part = int(x.value)
	case Float:
		part = int(x.value)
	default:
		return nil, ErrOperation
	}
	if part > len(s.str) {
		s.str = ""
		return s, nil
	}
	if part < 0 {
		s.str = s.str[-part:]
	} else {
		s.str = s.str[:part]
	}
	return s, nil
}

func (s String) Div(other Primitive) (Primitive, error) {
	var part int
	switch x := other.(type) {
	case Int:
		part = int(x.value)
	case Float:
		part = int(x.value)
	default:
		return nil, ErrOperation
	}
	if part == 0 {
		return s, nil
	}
	offset := len(s.str) / part
	s.str = s.str[:offset]
	return s, nil
}

func (s String) Mul(other Primitive) (Primitive, error) {
	var count int
	switch x := other.(type) {
	case Int:
		count = int(x.value)
	case Float:
		count = int(x.value)
	default:
		return nil, ErrOperation
	}
	s.str = strings.Repeat(s.str, count)
	return s, nil
}

func (_ String) Mod(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (_ String) Pow(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (s String) True() bool {
	return s.str != ""
}

func (s String) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(s.str == x.str), nil
}

func (s String) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(s.str != x.str), nil
}

func (s String) Lt(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(s.str < x.str), nil
}

func (s String) Le(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(s.str <= x.str), nil
}

func (s String) Gt(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(s.str > x.str), nil
}

func (s String) Ge(other Primitive) (Primitive, error) {
	x, ok := other.(String)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(s.str >= x.str), nil
}

type Bool struct {
	value bool
}

func CreateBool(b bool) Primitive {
	return Bool{
		value: b,
	}
}

func (b Bool) Raw() any {
	return b.value
}

func (_ Bool) Rev() (Primitive, error) {
	return nil, ErrOperation
}

func (b Bool) Not() (Primitive, error) {
	b.value = !b.value
	return b, nil
}

func (b Bool) String() string {
	return strconv.FormatBool(b.value)
}

func (_ Bool) Add(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (_ Bool) Sub(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (_ Bool) Div(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (_ Bool) Mul(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (_ Bool) Mod(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (_ Bool) Pow(_ Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (b Bool) True() bool {
	return b.value
}

func (b Bool) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(Bool)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(b.value == x.value), nil
}

func (b Bool) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(Bool)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(b.value != x.value), nil
}

func (b Bool) Lt(other Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (b Bool) Le(other Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (b Bool) Gt(other Primitive) (Primitive, error) {
	return nil, ErrOperation
}

func (b Bool) Ge(other Primitive) (Primitive, error) {
	return nil, ErrOperation
}

type Float struct {
	value float64
}

func CreateFloat(f float64) Primitive {
	return Float{
		value: f,
	}
}

func (f Float) Raw() any {
	return f.value
}

func (f Float) Rev() (Primitive, error) {
	f.value = -f.value
	return f, nil
}

func (_ Float) Not() (Primitive, error) {
	return nil, ErrOperation
}

func (f Float) String() string {
	return strconv.FormatFloat(f.value, 'g', -1, 64)
}

func (f Float) Add(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value += float64(x.value)
	case Float:
		f.value += x.value
	case String:
		s := f.String() + x.String()
		return String{str: s}, nil
	default:
		return nil, ErrOperation
	}
	return f, nil
}

func (f Float) Sub(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value -= float64(x.value)
	case Float:
		f.value -= x.value
	default:
		return nil, ErrOperation
	}
	return f, nil
}

func (f Float) Div(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value /= float64(x.value)
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value /= x.value
	default:
		return nil, ErrOperation
	}
	return f, nil
}

func (f Float) Mul(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value *= float64(x.value)
	case Float:
		f.value *= x.value
	default:
		return nil, ErrOperation
	}
	return f, nil
}

func (f Float) Mod(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value = math.Mod(f.value, float64(x.value))
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f.value = math.Mod(f.value, x.value)
	default:
		return nil, ErrOperation
	}
	return f, nil
}

func (f Float) Pow(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		f.value = math.Pow(f.value, float64(x.value))
	case Float:
		f.value = math.Pow(f.value, x.value)
	default:
		return nil, ErrOperation
	}
	return f, nil
}

func (f Float) True() bool {
	return f.value != 0
}

func (f Float) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(f.value == x.value), nil
}

func (f Float) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(f.value != x.value), nil
}

func (f Float) Lt(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(f.value < x.value), nil
}

func (f Float) Le(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(f.value <= x.value), nil
}

func (f Float) Gt(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(f.value > x.value), nil
}

func (f Float) Ge(other Primitive) (Primitive, error) {
	x, ok := other.(Float)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(f.value >= x.value), nil
}

type Int struct {
	value int64
}

func CreateInt(i int64) Primitive {
	return Int{
		value: i,
	}
}

func (i Int) Raw() any {
	return i.value
}

func (i Int) Rev() (Primitive, error) {
	i.value = -i.value
	return i, nil
}

func (_ Int) Not() (Primitive, error) {
	return nil, ErrOperation
}

func (i Int) String() string {
	return strconv.FormatInt(i.value, 10)
}

func (i Int) Add(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		i.value += x.value
	case Float:
		f := float64(i.value) + x.value
		return Float{value: f}, nil
	case String:
		s := i.String() + x.String()
		return String{str: s}, nil
	default:
		return nil, ErrOperation
	}
	return i, nil
}

func (i Int) Sub(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		i.value -= x.value
	case Float:
		f := float64(i.value) - x.value
		return Float{value: f}, nil
	default:
		return nil, ErrOperation
	}
	return i, nil
}

func (i Int) Div(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		i.value /= x.value
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f := float64(i.value) / x.value
		return Float{value: f}, nil
	default:
		return nil, ErrOperation
	}
	return i, nil
}

func (i Int) Mul(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		i.value *= x.value
	case Float:
		f := float64(i.value) * x.value
		return Float{value: f}, nil
	default:
		return nil, ErrOperation
	}
	return i, nil
}

func (i Int) Mod(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		if x.value == 0 {
			return nil, ErrZero
		}
		i.value %= x.value
	case Float:
		if x.value == 0 {
			return nil, ErrZero
		}
		f := math.Mod(float64(i.value), x.value)
		return Float{value: f}, nil
	default:
		return nil, ErrOperation
	}
	return i, nil
}

func (i Int) Pow(other Primitive) (Primitive, error) {
	switch x := other.(type) {
	case Int:
		v := math.Pow(float64(i.value), float64(x.value))
		i.value = int64(v)
	case Float:
		f := math.Pow(float64(i.value), x.value)
		return Float{value: f}, nil
	default:
		return nil, ErrOperation
	}
	return i, nil
}

func (i Int) True() bool {
	return i.value != 0
}

func (i Int) Eq(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(i.value == x.value), nil
}

func (i Int) Ne(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(i.value != x.value), nil
}

func (i Int) Lt(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(i.value < x.value), nil
}

func (i Int) Le(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(i.value <= x.value), nil
}

func (i Int) Gt(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(i.value > x.value), nil
}

func (i Int) Ge(other Primitive) (Primitive, error) {
	x, ok := other.(Int)
	if !ok {
		return nil, ErrIncompatible
	}
	return CreateBool(i.value >= x.value), nil
}

type Array struct {
	values []Primitive
}

func (a Array) Len() int {
	return len(a.values)
}

type Dict struct {
	values map[any]Primitive
}

func (d Dict) Len() int {
	return len(d.values)
}

package parse

import (
	"github.com/midbel/buddy/token"
)

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

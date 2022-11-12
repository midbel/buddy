package visitors

import (
	"fmt"

	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
	"github.com/midbel/buddy/token"
	"github.com/midbel/slices"
)

type loopVisitor struct {
	stack *slices.Stack[bool]
	list  faults.ErrorList
	limit int
}

func Loop() Visitor {
	return &loopVisitor{
		list:  make(faults.ErrorList, 0, faults.MaxErrorCount),
		limit: faults.MaxErrorCount,
		stack: slices.New[bool](),
	}
}

func (v *loopVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	err := v.visit(expr)
	if err == nil && v.list.Size() > 0 {
		err = &v.list
	}
	return expr, err
}

func (v *loopVisitor) visit(expr ast.Expression) error {
	switch e := expr.(type) {
	case ast.Literal:
	case ast.Double:
	case ast.Integer:
	case ast.Boolean:
	case ast.Variable:
	case ast.Array:
		for i := range e.List {
			v.reject(e.List[i])
		}
	case ast.Dict:
		for i := range e.List {
			v.reject(e.List[i])
		}
	case ast.Index:
		v.reject(e.Arr)
		for i := range e.List {
			v.reject(e.List[i])
		}
	case ast.Slice:
		v.reject(e.Start)
		v.reject(e.End)
		v.reject(e.Step)
	case ast.Path:
		v.reject(e.Right)
	case ast.Call:
		for i := range e.Args {
			v.reject(e.Args[i])
		}
	case ast.Parameter:
		v.reject(e.Expr)
	case ast.Assert:
		v.reject(e.Expr)
	case ast.Assign:
		v.reject(e.Ident)
		v.reject(e.Right)
	case ast.Unary:
		v.reject(e.Right)
	case ast.Binary:
		v.reject(e.Left)
		v.reject(e.Right)
	case ast.ListComp:
	case ast.DictComp:
	case ast.Test:
		v.reject(e.Cdt)
		if err := v.visit(e.Csq); err != nil {
			v.list.Append(err)
		}
		if err := v.visit(e.Alt); err != nil {
			v.list.Append(err)
		}
	case ast.While:
		v.reject(e.Cdt)
		v.accept(e.Body)
	case ast.For:
		v.reject(e.Init)
		v.reject(e.Cdt)
		v.reject(e.Incr)
		v.accept(e.Body)
	case ast.ForEach:
		v.reject(e.Iter)
		v.accept(e.Body)
	case ast.Import:
	case ast.Script:
		for i := range e.List {
			if err := v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Function:
		for i := range e.Params {
			v.reject(e.Params[i])
		}
		v.reject(e.Body)
	case ast.Return:
		v.reject(e.Right)
	case ast.Break:
		if !v.inLoop() {
			return notInLoop(e.Literal, e.Position)
		}
	case ast.Continue:
		if !v.inLoop() {
			return notInLoop(e.Literal, e.Position)
		}
	default:
	}
	if !v.noLimit() && v.list.Size() > v.limit {
		return &v.list
	}
	return nil
}

func (v *loopVisitor) noLimit() bool {
	return v.limit <= 0
}

func (v *loopVisitor) accept(expr ast.Expression) {
	v.stack.Push(true)
	defer v.stack.Pop()
	if err := v.visit(expr); err != nil {
		v.list.Append(err)
	}
}

func (v *loopVisitor) reject(expr ast.Expression) {
	v.stack.Push(false)
	defer v.stack.Pop()
	if err := v.visit(expr); err != nil {
		v.list.Append(err)
	}
}

func (v *loopVisitor) pop() {
	v.stack.Pop()
}

func (v *loopVisitor) inLoop() bool {
	return v.stack.Top()
}

func notInLoop(kw string, pos token.Position) error {
	return fmt.Errorf("[%s] %s is not in a loop", pos, kw)
}

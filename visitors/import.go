package visitors

import (
	"github.com/midbel/buddy/ast"
	"github.com/midbel/buddy/faults"
	"github.com/midbel/buddy/token"
)

type importVisitor struct {
	env     *Counter[string]
	list    faults.ErrorList
	modules map[string]token.Token
	limit   int
}

func Import() Visitor {
	return &importVisitor{
		env:     EmptyCounter[string](),
		list:    make(faults.ErrorList, 0, faults.MaxErrorCount),
		modules: make(map[string]token.Token),
		limit:   faults.MaxErrorCount,
	}
}

func (v *importVisitor) Visit(expr ast.Expression) (ast.Expression, error) {
	err := v.visit(expr)
	v.unused()
	if err == nil && v.list.Size() > 0 {
		err = &v.list
	}
	return expr, err
}

func (v *importVisitor) visit(expr ast.Expression) error {
	var err error
	switch e := expr.(type) {
	case ast.Literal:
		// PASS: to be removed later
	case ast.Double:
		// PASS: to be removed later
	case ast.Integer:
		// PASS: to be removed later
	case ast.Boolean:
		// PASS: to be removed later
	case ast.Variable:
		// PASS: to be removed later
	case ast.Array:
		for i := range e.List {
			if err = v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Dict:
		for i := range e.List {
			if err = v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Index:
		if err = v.visit(e.Arr); err != nil {
			v.list.Append(err)
		}
		for i := range e.List {
			if err = v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Path:
		if err := v.exists(e); err != nil {
			v.list.Append(err)
		} else {
			v.env.Incr(e.Ident)
		}
		if err := v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.Slice:
		if err = v.visit(e.Start); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.End); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Step); err != nil {
			v.list.Append(err)
		}
	case ast.Call:
		for i := range e.Args {
			if err = v.visit(e.Args[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Parameter:
		if err = v.visit(e.Expr); err != nil {
			v.list.Append(err)
		}
	case ast.Assert:
		if err = v.visit(e.Expr); err != nil {
			v.list.Append(err)
		}
	case ast.Assign:
		if err := v.visit(e.Ident); err != nil {
			v.list.Append(err)
		}
		if err := v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.Unary:
		err = v.visit(e.Right)
		if err != nil {
			v.list.Append(err)
		}
	case ast.Binary:
		if err = v.visit(e.Left); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.ListComp:
		for i := range e.List {
			if err = v.visit(e.List[i].Iter); err != nil {
				v.list.Append(err)
			}
			for j := range e.List[i].Cdt {
				if err = v.visit(e.List[i].Cdt[j]); err != nil {
					v.list.Append(err)
				}
			}
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.DictComp:
		for i := range e.List {
			if err = v.visit(e.List[i].Iter); err != nil {
				v.list.Append(err)
			}
			for j := range e.List[i].Cdt {
				if err = v.visit(e.List[i].Cdt[j]); err != nil {
					v.list.Append(err)
				}
			}
		}
		if err = v.visit(e.Key); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Val); err != nil {
			v.list.Append(err)
		}
	case ast.Test:
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Csq); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Alt); err != nil {
			v.list.Append(err)
		}
	case ast.While:
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.For:
		if err = v.visit(e.Init); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Cdt); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Incr); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.ForEach:
		if err = v.visit(e.Iter); err != nil {
			v.list.Append(err)
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.Import:
		v.modules[e.Alias] = e.Token
		v.env.Incr(e.Alias)
	case ast.Function:
		for i := range e.Params {
			if err := v.visit(e.Params[i]); err != nil {
				v.list.Append(err)
			}
		}
		if err = v.visit(e.Body); err != nil {
			v.list.Append(err)
		}
	case ast.Script:
		for i := range e.List {
			if err := v.visit(e.List[i]); err != nil {
				v.list.Append(err)
			}
		}
	case ast.Return:
		if err := v.visit(e.Right); err != nil {
			v.list.Append(err)
		}
	case ast.Break:
		// PASS: to be removed later
	case ast.Continue:
		// PASS: to be removed later
	default:
	}
	if !v.noLimit() && v.list.Size() > v.limit {
		return &v.list
	}
	return nil
}

func (v importVisitor) exists(e ast.Path) error {
	ok := v.env.Exists(e.Ident)
	if !ok {
		return undefinedImport(e.Ident, e.Position)
	}
	return nil
}

func (v *importVisitor) unused() {
	for _, i := range v.env.Zero() {
		tok, ok := v.modules[i]
		if !ok {
			continue
		}
		v.list.Append(unusedImport(i, tok.Position))
	}
}

func (v importVisitor) noLimit() bool {
	return v.limit <= 0
}

func unusedImport(ident string, pos token.Position) error {
	return IdentError{
		Position: pos,
		Ident:    ident,
		What:     "module imported but not used",
	}
}

func undefinedImport(ident string, pos token.Position) error {
	return IdentError{
		Position: pos,
		Ident:    ident,
		What:     "module used but not imported",
	}
}

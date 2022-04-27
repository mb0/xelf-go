package lib

import (
	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

var Mut = &mutSpec{impl("<form mut any data? tupl?|tag _>")}

// TODO add a custom resl method to mutSpec so we can detect incompatible types during resolution

type mutSpec struct{ exp.SpecBase }

func (s *mutSpec) Value() lit.Val { return s }
func (s *mutSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	fst, err := p.Eval(c.Env, c.Args[0])
	if err != nil {
		return nil, err
	}
	// TODO make sure fst.Val is of fst.Res and mutable otherwise set the type
	mut, ok := fst.Val.(lit.Mut)
	if !ok {
		return nil, ast.ErrEval(fst.Src, "not a mutable value", nil)
	}
	assign := c.Args[1]
	if assign != nil {
		a, err := p.Eval(c.Env, assign)
		if err != nil {
			return nil, err
		}
		err = mut.Assign(a.Val)
		if err != nil {
			return nil, err
		}
	}
	edit, _ := c.Args[2].(*exp.Tupl)
	delta := make(lit.Delta, 0, len(edit.Els))
	for _, el := range edit.Els {
		tag := el.(*exp.Tag)
		ta, err := p.Eval(c.Env, tag.Exp)
		if err != nil {
			return nil, err
		}
		delta = append(delta, lit.KeyVal{Key: tag.Tag, Val: ta.Val})
	}
	return fst, lit.Apply(p.Reg, mut, delta)
}

var Append = &appendSpec{impl("<form append list tupl? _>")}

type appendSpec struct{ exp.SpecBase }

func (s *appendSpec) Value() lit.Val { return s }
func (s *appendSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	fst, err := p.Eval(c.Env, c.Args[0])
	if err != nil {
		return nil, err
	}
	mut, ok := fst.Val.(lit.Apdr)
	if !ok {
		return nil, ast.ErrEval(fst.Src, "not a appendable value", nil)
	}
	vals, _ := c.Args[1].(*exp.Tupl)
	for _, el := range vals.Els {
		v, err := p.Eval(c.Env, el)
		if err != nil {
			return nil, err
		}
		err = mut.Append(v.Val)
		if err != nil {
			return nil, ast.ErrEval(v.Src, "append failed", err)
		}
	}
	return fst, nil
}

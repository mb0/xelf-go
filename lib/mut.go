package lib

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
)

var Mut = &mutSpec{impl("<form mut <alt list keyr> tupl? tupl?|tag .0>")}

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
	vals, _ := c.Args[1].(*exp.Tupl)
	if len(vals.Els) > 0 { // special modes
		k := mut.Type().Kind
		if k&knd.List != 0 { // append mode
			a, ok := mut.(lit.Apdr)
			if !ok {
				return nil, ast.ErrEval(fst.Src, "mut append expects list", nil)
			}
			for _, el := range vals.Els {
				v, err := p.Eval(c.Env, el)
				if err != nil {
					return nil, err
				}
				err = a.Append(v.Val)
				if err != nil {
					return nil, ast.ErrEval(v.Src, "mut append failed", err)
				}
			}
		} else if k&knd.Keyr != 0 { // merge mode
			a, ok := mut.(lit.Keyr)
			if !ok {
				return nil, ast.ErrEval(fst.Src, "mut merge expects keyr", nil)
			}
			for _, el := range vals.Els {
				v, err := p.Eval(c.Env, el)
				if err != nil {
					return nil, err
				}
				b, ok := v.Val.(lit.Keyr)
				if !ok {
					return nil, ast.ErrEval(v.Src, "mut merge not a keyr argument", nil)
				}
				err = b.IterKey(func(k string, v lit.Val) error {
					return a.SetKey(k, v)
				})
				if err != nil {
					return nil, ast.ErrEval(v.Src, "mut merge failed", err)
				}
			}
		} else {
			msg := fmt.Sprintf("mut unexpected special arguments for %s", mut.Type())
			return nil, ast.ErrEval(vals.Src, msg, nil)
		}
	}
	tags, _ := c.Args[2].(*exp.Tupl)
	for _, el := range tags.Els {
		tag := el.(*exp.Tag)
		ta, err := p.Eval(c.Env, tag.Exp)
		if err != nil {
			return nil, err
		}
		path, err := cor.ParsePath(tag.Tag)
		if err != nil {
			return nil, ast.ErrEval(tag.Src, "mut invalid path tag", nil)
		}
		err = lit.CreatePath(p.Reg, mut, path, ta.Val)
		if err != nil {
			return nil, ast.ErrEval(tag.Src, "mut create path", err)
		}
	}
	return fst, nil
}

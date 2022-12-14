package lib

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// TODO add a custom resl method to mut specs so we can detect incompatible types during resolution

var Mut = &mutSpec{impl("<form@mut any@ tupl?|exp _>")}

type mutSpec struct{ exp.SpecBase }

func (s *mutSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
	tupl, _ := c.Args[1].(*exp.Tupl)
	if len(tupl.Els) == 0 {
		return c.Args[0], nil
	}
	if sym, ok := tupl.Els[0].(*exp.Sym); ok && sym.Sym == "+" {
		tupl.Els[0] = &exp.Lit{Val: lit.AnyWrap(typ.Sym)}
	}
	return s.SpecBase.Resl(p, env, c, h)
}

func (s *mutSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	val, err := p.Eval(c.Env, c.Args[0])
	if err != nil {
		return nil, err
	}
	mut := val.Mut()
	tupl := c.Args[1].(*exp.Tupl)
	return mutate(p, c.Env, mut, tupl)
}

func mutate(p *exp.Prog, env exp.Env, mut lit.Mut, tupl *exp.Tupl) (lit.Mut, error) {
	fst := tupl.Els[0]
	if _, ok := fst.(*exp.Tag); ok {
		delta := make(lit.Delta, 0, len(tupl.Els))
		for _, el := range tupl.Els {
			tag, _ := el.(*exp.Tag)
			if tag == nil {
				// error
			}
			ta, err := p.Eval(env, tag.Exp)
			if err != nil {
				return nil, err
			}
			delta = append(delta, lit.KeyVal{Key: tag.Tag, Val: ta})
		}
		return lit.Apply(mut, delta)
	}
	v, err := p.Eval(env, fst)
	if err != nil {
		return nil, err
	}
	if v.Type() != typ.Sym {
		// we expect single argument for simple assignment
		if len(tupl.Els) > 1 {
			next := tupl.Els[1]
			return nil, ast.ErrUnexpectedExp(next.Source(), next)
		}
		err = mut.Assign(v)
		if err != nil {
			return nil, err
		}
		return mut, nil
	}
	switch mt := mut.Type(); mt.Kind & knd.Data {
	case knd.List, knd.Idxr:
		apdr, ok := mut.Value().(lit.Appender)
		if !ok {
			// TODO check if mut is simply nil and create a zero value
		}
		for _, el := range tupl.Els[1:] {
			v, err := p.Eval(env, el)
			if err != nil {
				return nil, err
			}
			err = apdr.Append(v)
			if err != nil {
				return nil, ast.ErrEval(el.Source(), "append failed", err)
			}
		}
		return mut, nil
	case knd.Str, knd.Raw, knd.Char:
		r, err := cat(p, env, mut, tupl.Els[1:])
		if err != nil {
			return nil, err
		}
		return mut, mut.Assign(r)
	case knd.Num, knd.Int, knd.Real:
		r, err := add(p, env, mut, tupl.Els[1:])
		if err != nil {
			return nil, err
		}
		return mut, mut.Assign(r)
	}
	err = ast.ErrUnexpectedExp(fst.Source(), fst)
	return nil, fmt.Errorf("mut append or merge syntax: %w", err)
}

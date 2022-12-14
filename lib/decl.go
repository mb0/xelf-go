package lib

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var With = &withSpec{impl("<form@with any exp|@1 @1>")}

type withSpec struct{ exp.SpecBase }

func (s *withSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	de, ok := c.Env.(*exp.DotEnv)
	if !ok {
		de = &exp.DotEnv{Par: env}
		c.Env = de
	}
	dot, err := p.Resl(env, c.Args[0], typ.Void)
	if err != nil {
		return c, err
	}
	c.Args[0] = dot
	if l, _ := dot.(*exp.Lit); l != nil {
		de.Dot = l.Val
	} else {
		de.Dot = lit.AnyWrap(typ.Res(dot.Type()))
	}
	res, err := p.Resl(de, c.Args[1], h)
	if err != nil {
		return c, err
	}
	c.Args[1] = res
	return c, nil
}
func (s *withSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	de := c.Env.(*exp.DotEnv)
	dot, err := p.Eval(de.Par, c.Args[0])
	if err != nil {
		return nil, err
	}
	de.Dot = dot
	return p.Eval(de, c.Args[1])
}

var Let = &letSpec{impl("<form@let tupl|tag exp|@1 @1>")}

type letSpec struct{ exp.SpecBase }

func (s *letSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	le, ok := c.Env.(*LetEnv)
	if !ok {
		le = &LetEnv{Par: env, Dot: lit.MakeObj(nil)}
		c.Env = le
	}
	tags := c.Args[0].(*exp.Tupl)
	for _, el := range tags.Els {
		tag := el.(*exp.Tag)
		ta, err := p.Resl(env, tag.Exp, typ.Void)
		if err != nil {
			return c, err
		}
		tag.Exp = ta
		p := typ.P(tag.Tag, typ.Res(ta.Type()))
		pb := le.Dot.Typ.Body.(*typ.ParamBody)
		pb.Params = append(pb.Params, p)
		le.Dot.Vals = append(le.Dot.Vals, lit.AnyWrap(p.Type))
	}
	res, err := p.Resl(le, c.Args[1], h)
	if err != nil {
		return c, err
	}
	c.Args[1] = res
	return c, nil
}
func (s *letSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	le := c.Env.(*LetEnv)
	tags := c.Args[0].(*exp.Tupl)
	for i, el := range tags.Els {
		tag := el.(*exp.Tag)
		v, err := p.Eval(le.Par, tag.Exp)
		if err != nil {
			return nil, err
		}
		le.Dot.Vals[i].Mut().Assign(v)
	}
	return p.Eval(le, c.Args[1])
}

type LetEnv struct {
	Par exp.Env
	Dot *lit.Obj
}

func (e *LetEnv) Parent() exp.Env { return e.Par }
func (e *LetEnv) Lookup(s *exp.Sym, p cor.Path, eval bool) (lit.Val, error) {
	v, err := exp.SelectLookup(e.Dot, p, eval)
	if err == nil && v != nil {
		s.Update(v.Type(), e, p)
		if !eval && v.Nil() {
			return nil, nil
		}
		return v, nil
	}
	return e.Par.Lookup(s, p, eval)
}

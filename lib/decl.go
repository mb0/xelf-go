package lib

import (
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
	dot, err := p.Resl(env, c.Args[0], typ.Any)
	if err != nil {
		return c, err
	}
	de.Dot = dot.(*exp.Lit).Val
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
		le = &LetEnv{Par: env}
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
		le.Dot.SetKey(tag.Tag, lit.AnyWrap(typ.Res(ta.Type())))
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
	for _, el := range tags.Els {
		tag := el.(*exp.Tag)
		v, err := p.Eval(le.Par, tag.Exp)
		if err != nil {
			return nil, err
		}
		le.Dot.SetKey(tag.Tag, v)
	}
	return p.Eval(le, c.Args[1])
}

type LetEnv struct {
	Par exp.Env
	Dot lit.Keyed
}

func (e *LetEnv) Parent() exp.Env { return e.Par }
func (e *LetEnv) Lookup(s *exp.Sym, k string, eval bool) (lit.Val, error) {
	v, err := exp.SelectLookup(&e.Dot, k, eval)
	if err == nil && v != nil {
		if s.Update(v.Type(), e, k); !eval {
			return nil, nil
		}
		return v, nil
	}
	return e.Par.Lookup(s, k, eval)
}

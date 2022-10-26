package lib

import (
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Dot = &dotSpec{impl("<form@dot any exp|@1 @1>")}

type dotSpec struct{ exp.SpecBase }

func (s *dotSpec) Value() lit.Val { return s }
func (s *dotSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	de, ok := c.Env.(*exp.DotEnv)
	if !ok {
		de = &exp.DotEnv{Par: env, Dot: &exp.Lit{}}
		c.Env = de
	}
	dot, err := p.Resl(env, c.Args[0], typ.Any)
	if err != nil {
		return c, err
	}
	de.Dot.Res = resType(dot.Resl())
	res, err := p.Resl(de, c.Args[1], h)
	if err != nil {
		return c, err
	}
	c.Args[1] = res
	return c, nil
}
func (s *dotSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	de := c.Env.(*exp.DotEnv)
	dot, err := p.Eval(de.Par, c.Args[0])
	if err != nil {
		return nil, err
	}
	*de.Dot = *dot
	return p.Eval(de, c.Args[1])
}

var Let = &letSpec{impl("<form@let tupl|tag exp|@1 @1>")}

type letSpec struct{ exp.SpecBase }

func (s *letSpec) Value() lit.Val { return s }
func (s *letSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	le, ok := c.Env.(*LetEnv)
	if !ok {
		le = &LetEnv{Par: env, Lets: make(map[string]*exp.Lit)}
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
		a := le.Lets[tag.Tag]
		if a == nil {
			a = &exp.Lit{Res: ta.Resl()}
			le.Lets[tag.Tag] = a
		} else {
			a.Res = ta.Resl()
		}
	}
	res, err := p.Resl(le, c.Args[1], h)
	if err != nil {
		return c, err
	}
	c.Args[1] = res
	return c, nil
}
func (s *letSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	le := c.Env.(*LetEnv)
	tags := c.Args[0].(*exp.Tupl)
	for _, el := range tags.Els {
		tag := el.(*exp.Tag)
		ta, err := p.Eval(le.Par, tag.Exp)
		if err != nil {
			return nil, err
		}
		a := le.Lets[tag.Tag]
		*a = *ta
	}
	return p.Eval(le, c.Args[1])
}

type LetEnv struct {
	Par  exp.Env
	Lets map[string]*exp.Lit
}

func (e *LetEnv) Parent() exp.Env { return e.Par }
func (e *LetEnv) Lookup(s *exp.Sym, k string, eval bool) (exp.Exp, error) {
	if a := e.Lets[k]; a != nil {
		if eval {
			return a, nil
		}
		s.Type, s.Env, s.Rel = a.Res, e, k
		return s, nil
	}
	return e.Par.Lookup(s, k, eval)
}

func resType(t typ.Type) typ.Type {
	if t.Kind&knd.Exp == knd.Exp || t.Kind&knd.Exp != 0 {
		if b, ok := t.Body.(*typ.ElBody); ok && b.El.Kind != knd.Void {
			return b.El
		}
		return typ.Any
	}
	return t
}

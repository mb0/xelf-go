package lib

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var With = &withSpec{impl("<form@with dot?:any lets:tupl?|tag exp|@1 @1>")}

type withSpec struct{ exp.SpecBase }

func (s *withSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	de, ok := c.Env.(*DotEnv)
	if !ok {
		de = &DotEnv{Par: env, Lets: lit.MakeObj(nil)}
		c.Env = de
	}
	if a := c.Args[0]; a != nil {
		dot, err := p.Resl(env, a, typ.Void)
		if err != nil {
			return c, err
		}
		c.Args[0] = dot
		if l, _ := dot.(*exp.Lit); l != nil {
			de.Dot = l.Val
		} else {
			de.Dot = lit.AnyWrap(typ.Res(dot.Type()))
		}
	} else {
		de.Dot = lit.Null{}
	}
	if a := c.Args[1]; a != nil {
		tags := a.(*exp.Tupl)
		for _, el := range tags.Els {
			tag := el.(*exp.Tag)
			ta, err := p.Resl(de, tag.Exp, typ.Void)
			if err != nil {
				return c, err
			}
			tag.Exp = ta
			p := typ.P(tag.Tag, typ.Res(ta.Type()))
			pb := de.Lets.Typ.Body.(*typ.ParamBody)
			pb.Params = append(pb.Params, p)
			de.Lets.Vals = append(de.Lets.Vals, lit.AnyWrap(p.Type))
		}
	}
	res, err := p.Resl(de, c.Args[2], h)
	if err != nil {
		return c, err
	}
	c.Args[2] = res
	return c, nil
}
func (s *withSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	de := c.Env.(*DotEnv)
	if a := c.Args[0]; a != nil {
		dot, err := p.Eval(de.Par, a)
		if err != nil {
			return nil, err
		}
		de.Dot = dot
	}
	if a := c.Args[1]; a != nil {
		tags := a.(*exp.Tupl)
		for i, el := range tags.Els {
			tag := el.(*exp.Tag)
			v, err := p.Eval(de, tag.Exp)
			if err != nil {
				return nil, err
			}
			de.Lets.Vals[i].Mut().Assign(v)
		}
	}
	return p.Eval(de, c.Args[2])
}

// DotEnv is a child scope that supports relative paths into either a dot literals or name values.
type DotEnv struct {
	Par  exp.Env
	Dot  lit.Val
	Lets *lit.Obj
}

func (e *DotEnv) Parent() exp.Env { return e.Par }
func (e *DotEnv) Lookup(s *exp.Sym, p cor.Path, eval bool) (lit.Val, error) {
	p, ok := exp.DotPath(p)
	if ok {
		v, err := exp.SelectLookup(e.Dot, p, eval)
		if err == nil && v != nil {
			s.Update(v.Type(), e, p)
			if !eval && v.Nil() {
				return nil, nil
			}
			return v, nil
		}
		return nil, exp.ErrSymNotFound
	}
	v, err := exp.SelectLookup(e.Lets, p, eval)
	if err == nil && v != nil {
		s.Update(v.Type(), e, p)
		if !eval && v.Nil() {
			return nil, nil
		}
		return v, nil
	}
	return e.Par.Lookup(s, p, eval)
}

package lib

import (
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Make = &makeSpec{impl("<form@make typ tupl?|exp lit|_>")}

type makeSpec struct{ exp.SpecBase }

func (s *makeSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
	fst, err := p.Resl(env, c.Args[0], typ.Typ)
	if err != nil {
		return nil, err
	}
	at, ok := fst.(*exp.Lit)
	if !ok {
		c.Env = env
		return c, nil
	}
	t, err := typ.ToType(at.Val)
	if err != nil {
		return c, err
	}
	tupl, _ := c.Args[1].(*exp.Tupl)
	if len(tupl.Els) == 0 {
		return exp.LitSrc(p.Reg.ZeroWrap(t), c.Src), nil
	}
	if len(tupl.Els) > 0 {
		if sym, ok := tupl.Els[0].(*exp.Sym); ok && sym.Sym == "+" {
			tupl.Els[0] = &exp.Lit{Val: lit.AnyWrap(typ.Sym)}
		}
	}
	rp := exp.SigRes(c.Sig)
	rp.Type = t
	c.Sig, err = p.Sys.Update(c.Sig)
	if err != nil {
		return nil, err
	}
	return s.SpecBase.Resl(p, env, c, h)
}
func (s *makeSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	fst, err := p.Eval(c.Env, c.Args[0])
	if err != nil {
		return nil, err
	}
	t, err := typ.ToType(fst)
	if err != nil {
		return nil, err
	}
	wrap := p.Reg.ZeroWrap(typ.Res(t))
	tupl, _ := c.Args[1].(*exp.Tupl)
	return mutate(p, c.Env, wrap, tupl)
}

package lib

import (
	"fmt"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Do = &doSpec{impl("<form@do tupl|exp @>")}

type doSpec struct{ exp.SpecBase }

func (s *doSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
	// we must resolve all expressions in order for side effects
	els := c.Args[0].(*exp.Tupl).Els
	switch len(els) {
	case 0:
		return nil, fmt.Errorf("empty do call")
	case 1:
		return p.Resl(env, els[0], h)
	}
	if c.Env == nil {
		c.Env = env
	}
	var lst exp.Exp
	for i, e := range els {
		eh := typ.Void
		if i == len(els)-1 {
			eh = h
		}
		res, err := p.Resl(env, e, eh)
		if err != nil {
			return c, err
		}
		if res != e {
			els[i] = res
		}
		lst = res
	}
	if lst != nil {
		rp := exp.SigRes(c.Sig)
		rp.Type = typ.Res(lst.Type())
		var err error
		c.Sig, err = p.Sys.Update(c.Sig)
		if err != nil {
			return c, err
		}
	}
	return c, nil
}
func (s *doSpec) Eval(p *exp.Prog, c *exp.Call) (lst lit.Val, _ error) {
	// we must evaluate all expressions in order for side effects
	d := c.Args[0].(*exp.Tupl)
	for _, e := range d.Els {
		res, err := p.Eval(c.Env, e)
		if err != nil {
			return nil, err
		}
		lst = res
	}
	return lst, nil
}

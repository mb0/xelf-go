package lib

import (
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Use = &useSpec{impl("<form@use name:sym path?:lit|str void>")}

type useSpec struct{ exp.SpecBase }

func (s *useSpec) Value() lit.Val { return s }
func (s *useSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	if c.Env == nil {
		c.Env = env
		name := c.Args[0].String()
		p.Reg.SetRef(name, typ.Ref(name), nil)
	}
	return c, nil
}
func (s *useSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	return &exp.Lit{}, nil
}

package lib

import (
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

var If = &ifSpec{impl("<form@if <tupl cond:any then:exp|@1> else:exp?|@1 @1>")}

type ifSpec struct{ exp.SpecBase }

func (s *ifSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	els := c.Args[0].(*exp.Tupl).Els
	for i := 0; i < len(els); i += 2 {
		res, err := p.Eval(c.Env, els[i])
		if err != nil {
			return nil, err
		}
		if !res.Val.Zero() {
			return p.Eval(c.Env, els[i+1])
		}
	}
	// else
	if c.Args[1] != nil {
		return p.Eval(c.Env, c.Args[1])
	}
	rt := exp.SigRes(c.Sig).Type
	return exp.LitSrc(lit.ZeroWrap(rt), c.Src), nil
}

var Swt = &swtSpec{impl("<form@swt @1 <tupl case:@1 then:exp|@2> else:exp?|@2 @2>")}

type swtSpec struct{ exp.SpecBase }

func (s *swtSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	arg, err := p.Eval(c.Env, c.Args[0])
	if err != nil {
		return nil, err
	}
	// cases
	els := c.Args[1].(*exp.Tupl).Els
	for i := 0; i < len(els); i += 2 {
		res, err := p.Eval(c.Env, els[i])
		if err != nil {
			return nil, err
		}
		if ok := lit.Equal(arg.Val, res.Val); ok {
			return p.Eval(c.Env, els[i+1])
		}
	}
	// else
	if c.Args[2] != nil {
		return p.Eval(c.Env, c.Args[2])
	}
	rt := exp.SigRes(c.Sig).Type
	return exp.LitSrc(lit.ZeroWrap(rt), c.Src), nil
}

var Df = &dfSpec{impl("<form@df tupl|@1 @1!>")}

type dfSpec struct{ exp.SpecBase }

func (s *dfSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	// cases
	for _, cas := range c.Args[0].(*exp.Tupl).Els {
		res, err := p.Eval(c.Env, cas)
		if err != nil {
			return nil, err
		}
		if !res.Val.Zero() {
			return res, nil
		}
	}
	// else
	t := exp.SigRes(c.Sig).Type
	return exp.LitSrc(lit.ZeroWrap(t), c.Src), nil
}

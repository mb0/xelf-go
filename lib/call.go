package lib

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Call = &callSpec{impl("<form@call spec tupl|exp lit|_>")}

type callSpec struct{ exp.SpecBase }

func (s *callSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	if c.Env == nil {
		c.Env = env
	}
	fst := c.Args[0]
	l, err := p.Resl(env, fst, typ.Spec)
	if err != nil {
		return nil, ast.ErrEval(fst.Source(), fmt.Sprintf("call resl failed for %s", fst), err)
	}
	if l.Type().Kind&knd.Lit == 0 {
		// TODO lets check the resolved type of fst
		// we may know a sugar spec for lit res or impossible specs, if we expect a spec
		// we need to evaluate the argument before we can determine the spec
		return c, nil
	}
	tupl := c.Args[1].(*exp.Tupl)
	spec, _ := l.(*exp.Lit).Value().(*exp.SpecRef)
	sig, _ := p.Sys.Inst(exp.LookupType(env), spec.Type())
	args, err := exp.LayoutSpec(sig, tupl.Els)
	if err != nil {
		return nil, fmt.Errorf("layout error for %s %s: %v", sig.Type(), args, err)
	}
	cc := &exp.Call{Sig: sig, Spec: spec, Args: args, Src: c.Src}
	return spec.Resl(p, env, cc, h)
}
func (s *callSpec) Eval(p *exp.Prog, c *exp.Call) (_ lit.Val, err error) {
	fst := c.Args[0]
	l, err := p.Eval(c.Env, fst)
	if err != nil {
		return nil, ast.ErrEval(fst.Source(), fmt.Sprintf("call eval failed for %s", fst), err)
	}
	spec, _ := l.Value().(*exp.SpecRef)
	sig, _ := p.Sys.Inst(exp.LookupType(c.Env), spec.Type())
	tupl := c.Args[1].(*exp.Tupl)
	args, err := exp.LayoutSpec(sig, tupl.Els)
	if err != nil {
		return nil, fmt.Errorf("layout error for %s %s: %v", sig.Type(), args, err)
	}
	cc := &exp.Call{Sig: sig, Spec: spec, Args: args, Src: c.Src}
	res, err := spec.Resl(p, c.Env, cc, typ.Void)
	if err != nil {
		return nil, err
	}
	return p.Eval(c.Env, res)
}

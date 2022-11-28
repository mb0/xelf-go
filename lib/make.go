package lib

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Make = &makeSpec{impl("<form@make typ tupl? tupl?|tag lit|_>")}

type makeSpec struct{ exp.SpecBase }

func (s *makeSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
	if c.Env == nil {
		c.Env = env
	}
	fst, err := p.Resl(env, c.Args[0], typ.Typ)
	if err != nil {
		return nil, err
	}
	at, ok := fst.(*exp.Lit)
	if !ok {
		return c, nil
	}
	t, err := typ.ToType(at.Val)
	if err != nil {
		return c, err
	}
	args, aok := c.Args[1].(*exp.Tupl)
	tags, tok := c.Args[2].(*exp.Tupl)
	if (!aok || len(args.Els) == 0) && (!tok || len(tags.Els) == 0) {
		return &exp.Lit{Res: t, Val: p.Reg.Zero(t), Src: c.Src}, nil
	}
	rp := exp.SigRes(c.Sig)
	rp.Type = t
	c.Sig, err = p.Sys.Update(c.Sig)
	if err != nil {
		return nil, err
	}
	return s.SpecBase.Resl(p, env, c, h)
}
func (s *makeSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	fst, err := p.Eval(c.Env, c.Args[0])
	if err != nil {
		return nil, err
	}
	t, err := typ.ToType(fst.Val)
	if err != nil {
		return nil, err
	}
	t = typ.ResEl(t)
	els, err := p.Eval(c.Env, c.Args[1])
	if err != nil {
		return nil, err
	}
	plain, pok := els.Val.(*lit.List)
	pok = pok && len(plain.Vals) > 0
	tags, tok := c.Args[2].(*exp.Tupl)
	tok = tok && len(tags.Els) > 0
	res := p.Reg.Zero(t)
	if pok {
		apdr, ok := res.(lit.Apdr)
		if ok {
			for _, v := range plain.Vals {
				err = apdr.Append(v)
				if err != nil {
					return nil, err
				}
			}
		} else if len(plain.Vals) != 1 {
			return nil, fmt.Errorf("make non-idxr type %s for vals", t)
		} else {
			err = res.Assign(plain.Vals[0])
			if err != nil {
				return nil, err
			}
		}
	}
	if tok {
		keyr, ok := res.(lit.Keyr)
		if !ok {
			return nil, ast.ErrEval(c.Src, c.Sig.Ref, fmt.Errorf("make non-keyr type %s for tags", t))
		}
		// eval tags and set keyr
		for _, el := range tags.Els {
			tag := el.(*exp.Tag)
			var tv lit.Val
			if tag.Exp != nil {
				ta, err := p.Eval(c.Env, tag.Exp)
				if err != nil {
					return nil, err
				}
				tv = ta.Val
			}
			err = keyr.SetKey(tag.Tag, tv)
			if err != nil {
				return nil, err
			}
		}
	}
	return &exp.Lit{Res: t, Val: res.Value(), Src: c.Src}, nil
}

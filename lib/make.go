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
		return exp.LitSrc(p.Reg.ZeroWrap(t), c.Src), nil
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
	els, err := p.Eval(c.Env, c.Args[1])
	if err != nil {
		return nil, err
	}
	plain, pok := els.(*lit.List)
	pok = pok && len(plain.Vals) > 0
	tags, tok := c.Args[2].(*exp.Tupl)
	tok = tok && len(tags.Els) > 0
	if pok {
		res := lit.Unwrap(wrap)
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
			err = wrap.Assign(plain.Vals[0])
			if err != nil {
				return nil, err
			}
		}
	}
	if tok {
		res := lit.Unwrap(wrap)
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
				tv = ta
			}
			err = keyr.SetKey(tag.Tag, tv)
			if err != nil {
				return nil, err
			}
		}
	}
	return wrap, nil
}

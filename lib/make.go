package lib

import (
	"fmt"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Make = &makeSpec{impl("<form make typ tupl? tupl?|tag lit|.0>")}

type makeSpec struct{ exp.SpecBase }

func (s *makeSpec) Value() lit.Val { return s }
func (s *makeSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
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
		mut, err := p.Reg.Zero(t)
		if err != nil {
			return c, err
		}
		return &exp.Lit{Res: t, Val: mut.Value(), Src: c.Src}, nil
	}
	rp := exp.SigRes(c.Sig)
	rp.Type = t
	c.Sig = p.Sys.Update(c.Sig)
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
	res, err := p.Reg.Zero(t)
	if err != nil {
		return nil, err
	}
	if tok {
		keyr, ok := res.(lit.Keyr)
		if !ok {
			return nil, fmt.Errorf("make non-keyr type %s for tags", t)
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
	} else if pok {
		apdr, ok := res.(lit.Apdr)
		if !ok {
			if len(plain.Vals) == 1 {
				err := res.Assign(plain.Vals[0])
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("make non-idxr type %s for vals", t)
			}
		} else {
			et := typ.ContEl(res.Type())
			for _, v := range plain.Vals {
				if et != typ.Any {
					prx, err := p.Reg.Zero(et)
					if err != nil {
						return nil, err
					}
					err = prx.Assign(v)
					if err != nil {
						return nil, err
					}
					v = prx
				}
				err = apdr.Append(v)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return &exp.Lit{Res: t, Val: res.Value(), Src: c.Src}, nil
}

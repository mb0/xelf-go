package lib

import (
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Add = &addSpec{impl("<form add num@ tupl?|num@ .0>")}

type addSpec struct{ exp.SpecBase }

func (s *addSpec) Value() lit.Val { return s }
func (s *addSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	r, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	for _, v := range args[1].Val.(*lit.List).Vals {
		rr, err := lit.ToReal(v)
		if err != nil {
			return nil, err
		}
		r += rr
	}
	return toNum(c.Sig, r)
}

var Mul = &mulSpec{impl("<form mul num@ tupl?|num@ .0>")}

type mulSpec struct{ exp.SpecBase }

func (s *mulSpec) Value() lit.Val { return s }
func (s *mulSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	r, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	for _, v := range args[1].Val.(*lit.List).Vals {
		rr, err := lit.ToReal(v)
		if err != nil {
			return nil, err
		}
		r *= rr
	}
	return toNum(c.Sig, r)
}

var Sub = &subSpec{impl("<form sub num@ tupl|num@ .0>")}

type subSpec struct{ exp.SpecBase }

func (s *subSpec) Value() lit.Val { return s }
func (s *subSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	f, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	var r lit.Real
	for _, v := range args[1].Val.(*lit.List).Vals {
		rr, err := lit.ToReal(v)
		if err != nil {
			return nil, err
		}
		r += rr
	}
	return toNum(c.Sig, f-r)
}

var Div = &divSpec{impl("<form div num@ tupl|num@ .0>")}

type divSpec struct{ exp.SpecBase }

func (s *divSpec) Value() lit.Val { return s }
func (s *divSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	f, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	var r lit.Real = 1
	for _, v := range args[1].Val.(*lit.List).Vals {
		rr, err := lit.ToReal(v)
		if err != nil {
			return nil, err
		}
		r *= rr
	}
	return toNum(c.Sig, f/r)
}

var Rem = &remSpec{impl("<form rem int int int>")}

type remSpec struct{ exp.SpecBase }

func (s *remSpec) Value() lit.Val { return s }
func (s *remSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	f, err := lit.ToInt(args[0].Val)
	if err != nil {
		return nil, err
	}
	r, err := lit.ToInt(args[1].Val)
	if err != nil {
		return nil, err
	}
	return &exp.Lit{Res: typ.Int, Val: f % r}, nil
}

var Abs = &absSpec{impl("<form abs num@ .0>")}

type absSpec struct{ exp.SpecBase }

func (s *absSpec) Value() lit.Val { return s }
func (s *absSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	r, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	if r < 0 {
		r = -r
	}
	return toNum(c.Sig, r)
}

var Neg = &negSpec{impl("<form neg num@ .0>")}

type negSpec struct{ exp.SpecBase }

func (s *negSpec) Value() lit.Val { return s }
func (s *negSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	r, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	return toNum(c.Sig, -r)
}

var Min = &minSpec{impl("<form min num@ tupl?|num@ .0>")}

type minSpec struct{ exp.SpecBase }

func (s *minSpec) Value() lit.Val { return s }
func (s *minSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	r, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	for _, v := range args[1].Val.(*lit.List).Vals {
		rr, err := lit.ToReal(v)
		if err != nil {
			return nil, err
		}
		if rr < r {
			r = rr
		}
	}
	return toNum(c.Sig, r)
}

var Max = &maxSpec{impl("<form max num@ tupl?|num@ .0>")}

type maxSpec struct{ exp.SpecBase }

func (s *maxSpec) Value() lit.Val { return s }
func (s *maxSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	r, err := lit.ToReal(args[0].Val)
	if err != nil {
		return nil, err
	}
	for _, v := range args[1].Val.(*lit.List).Vals {
		rr, err := lit.ToReal(v)
		if err != nil {
			return nil, err
		}
		if rr > r {
			r = rr
		}
	}
	return toNum(c.Sig, r)
}

func toNum(sig typ.Type, r lit.Real) (*exp.Lit, error) {
	t := exp.SigRes(sig).Type
	if t.Kind&knd.Int != 0 {
		i := lit.Int(r)
		if t.Kind&knd.Num == knd.Int || r == lit.Real(i) {
			return &exp.Lit{Res: t, Val: i}, nil
		}
	}
	return &exp.Lit{Res: t, Val: r}, nil
}

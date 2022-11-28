package lib

import (
	"fmt"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

type compSpec struct {
	exp.SpecBase
	want int8
	neg  bool
}

var (
	Eq    = &compSpec{impl("<form@eq any tupl bool>"), 0, false}
	Equal = &compSpec{impl("<form@equal any tupl bool>"), 0, false}
	Ne    = &compSpec{impl("<form@ne any tupl bool>"), 0, true}
	Lt    = &compSpec{impl("<form@lt <alt num str span time> tupl|_ bool>"), -1, false}
	Ge    = &compSpec{impl("<form@ge <alt num str span time> tupl|_ bool>"), -1, true}
	Gt    = &compSpec{impl("<form@gt <alt num str span time> tupl|_ bool>"), 1, false}
	Le    = &compSpec{impl("<form@le <alt num str span time> tupl|_ bool>"), 1, true}
)

func (s *compSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	r := true
	cur := args[0].Val
	for _, v := range args[1].Val.(*lit.List).Vals {
		if s.want == 0 {
			ok := lit.Equal(cur, v)
			if s.neg == ok {
				r = false
			}
		} else {
			cmp, err := lit.Compare(cur, v)
			if err != nil {
				return nil, err
			}
			if (cmp == s.want) != s.neg {
				cur = v
			} else {
				r = false
			}
		}
		if !r {
			break
		}
	}
	return &exp.Lit{Res: typ.Bool, Val: lit.Bool(r)}, nil
}

var (
	In = &inSpec{impl("<form@in any tupl|list bool>"), false}
	Ni = &inSpec{impl("<form@ni any tupl|list bool>"), true}
)

type inSpec struct {
	exp.SpecBase
	neg bool
}

func (s *inSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	fst := args[0].Val
	var r bool
	var vals *lit.Vals
	switch v := args[1].Val.(type) {
	case *lit.Vals:
		vals = v
	case *lit.List:
		vals = &v.Vals
	default:
		return nil, fmt.Errorf("unexpected idxr %T", args[1].Val)
	}
	for _, val := range *vals {
		var vv *lit.Vals
		switch v := val.(type) {
		case *lit.Vals:
			vv = v
		case *lit.List:
			vv = &v.Vals
		default:
			return nil, fmt.Errorf("unexpected idxr %T", val)
		}
		for _, v := range *vv {
			if lit.Equal(fst, v) {
				r = true
				break
			}
		}
	}
	if s.neg {
		r = !r
	}
	return &exp.Lit{Res: typ.Bool, Val: lit.Bool(r)}, nil
}

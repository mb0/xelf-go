package lib

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Len = &lenSpec{impl("<form@len <alt? list keyr str raw> int>")}

type lenSpec struct{ exp.SpecBase }

func (s *lenSpec) Value() lit.Val { return s }
func (s *lenSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	var n int
	v := args[0].Val
	if v != nil && !v.Nil() {
		lenr, ok := v.(lit.Lenr)
		if ok {
			n = lenr.Len()
		} else {
			return nil, fmt.Errorf("unexpected argument %[1]T %[1]%s", v)
		}
	}
	return &exp.Lit{Res: typ.Int, Val: lit.Int(n)}, nil
}

var Fold = &foldSpec{impl("<form@fold list|@1 @2 <func @2 @1 @2> @2>"), false}
var Foldr = &foldSpec{impl("<form@foldr list|@1 @2 <func @2 @1 @2> @2>"), true}

type foldSpec struct {
	exp.SpecBase
	Right bool
}

func (s *foldSpec) Value() lit.Val { return s }
func (s *foldSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	trd := args[2]
	fun, ok := trd.Val.(exp.Spec)
	if !ok {
		return nil, fmt.Errorf("unexpected func %[1]T %[1]s", trd)
	}
	res := args[1]
	var vals *lit.Vals
	switch v := args[0].Val.(type) {
	case *lit.Vals:
		vals = v
	case *lit.List:
		vals = &v.Vals
	}
	if vals != nil {
		vs := *vals
		for i, el := range vs {
			if s.Right {
				el = vs[len(vs)-1-i]
			}
			args := []exp.Exp{res, &exp.Lit{Res: el.Type(), Val: el}}
			r, err := callFunc(p, c, fun, args, trd.Src)
			if err != nil {
				return nil, err
			}
			res.Val = r
		}
		return res, nil
	}
	return nil, fmt.Errorf("unexpected idxr %[1]T %[1]s", args[0])
}

var Range = &rangeSpec{impl("<form@range n:int f?:<func int @1> list|@1>")}

type rangeSpec struct{ exp.SpecBase }

func (s *rangeSpec) Value() lit.Val { return s }
func (s *rangeSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	fst := args[0]
	n, err := lit.ToInt(fst.Val)
	if err != nil {
		return nil, err
	}
	var list *lit.List
	res := make([]lit.Val, n)
	if snd := args[1]; snd != nil {
		f, ok := snd.Val.(exp.Spec)
		if !ok {
			return nil, fmt.Errorf("unexpected func %[1]T %[1]s", snd)
		}
		farg := &exp.Lit{Res: typ.Int, Src: fst.Src}
		fargs := []exp.Exp{farg}
		for i := range res {
			farg.Val = lit.Int(i)
			res[i], err = callFunc(p, c, f, fargs, snd.Src)
			if err != nil {
				return nil, err
			}
		}
		list = &lit.List{El: exp.SigRes(f.Type()).Type, Vals: res}
		return &exp.Lit{Res: list.Type(), Val: list, Src: c.Src}, nil
	} else {
		for i := range res {
			res[i] = lit.Int(i)
		}
		list = &lit.List{El: typ.Int, Vals: res}
	}
	return &exp.Lit{Res: list.Type(), Val: list, Src: c.Src}, nil
}

func callFunc(p *exp.Prog, c *exp.Call, s exp.Spec, org []exp.Exp, src ast.Src) (lit.Val, error) {
	sig := s.Type()
	args, err := exp.LayoutSpec(sig, org)
	if err != nil {
		return nil, fmt.Errorf("layout fun call %s with %s\n\t%v", sig, org, err)
	}
	cc := &exp.Call{Sig: sig, Spec: s, Args: args}
	_, err = s.Resl(p, c.Env, cc, typ.Void)
	if err != nil {
		return nil, err
	}
	res, err := s.Eval(p, cc)
	if err != nil {
		return nil, err
	}
	return res.Val, nil
}

package lib

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Len = &lenSpec{impl("<form len <alt@? list keyr str raw> int>")}

type lenSpec struct{ exp.SpecBase }

func (s *lenSpec) Value() lit.Val { return s }
func (s *lenSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	var n int
	switch v := args[0].Val.(type) {
	case lit.Null:
	case lit.Str:
		n = len(v)
	case *lit.Str:
		if v != nil {
			n = len(*v)
		}
	case lit.Raw:
		n = len(v)
	case *lit.Raw:
		if v != nil {
			n = len(*v)
		}
	case *lit.List:
		n = len(v.Vals)
	case *lit.Dict:
		n = len(v.Keyed)
	case *lit.Map:
		n = len(v.M)
	case *lit.Strc:
		n = v.Len()
	case *lit.ListPrx:
		n = v.Len()
	case *lit.MapPrx:
		n = v.Len()
	case *lit.StrcPrx:
		n = v.Len()
	default:
		lenr, ok := v.(interface{ Len() int })
		if ok {
			n = lenr.Len()
		} else {
			return nil, fmt.Errorf("unexpected argument %[1]T %[1]%s", v)
		}
	}
	return &exp.Lit{Res: typ.Int, Val: lit.Int(n)}, nil
}

var Fold = &foldSpec{impl("<form fold list|@1 @2 <func @2 @1 @2> @2>"), false}
var Foldr = &foldSpec{impl("<form foldr list|@1 @2 <func @2 @1 @2> @2>"), true}

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
	switch v := args[0].Val.(type) {
	case *lit.List:
		for i, el := range v.Vals {
			if s.Right {
				el = v.Vals[len(v.Vals)-1-i]
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

func callFunc(p *exp.Prog, c *exp.Call, s exp.Spec, args []exp.Exp, src ast.Src) (lit.Val, error) {
	sig := s.Type()
	args, err := exp.LayoutSpec(sig, args)
	if err != nil {
		return nil, err
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

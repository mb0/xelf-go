package lib

import (
	"bytes"
	"strings"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Cat = &catSpec{impl("<form cat tupl str>")}

type catSpec struct{ exp.SpecBase }

func (s *catSpec) Value() lit.Val { return s }
func (s *catSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for _, v := range args[0].Val.(*lit.List).Vals {
		b.WriteString(v.String())
	}
	return &exp.Lit{Res: typ.Str, Val: lit.Str(b.String())}, nil
}

var Sep = &sepSpec{impl("<form sep str tupl str>")}

type sepSpec struct{ exp.SpecBase }

func (s *sepSpec) Value() lit.Val { return s }
func (s *sepSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	sep, err := lit.ToStr(args[0])
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for i, v := range args[1].Val.(*lit.List).Vals {
		if i > 0 {
			b.WriteString(string(sep))
		}
		b.WriteString(v.String())
	}
	return &exp.Lit{Res: typ.Str, Val: lit.Str(b.String())}, nil
}

var Json = &rawSpec{impl("<form json any raw>"), true}
var Xelf = &rawSpec{impl("<form xelf any any raw>"), false}

type rawSpec struct {
	exp.SpecBase
	JSON bool
}

func (s *rawSpec) Value() lit.Val { return s }
func (s *rawSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	fst := args[0]
	var b bytes.Buffer
	if fst == nil || fst.Val == nil {
		b.WriteString("null")
	} else {
		ctx := &bfr.P{Writer: &b, JSON: s.JSON}
		err = args[0].Val.Print(ctx)
		if err != nil {
			return nil, err
		}
	}
	return &exp.Lit{Res: typ.Str, Val: lit.Raw(b.Bytes())}, nil
}

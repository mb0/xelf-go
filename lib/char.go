package lib

import (
	"bytes"
	"strings"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

var Cat = &catSpec{impl("<form@cat tupl str>")}

type catSpec struct{ exp.SpecBase }

func (s *catSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for _, v := range args[0].(*lit.List).Vals {
		b.WriteString(v.String())
	}
	return lit.Str(b.String()), nil
}

var Sep = &sepSpec{impl("<form@sep str tupl str>")}

type sepSpec struct{ exp.SpecBase }

func (s *sepSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	sep, err := lit.ToStr(args[0])
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	for i, v := range args[1].(*lit.List).Vals {
		if i > 0 {
			b.WriteString(string(sep))
		}
		b.WriteString(v.String())
	}
	return lit.Str(b.String()), nil
}

var Json = &rawSpec{impl("<form@json any raw>"), true}
var Xelf = &rawSpec{impl("<form@xelf any raw>"), false}

type rawSpec struct {
	exp.SpecBase
	JSON bool
}

func (s *rawSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	fst := args[0]
	var b bytes.Buffer
	if fst == nil {
		b.WriteString("null")
	} else {
		ctx := &bfr.P{Writer: &b, JSON: s.JSON}
		err = fst.Print(ctx)
		if err != nil {
			return nil, err
		}
	}
	return lit.Raw(b.Bytes()), nil
}

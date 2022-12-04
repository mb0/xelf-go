package lib

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

var Sel = &selSpec{impl("<form@sel sym <tupl?|alt str int> @>")}

type selSpec struct{ exp.SpecBase }

func (s *selSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	args, err := p.EvalArgs(c)
	if err != nil {
		return nil, err
	}
	tup := args[1].Value().(*lit.List)
	vars := make([]string, 0, len(tup.Vals))
	for _, val := range tup.Vals {
		vars = append(vars, val.String())
	}
	sym := args[0].String()
	path, err := cor.FillPath(sym, vars...)
	if err != nil {
		return nil, err
	}
	return p.Eval(c.Env, &exp.Sym{Sym: sym, Env: c.Env, Path: path})
}

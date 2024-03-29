package lib

import (
	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

var (
	Or  = &logicSpec{impl("<form@or  tupl? bool>"), false, false, true}
	Ok  = &logicSpec{impl("<form@ok  tupl? bool>"), false, true, false}
	And = &logicSpec{impl("<form@and tupl? bool>"), true, true, false}
	Not = &logicSpec{impl("<form@not tupl? bool>"), true, true, true}
)

type logicSpec struct {
	exp.SpecBase
	zero, init, neg bool
}

func (s *logicSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	r := s.zero
	args := c.Args[0].(*exp.Tupl).Els
	if len(args) > 0 {
		r = s.init
		for _, x := range args {
			e, err := p.Eval(c.Env, x)
			if err != nil {
				return nil, err
			}
			if !e.Zero() == s.neg {
				r = !s.init
				break
			}
		}
	}
	return lit.Bool(r), nil
}

var Err = &errSpec{impl("<form@err tupl?|exp @>")}

type errSpec struct{ exp.SpecBase }

func (s *errSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	return nil, ast.ErrUserErr(c.Src, c.String(), nil)
}

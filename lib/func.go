package lib

import (
	"fmt"
	"strconv"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Fn = &fnSpec{impl("<form@fn tupl?|tag|typ exp|@1 func@|@1>")}

type fnSpec struct{ exp.SpecBase }

func (s *fnSpec) Value() lit.Val { return s }
func (s *fnSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	fe := &FuncEnv{Par: env}
	tags, ok := c.Args[0].(*exp.Tupl)
	if ok && len(tags.Els) > 0 {
		explicitArgs(p, fe, tags.Els)
	}

	fe.mock = true
	x, err := p.Resl(fe, c.Args[1], typ.Void)
	fe.mock = false
	if err != nil {
		return c, err
	}

	ps := make([]typ.Param, 0, len(fe.Def)+1)
	for _, kl := range fe.Def {
		a := kl.Exp.(*exp.Lit)
		ps = append(ps, typ.P(kl.Tag, a.Res))
	}
	ps = append(ps, typ.P("", x.Resl()))

	ft := typ.Func(fmt.Sprintf("fn%d", p.NextFnID()), ps...)
	ft, err = p.Sys.Update(ft)
	if err != nil {
		return c, err
	}

	spec := makeFunc(fe, ft, x)
	if fe.rec {
		fe.recur = &recurSpec{exp.SpecBase{Decl: ft}, fe, fe.Def, x.Clone(), nil}
	}
	return &exp.Lit{Res: ft, Val: spec}, nil
}

func (s *fnSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	return nil, fmt.Errorf("unexpected fn eval %s", c)
}

func makeFunc(fe *FuncEnv, ft typ.Type, x exp.Exp) *funcSpec {
	return &funcSpec{SpecBase: exp.SpecBase{Decl: ft}, env: fe, act: x}
}

func explicitArgs(p *exp.Prog, fe *FuncEnv, es []exp.Exp) error {
	keys := make(exp.Tags, 0, len(es))
	for _, el := range es {
		tag := el.(*exp.Tag)
		pa, err := p.Eval(fe.Par, tag.Exp)
		if err != nil {
			return err
		}
		pv, ok := pa.Val.(typ.Type)
		if ok {
			t := *tag
			t.Exp = &exp.Lit{Res: pv, Src: tag.Src, Val: lit.Null{}}
			keys = append(keys, t)
			continue
		}
		return fmt.Errorf("expect type got %[1]T %[1]s", pa)
	}
	fe.expl = true
	fe.Def = keys
	return nil
}

type funcSpec struct {
	exp.SpecBase
	env *FuncEnv
	act exp.Exp
}

func (s *funcSpec) Value() lit.Val { return s }
func (s *funcSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
	_, err := s.SpecBase.Resl(p, env, c, h)
	if err != nil {
		return c, err
	}
	rp := exp.SigRes(c.Sig)
	s.act, err = p.Resl(s.env, s.act, rp.Type)
	if err != nil {
		return c, err
	}
	rp.Type, err = p.Sys.Update(rp.Type)
	return c, err
}

func (s *funcSpec) Eval(p *exp.Prog, c *exp.Call) (l *exp.Lit, err error) {
	for i, arg := range c.Args {
		// set arg vals in env
		kl := s.env.Def[i].Exp.(*exp.Lit)
		switch a := arg.(type) {
		case *exp.Tupl:
			switch len(a.Els) {
			case 0:
			case 1:
				l, err = p.Eval(c.Env, a.Els[0])
				if err != nil {
					return nil, err
				}
				kl.Val = l.Val
			default:
				return nil, fmt.Errorf("unexpected tupl")
			}
		default:
			l, err = p.Eval(c.Env, arg)
			if err != nil {
				return nil, err
			}
			kl.Val = l.Val
		}
	}
	return p.Eval(s.env, s.act)
}

type recurSpec struct {
	exp.SpecBase
	par  exp.Env
	def  exp.Tags
	act  exp.Exp
	spec *funcSpec
}

func (s *recurSpec) Value() lit.Val { return s }
func (s *recurSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	// we want to resolve the first layer of a recursion once
	if s.spec == nil {
		var n int // lets count up to two parent func envs
		for e := env; e != nil; e = e.Parent() {
			if _, ok := e.(*FuncEnv); !ok {
				continue
			}
			if n++; n > 1 {
				break
			}
		}
		s.spec = makeFunc(&FuncEnv{Par: s.par, Def: s.def}, s.Decl, s.act)
		if n < 2 { // only resolve the first recursion
			return s.spec.Resl(p, env, c, h)
		}
		// set the env otherwise so we can resolve on eval
		c.Env = env
	}
	return c, nil
}

func (s *recurSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	// we need to resolve second recursions, checking whether we are in the first
	// is more costly than re-resolving the first element.
	x, err := s.spec.Resl(p, c.Env, c, typ.Void)
	if err != nil {
		return nil, err
	}
	return s.spec.Eval(p, x.(*exp.Call))
}

type FuncEnv struct {
	Par   exp.Env
	Def   exp.Tags
	expl  bool
	mock  bool
	rec   bool
	recur *recurSpec
}

func (e *FuncEnv) Parent() exp.Env { return e.Par }
func (e *FuncEnv) Lookup(s *exp.Sym, k string, eval bool) (exp.Exp, error) {
	if k == "recur" {
		if e.mock {
			e.rec = true
			return s, nil
		}
		if e.recur != nil {
			// we want to copy the argument def when we recur
			// as not to reuse values from previous calls
			r := *e.recur
			r.act = e.recur.act.Clone()
			r.def = make(exp.Tags, len(e.Def))
			for i, kv := range e.Def {
				l := *(kv.Exp.(*exp.Lit))
				l.Val = lit.Null{}
				kv.Exp = &l
				r.def[i] = kv
			}
			return &exp.Lit{Res: r.Decl, Val: &r}, nil
		}
	}
	k, ok := dotkey(k)
	if !ok {
		return e.Par.Lookup(s, k, eval)
	}

	l, err := e.Def.Select(k)
	if eval {
		return l, err
	}
	if err != nil {
		if !e.mock || e.expl {
			return s, err
		}
		idx, kk := -1, k[1:]
		if b := kk[0]; b >= '0' && b <= '9' {
			i, err := strconv.Atoi(kk)
			if err == nil {
				idx = i
			}
		}
		p := exp.FindProg(e.Par)
		t := p.Sys.Bind(typ.Var(-1, typ.Void))
		l = &exp.Lit{Res: t, Val: lit.Null{}}
		if idx >= 0 {
			if idx >= len(e.Def) {
				for len(e.Def) <= idx {
					e.Def = append(e.Def, exp.Tag{})
				}
			}
			e.Def[idx].Exp = l
		} else {
			e.Def = append(e.Def, exp.Tag{Tag: kk, Exp: l})
		}
	}
	s.Type, s.Env, s.Rel = l.Res, e, k
	return s, nil
}
func dotkey(k string) (string, bool) {
	if k == "_" {
		k = ".0"
	}
	return exp.DotKey(k)
}

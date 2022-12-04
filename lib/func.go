package lib

import (
	"fmt"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Fn = &fnSpec{impl("<form@fn tupl?|tag|typ exp|@1 func@|@1>")}

type fnSpec struct{ exp.SpecBase }

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
	for _, kv := range fe.Def {
		ps = append(ps, typ.P(kv.Key, kv.Val.Type()))
	}
	ps = append(ps, typ.P("", typ.Res(x.Type())))

	ft := typ.Func(fmt.Sprintf("fn%d", p.NextFnID()), ps...)
	ft, err = p.Sys.Update(ft)
	if err != nil {
		return c, err
	}

	spec := makeFunc(fe, ft, x)
	if fe.rec {
		fe.recur = &recurSpec{exp.SpecBase{Decl: ft}, fe, fe.Def, x.Clone(), nil}
	}
	return exp.LitSrc(exp.NewSpecRef(spec), c.Src), nil
}

func (s *fnSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	return nil, fmt.Errorf("unexpected fn eval %s", c)
}

func makeFunc(fe *FuncEnv, ft typ.Type, x exp.Exp) *funcSpec {
	return &funcSpec{SpecBase: exp.SpecBase{Decl: ft}, env: fe, act: x}
}

func explicitArgs(p *exp.Prog, fe *FuncEnv, es []exp.Exp) (err error) {
	keys := make(lit.Keyed, 0, len(es))
	for _, el := range es {
		tag := el.(*exp.Tag)
		tag.Exp, err = p.Resl(fe.Par, tag.Exp, typ.Typ)
		if err != nil {
			return err
		}
		pa, err := p.Eval(fe.Par, tag.Exp)
		if err != nil {
			return err
		}
		pv, ok := pa.(typ.Type)
		if ok {
			keys = append(keys, lit.KeyVal{Key: tag.Tag, Val: lit.AnyWrap(pv)})
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

func (s *funcSpec) Eval(p *exp.Prog, c *exp.Call) (v lit.Val, err error) {
	for i, arg := range c.Args {
		// set arg vals in env
		kv := &s.env.Def[i]
		switch a := arg.(type) {
		case *exp.Tupl:
			switch len(a.Els) {
			case 0:
			case 1:
				v, err = p.Eval(c.Env, a.Els[0])
				if err != nil {
					return nil, err
				}
				kv.Val = v
			default:
				return nil, fmt.Errorf("unexpected tupl")
			}
		default:
			v, err = p.Eval(c.Env, arg)
			if err != nil {
				return nil, err
			}
			kv.Val = v
		}
	}
	return p.Eval(s.env, s.act)
}

type recurSpec struct {
	exp.SpecBase
	par  exp.Env
	def  lit.Keyed
	act  exp.Exp
	spec *funcSpec
}

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

func (s *recurSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
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
	Def   lit.Keyed
	expl  bool
	mock  bool
	rec   bool
	recur *recurSpec
}

func (e *FuncEnv) Parent() exp.Env { return e.Par }
func (e *FuncEnv) Lookup(s *exp.Sym, p cor.Path, eval bool) (lit.Val, error) {
	if p.Plain() == "recur" {
		if e.mock {
			e.rec = true
			return nil, nil
		}
		if e.recur != nil {
			// we want to copy the argument def when we recur
			// as not to reuse values from previous calls
			r := *e.recur
			r.act = e.recur.act.Clone()
			r.def = make(lit.Keyed, len(e.Def))
			for i, kv := range e.Def {
				kv.Val = lit.AnyWrap(kv.Type())
				r.def[i] = kv
			}
			return exp.NewSpecRef(&r), nil
		}
	}
	p, ok := dotkey(p)
	if !ok {
		return e.Par.Lookup(s, p, eval)
	}
	v, err := lit.SelectPath(&e.Def, p)
	if v == nil || err != nil {
		if eval || !e.mock || e.expl {
			return nil, err
		}
		fst, idx := p.Fst(), -1
		if fst.IsIdx() {
			idx = fst.Idx
		}
		t := exp.FindProg(e.Par).Sys.Bind(typ.Var(-1, typ.Void))
		v = lit.AnyWrap(t)
		if idx >= 0 {
			if idx >= len(e.Def) {
				for len(e.Def) <= idx {
					e.Def = append(e.Def, lit.KeyVal{})
				}
			}
			e.Def[idx].Val = v
		} else {
			e.Def = append(e.Def, lit.KeyVal{Key: fst.Key, Val: v})
		}
	}
	if s.Update(v.Type(), e, p); !eval {
		return nil, nil
	}
	return v, nil
}
func dotkey(p cor.Path) (cor.Path, bool) {
	if f := &p[0]; f.Sep() == 0 && f.Key == "_" {
		*f = cor.Seg{Sel: '.'}
	}
	return exp.DotPath(p)
}

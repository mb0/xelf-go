package lib

import (
	"fmt"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Fn = &fnSpec{impl("<form fn tupl?|tag|typ exp|@1 func@|@1>")}

type fnSpec struct{ exp.SpecBase }

func (s *fnSpec) Value() lit.Val { return s }
func (s *fnSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (_ exp.Exp, err error) {
	fe := &FuncEnv{Par: env, Def: &lit.Dict{}}
	var ft typ.Type
	tags, ok := c.Args[0].(*exp.Tupl)
	expl := ok && len(tags.Els) > 0
	if !expl {
		ft, err = implicitFn(p, fe, c, h)
	} else {
		ft, err = explicitFn(p, fe, c, tags.Els, h)
	}
	if err != nil {
		return c, err
	}
	x := c.Args[1]
	spec := makeFunc(fe, ft, x)
	if expl {
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

func explicitFn(p *exp.Prog, fe *FuncEnv, c *exp.Call, es []exp.Exp, h typ.Type) (typ.Type, error) {
	keys := make([]lit.KeyVal, 0, len(es))
	ps := make([]typ.Param, 0, len(es)+1)
	for _, el := range es {
		tag := el.(*exp.Tag)
		pa, err := p.Eval(fe.Par, tag.Exp)
		if err != nil {
			return typ.Void, err
		}
		if pa.Res.Kind&knd.Typ != 0 {
			pv, ok := pa.Val.(typ.Type)
			if ok {
				ps = append(ps, typ.P(tag.Tag, pv))
				keys = append(keys, lit.KeyVal{Key: tag.Tag, Val: &exp.Lit{Res: pv, Val: lit.Null{}}})
				continue
			}
		}
		return typ.Void, fmt.Errorf("expect type got %[1]T %[1]s", pa)
	}
	ps = append(ps, typ.P("", typ.Type{Kind: knd.Var}))
	fe.Def.Keyed = keys
	fn := fmt.Sprintf("fn%d", p.NextFnID())
	return p.Sys.Inst(typ.Func(fn, ps...))
}

func implicitFn(p *exp.Prog, fe *FuncEnv, c *exp.Call, h typ.Type) (typ.Type, error) {
	rt := p.Sys.Bind(typ.Var(0, typ.Void))
	fe.mock = true
	_, err := p.Resl(fe, c.Args[1], rt)
	fe.mock = false
	if err != nil {
		return typ.Void, err
	}
	ps := make([]typ.Param, 0, len(fe.Def.Keyed)+1)
	for _, kl := range fe.Def.Keyed {
		a := kl.Val.(*exp.Lit)
		ps = append(ps, typ.P(kl.Key, a.Res))
	}
	rt = p.Sys.Update(rt)
	ps = append(ps, typ.P("", rt))
	fn := fmt.Sprintf("fn%d", p.NextFnID())
	return typ.Func(fn, ps...), nil
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
	rp.Type = p.Sys.Update(rp.Type)
	return c, nil
}

func (s *funcSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	for i, arg := range c.Args {
		// set arg vals in env
		r, err := p.Eval(c.Env, arg)
		if err != nil {
			return nil, err
		}
		kl := &s.env.Def.Keyed[i]
		kl.Val.(*exp.Lit).Val = r.Val
	}
	return p.Eval(s.env, s.act)
}

type recurSpec struct {
	exp.SpecBase
	par  exp.Env
	def  *lit.Dict
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
	Def   *lit.Dict
	mock  bool
	recur *recurSpec
}

func (e *FuncEnv) Parent() exp.Env { return e.Par }
func (e *FuncEnv) Resl(p *exp.Prog, s *exp.Sym, k string) (exp.Exp, error) {
	if k == "recur" && e.recur != nil {
		// we want to copy the argument def when we recur
		// as not to reuse values from previous calls
		r := *e.recur
		r.act = e.recur.act.Clone()
		r.def = &lit.Dict{Keyed: make([]lit.KeyVal, len(e.Def.Keyed))}
		for i, kv := range e.Def.Keyed {
			l := kv.Val.(*exp.Lit)
			kv.Val = &exp.Lit{Res: l.Res, Val: lit.Null{}, Src: l.Src}
			r.def.Keyed[i] = kv
		}
		return &exp.Lit{Res: r.Type(), Val: &r}, nil
	}
	k, ok := dotkey(k)
	if !ok {
		return e.Par.Resl(p, s, k)
	}
	l, err := lit.Select(e.Def, k)
	if err != nil {
		kk := k[1:]
		if !e.mock {
			return s, nil
		}
		l = &exp.Lit{Res: p.Sys.Bind(typ.Var(0, typ.Void)), Val: lit.Null{}}
		e.Def.Keyed = append(e.Def.Keyed, lit.KeyVal{Key: kk, Val: l})
	}
	res, ok := l.(*exp.Lit)
	if ok {
		s.Type = res.Res
	} else {
		s.Type = l.Type()
	}
	s.Env, s.Rel = e, k
	return s, nil
}
func (e *FuncEnv) Eval(p *exp.Prog, s *exp.Sym, k string) (*exp.Lit, error) {
	k, ok := dotkey(k)
	if !ok {
		return e.Par.Eval(p, s, k)
	}
	l, err := lit.Select(e.Def, k)
	if err != nil {
		return nil, err
	}
	res, ok := l.(*exp.Lit)
	if !ok {
		res = &exp.Lit{Res: l.Type(), Val: l}
	}
	return res, nil
}
func dotkey(k string) (string, bool) {
	if k == "_" {
		k = ".0"
	} else if k[0] != '.' {
		return k, false
	} else {
		if len(k) > 1 && k[1] == '.' {
			return k[1:], false
		}
	}
	return k, true
}

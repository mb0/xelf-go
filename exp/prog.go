package exp

import (
	"context"
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// ErrDefer is a marker error used to indicate a deferred resolution and not a failure per-se.
// The user can errors.Is(err, ErrDefer) and resume program resolution with more context provided.
var ErrDefer = fmt.Errorf("deferred resolution")

// Prog is the entry context to resolve an expression in an environment.
// Programs are bound to their expression and cannot be reused.
type Prog struct {
	Ctx  context.Context
	Root Env
	Sys  *typ.Sys
	Reg  lit.Regs
	Arg  *Lit

	File File
	// Files collects all external files loaded by the program
	Files map[string]*File
	// Birth holds the uri for actively loading files to break recursive loads
	Birth map[string]struct{}

	fnid uint
	dyn  Spec
}

// NewProg returns a new program using the given registry, environment and expression.
// The registry argument can be nil, a new registry will be used by default.
func NewProg(env Env, args ...interface{}) *Prog {
	p := &Prog{Ctx: context.Background(), Root: env, Sys: typ.NewSys(), Reg: *lit.GlobalRegs()}
	for _, arg := range args {
		switch a := arg.(type) {
		case *lit.Regs:
			p.Reg = *lit.DefaultRegs(a)
		case context.Context:
			p.Ctx = a
		case lit.MutReg:
			p.Reg.MutReg = a
		case *lit.PrxReg:
			p.Reg.PrxReg = a
		}
	}
	if dyn, _ := LookupKey(env, "dyn"); dyn != nil {
		p.dyn, _ = dyn.Val.(Spec)
	}
	return p
}

// Run resolves and evaluates the input expression and returns the result or an error.
func (p *Prog) Run(x Exp, arg *Lit) (res *Lit, err error) {
	p.Arg = arg
	x, err = p.Resl(p, x, typ.Void)
	if err != nil {
		return nil, err
	}
	return p.Eval(p, x)
}

// RunStr resolves and evaluates the input string and returns the result or an error.
func (p *Prog) RunStr(str string, arg *Lit) (res *Lit, err error) {
	x, err := Parse(str)
	if err != nil {
		return nil, err
	}
	return p.Run(x, arg)
}

func FindProg(env Env) *Prog {
	for ; env != nil; env = env.Parent() {
		if p, _ := env.(*Prog); p != nil {
			return p
		}
	}
	return nil
}

func (p *Prog) Parent() Env { return p.Root }

func (p *Prog) Lookup(s *Sym, k string, eval bool) (res Exp, err error) {
	switch k[0] {
	case '$':
		if p.Arg == nil {
			break
		}
		l, err := SelectLookup(p.Arg, cor.Keyed(k[1:]), eval)
		if err != nil {
			return l, err
		}
		if s.Update(l.Res, p, k); !eval {
			return s, nil
		}
		return l, nil
	default:
		if qual, rest := SplitQualifier(k); qual != "" {
			if l, err := LookupMod(p, qual, rest); err == nil {
				s.Update(l.Res, p, k)
				return l, nil
			}
		}
	}
	res, err = p.Root.Lookup(s, k, eval)
	if err == ErrSymNotFound {
		if t, err := typ.ParseSym(k, s.Src, nil); err == nil {
			t, err = p.Sys.Inst(LookupType(s.Env), t)
			if err != nil {
				return nil, err
			}
			return LitSrc(t, s.Src), nil
		}
	}
	return res, err
}

// Resl resolves an expression using a type hint and returns the result or an error.
func (p *Prog) Resl(env Env, e Exp, h typ.Type) (Exp, error) {
	switch a := e.(type) {
	case *Tag:
		if a.Exp != nil {
			x, err := p.Resl(env, a.Exp, typ.Res(h))
			if err != nil {
				return nil, err
			}
			a.Exp = x
		}
		return a, nil
	case *Sym:
		if h.Kind == knd.Sym {
			return &Lit{Res: typ.Sym, Val: lit.Str(a.Sym), Src: a.Src}, nil
		}
		if a.Sym[0] == '@' {
			t, err := typ.ParseSym(a.Sym, a.Src, nil)
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, a.Sym, err)
			}
			t, err = p.Sys.Inst(LookupType(env), t)
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, a.Sym, err)
			}
			return LitSrc(t, a.Src), nil
		}
		if a.Env == nil {
			a.Update(a.Res, env, a.Sym)
		}
		r, err := a.Env.Lookup(a, a.Rel, false)
		if err != nil {
			return nil, ast.ErrReslSym(a.Src, a.Sym, err)
		}
		if h != typ.Void {
			ut, err := p.Sys.Unify(typ.Res(r.Type()), h)
			if err != nil {
				return nil, ast.ErrUnify(a.Src, err.Error())
			}
			a.Res = ut
		}
		return r, nil
	case *Lit:
		if a.Res == typ.Spec {
			t := a.Val.Type()
			nt, err := p.Sys.Inst(LookupType(env), t)
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, t, err)
			}
			a.Res = nt
			return a, nil
		}
		if a.Res == typ.VarTyp {
			t, ok := a.Val.(typ.Type)
			if !ok {
				return nil, ast.ErrReslTyp(a.Src, a.Val,
					fmt.Errorf("unexpected type value %T", a.Val),
				)
			}
			r, err := p.Sys.Inst(LookupType(env), t)
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, t, err)
			}
			a.Res = typ.Typ
			a.Val = r
		}
		rt, err := p.Sys.Unify(a.Res, h)
		if err != nil {
			return nil, ast.ErrUnify(a.Src, err.Error())
		}
		a.Res = rt
		return a, nil
	case *Tupl:
		tt, tn := typ.TuplEl(a.Res)
		for i, arg := range a.Els {
			ah := tt
			if tn > 1 {
				ah = tt.Body.(*typ.ParamBody).Params[i%tn].Type
			}
			el, err := p.Resl(env, arg, ah)
			if err != nil {
				return nil, err
			}
			a.Els[i] = el
		}
		_, err := p.Sys.Unify(a.Res, h)
		if err != nil {
			return nil, ast.ErrUnify(a.Src, err.Error())
		}
		return a, nil
	case *Call:
		if a.Spec == nil {
			if len(a.Args) == 0 {
				return nil, ast.ErrReslSpec(a.Src, "unexpected empty call", nil)
			}
			fst, err := p.Resl(env, a.Args[0], typ.Void)
			if err != nil {
				return nil, err
			}
			ft := fst.Type()
			if ft.Kind == knd.Lit && ft.Body != nil && ft.Body.(*typ.Type).Kind&knd.Spec != 0 {
				if l, ok := fst.(*Lit); ok {
					if s, ok := l.Val.(Spec); ok {
						a.Spec = s
						a.Args = a.Args[1:]
					}
				}
			}
			if a.Spec == nil {
				if p.dyn == nil {
					return nil, ast.ErrReslSpec(a.Src, "unsupported dyn call", nil)
				}
				a.Spec = p.dyn
				a.Args[0] = fst
			}
			a.Sig, err = p.Sys.Inst(LookupType(env), a.Spec.Type())
			if err != nil {
				return nil, ast.ErrReslSpec(a.Src, p.dyn.Type().String(), err)
			}
			a.Args, err = LayoutSpec(a.Sig, a.Args)
			if err != nil {
				return nil, ast.ErrLayout(a.Src, a.Sig, err)
			}
		}
		return a.Spec.Resl(p, env, a, h)
	}
	return nil, ast.ErrUnexpectedExp(e.Source(), e)
}

// Eval evaluates a resolved expression and returns a literal or an error.
func (p *Prog) Eval(env Env, e Exp) (_ *Lit, err error) {
	switch a := e.(type) {
	case *Sym:
		res, err := env.Lookup(a, a.Sym, true)
		if err != nil {
			return nil, ast.ErrEval(a.Src, a.Sym, err)
		}
		if l, ok := res.(*Lit); ok {
			return l, nil
		}
		return nil, fmt.Errorf("runtime env %T eval did return %T result", env, res)
	case *Call:
		res, err := a.Spec.Eval(p, a)
		if err != nil {
			return nil, ast.ErrEval(a.Src, a.Sig.Ref, err)
		}
		return res, nil
	case *Tupl:
		vals := make([]lit.Val, len(a.Els))
		for i, arg := range a.Els {
			at, err := p.Eval(env, arg)
			if err != nil {
				return nil, err
			}
			vals[i] = at.Val
		}
		return &Lit{Val: &lit.List{Vals: vals}}, nil
	case *Lit:
		if a.Res == typ.Spec {
			t := a.Val.Type()
			nt, err := p.Sys.Inst(LookupType(env), t)
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, t, err)
			}
			a.Res = nt
			return a, nil
		}
		if a.Res == typ.Typ {
			if t, ok := a.Val.(typ.Type); ok {
				a.Val, err = p.Sys.Update(t)
				if err != nil {
					return nil, ast.ErrEval(a.Src, t.String(), err)
				}
			}
		}
		return a, nil
	}
	return nil, ast.ErrUnexpectedExp(e.Source(), e)
}

// EvalArgs evaluates resolved call arguments and returns the result or an error.
// This is a convenience method for the most basic needs of many spec implementations.
func (p *Prog) EvalArgs(c *Call) ([]*Lit, error) {
	res := make([]*Lit, len(c.Args))
	for i, arg := range c.Args {
		if arg == nil {
			continue
		}
		a, err := p.Eval(c.Env, arg)
		if err != nil {
			return nil, err
		}
		res[i] = a
	}
	return res, nil
}

// NextFnID returns a new number to identify an anonymous function.
func (p *Prog) NextFnID() uint {
	p.fnid++
	return p.fnid
}

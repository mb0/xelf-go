package exp

import (
	"context"
	"fmt"
	"strings"

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
	Arg  lit.Val

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
		p.dyn, _ = dyn.Value().(Spec)
	}
	return p
}

// Run resolves and evaluates the input expression and returns the result or an error.
func (p *Prog) Run(x Exp, arg lit.Val) (_ lit.Val, err error) {
	p.Arg = arg
	x, err = p.Resl(p, x, typ.Void)
	if err != nil {
		return nil, err
	}
	return p.Eval(p, x)
}

// RunStr resolves and evaluates the input string and returns the result or an error.
func (p *Prog) RunStr(str string, arg lit.Val) (lit.Val, error) {
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

func (p *Prog) Lookup(s *Sym, pp cor.Path, eval bool) (res lit.Val, err error) {
	if len(pp) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	if fst := &pp[0]; fst.Sep() == 0 {
		if fst.Key != "" && fst.Key[0] == '$' {
			if p.Arg != nil {
				org := fst.Key
				fst.Key = org[1:]
				v, err := SelectLookup(p.Arg, pp, eval)
				if err == nil && v != nil {
					fst.Key = org
					s.Update(v.Type(), p, pp)
					if !eval && v.Nil() {
						return nil, nil
					}
					return v, nil
				}
				return nil, ErrSymNotFound
			}
		} else if len(pp) > 1 && cor.IsKey(fst.Key) && strings.HasPrefix(s.Sym, fst.Key) {
			if v, err := LookupMod(p, fst.Key, pp[1:]); err == nil {
				s.Update(v.Type(), p, pp)
				return v, nil
			}
		}
	}
	res, err = p.Root.Lookup(s, pp, eval)
	if err == ErrSymNotFound {
		if t, err := typ.ParseSym(s.Sym, s.Src, nil); err == nil {
			t, err = p.Sys.Inst(LookupType(s.Env), t)
			if err != nil {
				return nil, err
			}
			s.Update(t.Type(), p, pp)
			return t, nil
		}
		return nil, err
	} else if err != nil || res == nil {
		return nil, err
	}
	s.Update(res.Type(), p, pp)
	return res, nil
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
			return LitSrc(lit.Wrap(lit.Str(a.Sym).Mut(), typ.Sym), a.Src), nil
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
			path, err := cor.ParsePath(a.Sym)
			if err != nil {
				return nil, ast.ErrReslSym(a.Src, a.Sym, err)
			}
			a.Update(a.Res, env, path)
		}
		r, err := a.Env.Lookup(a, a.Path, false)
		if err != nil {
			return nil, ast.ErrReslSym(a.Src, a.Sym, err)
		}
		ut, err := p.Sys.Unify(a.Res, h)
		if err != nil {
			return nil, ast.ErrUnify(a.Src, err.Error())
		}
		a.Res = ut
		if r != nil {
			return LitSrc(r, a.Src), nil
		}
		return a, nil
	case *Lit:
		t := typ.Res(a.Type())
		if t == typ.Spec {
			sp := UnwrapSpec(lit.Unwrap(a.Val))
			nt, err := p.Sys.Inst(LookupType(env), sp.Decl)
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, t, err)
			}
			sp.Decl = nt
			return a, nil
		}
		if t == typ.VarTyp {
			r, err := p.Sys.Inst(LookupType(env), a.Value().(typ.Type))
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, t, err)
			}
			a.Val = r
			_, err = p.Sys.Unify(r, h)
			if err != nil {
				return nil, ast.ErrUnify(a.Src, err.Error())
			}
			return a, nil
		}
		_, err := p.Sys.Unify(t, h)
		if err != nil {
			return nil, ast.ErrUnify(a.Src, err.Error())
		}
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
					if s, ok := l.Value().(Spec); ok {
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

// Eval evaluates a resolved expression and returns a value or an error.
func (p *Prog) Eval(env Env, e Exp) (_ lit.Val, err error) {
	switch a := e.(type) {
	case *Sym:
		if a.Env == nil {
			e, err = p.Resl(env, a, typ.Void)
			if err != nil {
				return nil, err
			}
			return p.Eval(env, e)
		}
		res, err := a.Env.Lookup(a, a.Path, true)
		if err != nil {
			return nil, ast.ErrEval(a.Src, a.Sym, err)
		}
		return res, nil
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
			vals[i] = at
		}
		return &lit.List{Typ: typ.List, Vals: vals}, nil
	case *Lit:
		t := typ.Res(a.Type())
		if t == typ.Spec {
			s := UnwrapSpec(a.Value())
			nt, err := p.Sys.Inst(LookupType(env), s.Decl)
			if err != nil {
				return nil, ast.ErrReslTyp(a.Src, t, err)
			}
			s.Decl = nt
			return s, nil
		}
		if t.Kind&knd.All == knd.Typ {
			t, err = p.Sys.Update(t)
			if err != nil {
				return nil, ast.ErrEval(a.Src, t.String(), err)
			}
			return typ.El(t), nil
		}
		return a.Val, nil
	}
	return nil, ast.ErrUnexpectedExp(e.Source(), e)
}

// EvalArgs evaluates resolved call arguments and returns the result or an error.
// This is a convenience method for the most basic needs of many spec implementations.
func (p *Prog) EvalArgs(c *Call) (lit.Vals, error) {
	res := make(lit.Vals, len(c.Args))
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

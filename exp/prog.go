package exp

import (
	"context"
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// ErrDefer is a marker error used to indicate a deferred resolution and not a failure per-se.
// The user can errors.Is(err, ErrDefer) and resume program resolution with more context provided.
var ErrDefer = fmt.Errorf("deferred resolution")

// Eval creates and evaluates a new program for str and returns the result or an error.
func Eval(ctx context.Context, reg *lit.Reg, env Env, str string) (*Lit, error) {
	if reg == nil {
		reg = &lit.Reg{}
	}
	x, err := Parse(reg, str)
	if err != nil {
		return nil, err
	}
	return EvalExp(ctx, reg, env, x)
}

// EvalExp creates and evaluates a new program for x and returns the result or an error.
func EvalExp(ctx context.Context, reg *lit.Reg, env Env, x Exp) (*Lit, error) {
	p := NewProg(ctx, reg, env, x)
	x, err := p.Resl(env, x, typ.Void)
	if err != nil {
		return nil, err
	}
	return p.Eval(env, x)
}

// Env is a scoped context to resolve symbols. Envs configure most of the program resolution.
type Env interface {
	// Parent returns the parent environment or nil.
	Parent() Env

	// Dyn returns a dyn spec for this environment or nil.
	Dyn() Spec

	// Resl resolves a part of a symbol and returns the result or an error.
	// If eval is true we expect a *exp.Lit result or an error.
	Resl(p *Prog, s *Sym, k string, eval bool) (Exp, error)
}

// Prog is the entry context to resolve an expression in an environment.
// Programs are bound to their expression and cannot be reused.
type Prog struct {
	Ctx  context.Context
	Reg  *lit.Reg
	Sys  *typ.Sys
	Root Env
	Exp  Exp
	fnid uint
}

// NewProg returns a new program using the given registry, environment and expression.
// The registry argument can be nil, a new registry will be used by default.
func NewProg(ctx context.Context, reg *lit.Reg, env Env, exp Exp) *Prog {
	if reg == nil {
		reg = &lit.Reg{}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return &Prog{Ctx: ctx, Reg: reg, Sys: typ.NewSys(reg), Root: env, Exp: exp}
}

// Resl resolves an expression using a type hint and returns the result or an error.
func (p *Prog) Resl(env Env, e Exp, h typ.Type) (Exp, error) {
	switch a := e.(type) {
	case *Tag:
		if a.Exp != nil {
			x, err := p.Resl(env, a.Exp, typ.ResEl(h))
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
		k := a.Sym
		if a.Env != nil {
			env = a.Env
			k = a.Rel
		}
		r, err := env.Resl(p, a, k, false)
		if err != nil {
			return nil, ast.ErrReslSym(a.Src, a.Sym, err)
		}
		// TODO check hint
		return r, nil
	case *Lit:
		if a.Res.Kind&knd.Typ != 0 {
			t, ok := a.Val.(typ.Type)
			if ok {
				a.Val = p.Sys.Update(t)
			}
		}
		rt, err := p.Sys.Unify(a.Res, h)
		if err != nil {
			return nil, ast.ErrUnify(a.Src, err.Error())
		}
		a.Res = rt
		return a, nil
	case *Tupl:
		tt, tn := typ.TuplEl(a.Type)
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
		ut, err := p.Sys.Unify(a.Type, h)
		if err != nil {
			return nil, ast.ErrUnify(a.Src, err.Error())
		}
		a.Type = ut
		return a, nil
	case *Call:
		if a.Spec == nil {
			spec, args, err := p.reslSpec(env, a)
			if err != nil {
				return nil, err
			}
			sig, err := p.Sys.Inst(spec.Type())
			if err != nil {
				return nil, ast.ErrLayout(a.Src, sig, err)
			}
			args, err = LayoutSpec(sig, args)
			if err != nil {
				return nil, ast.ErrLayout(a.Src, sig, err)
			}
			a.Sig, a.Spec, a.Args = sig, spec, args
		}
		return a.Spec.Resl(p, env, a, h)
	}
	return nil, ast.ErrUnexpectedExp(e.Source(), e)
}

// Eval evaluates a resolved expression and returns a literal or an error.
func (p *Prog) Eval(env Env, e Exp) (_ *Lit, err error) {
	switch a := e.(type) {
	case *Sym:
		res, err := env.Resl(p, a, a.Sym, true)
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
		if a.Res.Kind&knd.Typ != 0 {
			if t, ok := a.Val.(typ.Type); ok {
				a.Val = p.Sys.Update(t)
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

// NextFnID returns a new number to identify an anonymous functions.
func (p *Prog) NextFnID() uint {
	p.fnid++
	return p.fnid
}

func (p *Prog) reslSpec(env Env, c *Call) (Spec, []Exp, error) {
	if len(c.Args) == 0 {
		return nil, nil, ast.ErrReslSpec(c.Src, "unexpected empty call", nil)
	}
	fst, err := p.Resl(env, c.Args[0], typ.Void)
	if err != nil {
		return nil, nil, err
	}
	if fst.Kind() == knd.Lit && fst.Resl().Kind&knd.Spec != 0 {
		if l, ok := fst.(*Lit); ok {
			if s, ok := l.Val.(Spec); ok {
				return s, c.Args[1:], nil
			}
		}
	}
	dyn := env.Dyn()
	if dyn == nil {
		name := fmt.Sprintf("no dyn spec found for %s", fst)
		return nil, nil, ast.ErrReslSpec(c.Src, name, nil)
	}
	return dyn, c.Args, nil
}

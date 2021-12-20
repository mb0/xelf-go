package exp

import (
	"fmt"

	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// ErrSymNotFound is an error that indicates that a symbol was not found in the environment
var ErrSymNotFound = fmt.Errorf("sym not found")

// Builtins is a root environment to resolve symbols to builtin specs and at last as types.
type Builtins map[string]Spec

func (e Builtins) Parent() Env { return nil }
func (e Builtins) Dyn() Spec   { return e["dyn"] }

func (e Builtins) Resl(p *Prog, s *Sym, k string) (Exp, error) { return e.Eval(p, s, k) }
func (e Builtins) Eval(p *Prog, s *Sym, k string) (*Lit, error) {
	if sp := e[k]; sp != nil {
		return &Lit{Res: sp.Type(), Val: sp, Src: s.Src}, nil
	}
	t, err := typ.ParseSym(k, s.Src, nil)
	if err == nil {
		t, err = p.Sys.Inst(t)
		if err != nil {
			return nil, err
		}
		return &Lit{Res: typ.Typ, Val: t, Src: s.Src}, nil
	}
	return nil, ErrSymNotFound
}

// ArgEnv is a child scope that supports special paths starting with '$' into a literal value.
type ArgEnv struct {
	Par Env
	Typ typ.Type
	Val lit.Val
}

func (e *ArgEnv) Parent() Env { return e.Par }
func (e *ArgEnv) Dyn() Spec   { return e.Par.Dyn() }

func (e *ArgEnv) Resl(p *Prog, s *Sym, k string) (Exp, error) {
	if k[0] != '$' {
		return e.Par.Resl(p, s, k)
	}
	return e.Eval(p, s, k)
}
func (e *ArgEnv) Eval(p *Prog, s *Sym, k string) (*Lit, error) {
	if k[0] != '$' {
		return e.Par.Eval(p, s, k)
	}
	res, err := lit.Select(e.Val, k[1:])
	if err != nil {
		return nil, err
	}
	// TODO introduce exp Select to sort out correct literal type
	// this is a stop-gap hack only for arg env
	return &Lit{Res: typ.Abstract(res.Type()), Val: res, Src: s.Src}, nil
}

// DotEnv is a child scope that supports relative paths into a literal.
type DotEnv struct {
	Par Env
	Dot *Lit
}

func (e *DotEnv) Parent() Env { return e.Par }
func (e *DotEnv) Dyn() Spec   { return e.Par.Dyn() }

func (e *DotEnv) Resl(p *Prog, s *Sym, k string) (Exp, error) {
	k, ok := DotKey(k)
	if !ok {
		return e.Par.Resl(p, s, k)
	}
	t, err := typ.Select(e.Dot.Res, k)
	if err != nil {
		return nil, err
	}
	s.Type, s.Env, s.Rel = t, e, k
	return s, nil
}
func (e *DotEnv) Eval(p *Prog, s *Sym, k string) (*Lit, error) {
	k, ok := DotKey(k)
	if !ok {
		return e.Par.Eval(p, s, k)
	}
	v, err := lit.Select(e.Dot.Val, k)
	if err != nil {
		return nil, err
	}
	return &Lit{Res: v.Type(), Val: v, Src: s.Src}, nil
}

// DotKey returns whether k is a dot key or otherwise returns k with a leading dot removed.
func DotKey(k string) (string, bool) {
	if k[0] != '.' {
		return k, false
	}
	if len(k) > 1 && k[1] == '.' {
		return k[1:], false
	}
	return k, true
}

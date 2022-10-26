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

func (e Builtins) Lookup(s *Sym, k string, eval bool) (Exp, error) {
	if sp := e[k]; sp != nil {
		return &Lit{Res: sp.Type(), Val: sp, Src: s.Src}, nil
	}
	return nil, ErrSymNotFound
}

// ArgEnv is a child scope that supports special paths starting with '$' into a literal value.
type ArgEnv struct {
	Par Env
	Typ typ.Type
	Val lit.Val
}

func NewArgEnv(par Env, v lit.Val) *ArgEnv { return &ArgEnv{Par: par, Typ: v.Type(), Val: v} }

func (e *ArgEnv) Parent() Env { return e.Par }

func (e *ArgEnv) Lookup(s *Sym, k string, eval bool) (Exp, error) {
	if k[0] != '$' {
		return e.Par.Lookup(s, k, eval)
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

func (e *DotEnv) Lookup(s *Sym, k string, eval bool) (Exp, error) {
	k, ok := DotKey(k)
	if !ok {
		return e.Par.Lookup(s, k, eval)
	}
	if !eval {
		t, err := typ.Select(e.Dot.Res, k)
		if err != nil {
			return nil, err
		}
		s.Type, s.Env, s.Rel = t, e, k
		return s, nil
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

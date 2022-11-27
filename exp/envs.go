package exp

import (
	"fmt"

	"xelf.org/xelf/typ"
)

// ErrSymNotFound is an error that indicates that a symbol was not found in the environment
var ErrSymNotFound = fmt.Errorf("sym not found")

// Env is a scoped context to resolve symbols. Envs configure most of the program resolution.
type Env interface {
	// Parent returns the parent environment or nil.
	Parent() Env

	// Lookup resolves a part of a symbol and returns the result or an error.
	// If eval is true we expect a *exp.Lit result or an error.
	Lookup(s *Sym, k string, eval bool) (Exp, error)
}

// Builtins is a root environment to resolve symbols to builtin specs and at last as types.
type Builtins map[string]Spec

func (e Builtins) Parent() Env { return nil }

func (e Builtins) Lookup(s *Sym, k string, eval bool) (Exp, error) {
	if sp := e[k]; sp != nil {
		return &Lit{Res: typ.Spec, Val: sp, Src: s.Src}, nil
	}
	return nil, ErrSymNotFound
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
	l, err := SelectLookup(e.Dot, k, eval)
	if err != nil {
		return nil, err
	}
	if s.Update(l.Res, e, k); !eval {
		return s, nil
	}
	return l, nil
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

func LookupType(env Env) typ.Lookup {
	return func(k string) (_ typ.Type, err error) {
		l, err := LookupKey(env, k)
		if err != nil {
			return typ.Void, err
		}
		if t, ok := l.Val.(typ.Type); ok {
			return t, nil
		}
		return l.Res, nil
	}
}

func LookupKey(env Env, k string) (*Lit, error) {
	x, err := env.Lookup(&Sym{Sym: k, Env: env, Rel: k}, k, true)
	if err != nil {
		return nil, err
	}
	l, _ := x.(*Lit)
	if l == nil {
		return nil, ErrSymNotFound
	}
	return l, nil
}

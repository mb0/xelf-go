package exp

import (
	"fmt"

	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// ErrSymNotFound is an error that indicates that a symbol was not found in the environment
var ErrSymNotFound = fmt.Errorf("sym not found")

// Env is a scoped context to resolve symbols. Envs configure most of the program resolution.
type Env interface {
	// Parent returns the parent environment or nil.
	Parent() Env

	// Lookup resolves a part of a symbol and returns the result, an error or nothing.
	// We always update the symbol and return resolved values.
	// If eval is false the symbol resolves, but the value does not, we return nothing.
	// If the value is not resolved and eval is true we return an error.
	Lookup(s *Sym, k string, eval bool) (lit.Val, error)
}

// Builtins is a root environment to resolve symbols to builtin specs and at last as types.
type Builtins map[string]Spec

func (e Builtins) Parent() Env { return nil }

func (e Builtins) Lookup(s *Sym, k string, eval bool) (lit.Val, error) {
	if sp := e[k]; sp != nil {
		s.Update(typ.Spec, e, k)
		return lit.Wrap(NewSpecRef(sp), typ.Spec), nil
	}
	return nil, ErrSymNotFound
}

// DotEnv is a child scope that supports relative paths into a literal.
type DotEnv struct {
	Par Env
	Dot lit.Val
}

func (e *DotEnv) Parent() Env { return e.Par }

func (e *DotEnv) Lookup(s *Sym, k string, eval bool) (lit.Val, error) {
	k, ok := DotKey(k)
	if !ok {
		return e.Par.Lookup(s, k, eval)
	}
	v, err := SelectLookup(e.Dot, k, eval)
	if err != nil {
		return nil, err
	}
	s.Update(typ.Res(v.Type()), e, k)
	return v, nil
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
		// TODO we need to pass in the sym to determine the resolving env
		v, err := LookupKey(env, k)
		if err != nil {
			return typ.Void, err
		}
		t, ok := v.(typ.Type)
		if !ok {
			t = typ.Res(v.Type())
		}
		// TODO check if env is prog or is a root root env otherwise clear names
		// TODO check, restrict and edit type names if from mod env
		return t, nil
	}
}

func LookupKey(env Env, k string) (lit.Val, error) {
	return env.Lookup(&Sym{Sym: k, Env: env, Rel: k}, k, true)
}

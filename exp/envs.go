package exp

import (
	"fmt"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// ErrSymNotFound is an error that indicates that a symbol was not found in the environment
var ErrSymNotFound = fmt.Errorf("sym not found")

// Env is a scoped context to resolve symbols. Envs configure most of the program resolution.
type Env interface {
	// Parent returns the parent environment or nil.
	Parent() Env

	// Lookup resolves a symbol path and returns the result, an error or nothing.
	// We always update the symbol and return resolved values.
	// If eval is false the symbol resolves, but the value does not, we return nothing.
	// If the value is not resolved and eval is true we return an error.
	Lookup(s *Sym, path cor.Path, eval bool) (lit.Val, error)
}

// Builtins is a root environment to resolve symbols to builtin specs and at last as types.
type Builtins map[string]Spec

func (e Builtins) Parent() Env { return nil }

func (e Builtins) Lookup(s *Sym, p cor.Path, eval bool) (lit.Val, error) {
	if sp := e[p[0].Key]; sp != nil {
		s.Update(typ.Spec, e, p)
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

func (e *DotEnv) Lookup(s *Sym, p cor.Path, eval bool) (lit.Val, error) {
	p, ok := DotPath(p)
	if !ok {
		return e.Par.Lookup(s, p, eval)
	}
	v, err := SelectLookup(e.Dot, p, eval)
	if err != nil || v == nil {
		return nil, ErrSymNotFound
	}
	if s.Update(typ.Res(v.Type()), e, p); !eval && v.Nil() {
		return nil, nil
	}
	return v, nil
}

// DotPath returns whether p is a dot path or returns p with a leading dot segment removed.
func DotPath(p cor.Path) (cor.Path, bool) {
	fst := p[0]
	if fst.Sep() != '.' {
		return p, false
	}
	if fst.Empty() && len(p) > 1 {
		return p[1:], false
	}
	return p, true
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
	p, err := cor.ParsePath(k)
	if err != nil {
		return nil, err
	}
	return env.Lookup(&Sym{Sym: k, Env: env, Path: p}, p, true)
}

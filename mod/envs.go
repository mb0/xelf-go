package mod

import (
	"net/url"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

// LoaderEnv adds module awareness to a program environment.
// It provides the mod and use forms and holds the module loaders.
type LoaderEnv struct {
	Par     exp.Env
	Loaders []Loader
}

// NewLoaderEnv create a new module loader environment with the given parent env and loader.
// The parent env should be used as basis for external module loads.
func NewLoaderEnv(par exp.Env, ls ...Loader) *LoaderEnv {
	return &LoaderEnv{Par: par, Loaders: ls}
}

func FindLoaderEnv(env exp.Env) *LoaderEnv {
	for ; env != nil; env = env.Parent() {
		if le, _ := env.(*LoaderEnv); le != nil {
			return le
		}
	}
	return nil
}

func (le *LoaderEnv) LoadFile(p *exp.Prog, raw string) (f *File, err error) {
	url, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	for _, l := range le.Loaders {
		f, err = l.LoadFile(p, url)
		if err != nil {
			if err == ErrFileNotFound {
				continue
			}
			return nil, err
		}
		for _, m := range f.Decls {
			if m.Setup != nil {
				err := m.Setup(p, m.Mod)
				if err != nil {
					return nil, err
				}
			}
		}
		return f, err
	}
	return nil, ErrFileNotFound
}

func (le *LoaderEnv) Parent() exp.Env { return le.Par }

func (le *LoaderEnv) Lookup(s *exp.Sym, k string, eval bool) (exp.Exp, error) {
	var spec lit.Val
	// we return the mod and use spec only here so we can expect a loader env in those specs
	switch k {
	case "mod":
		spec = ModSpec
	case "use":
		spec = Use
	case "export":
		spec = Export
	default:
		return le.Par.Lookup(s, k, eval)
	}
	return &exp.Lit{Res: spec.Type(), Val: spec, Src: s.Src}, nil
}

// ModEnv encapsulates a module environment.
type ModEnv struct {
	Par exp.Env
	Mod *Mod
}

func (e *ModEnv) Parent() exp.Env { return e.Par }

func (e *ModEnv) Lookup(s *exp.Sym, k string, eval bool) (exp.Exp, error) {
	l, err := exp.Select(e.Mod.Res, k)
	if err != nil {
		return e.Par.Lookup(s, k, eval)
	}
	return l, nil
}

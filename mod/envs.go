package mod

import (
	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
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
func (le *LoaderEnv) LoadFile(prog *exp.Prog, loc *Loc) (f *File, err error) {
	base := ParseLoc(prog.File.URL)
	for _, l := range le.Loaders {
		src, err := l.LoadSrc(loc, base)
		if err != nil {
			if err == ErrFileNotFound {
				continue
			}
			return nil, err
		}
		if prog.Files == nil {
			prog.Files = make(map[string]*File)
		} else if f = prog.Files[src.URL]; f != nil {
			return f, nil
		}
		if src.Setup != nil {
			f, err = src.Setup(prog, src)
		} else {
			e, err := exp.ParseAll(src.Raw)
			if err != nil {
				return nil, err
			}
			// shallow copy the loader for every loaded file
			p := *prog
			p.File = File{URL: src.URL}
			e, err = p.Resl(&p, e, typ.Void)
			if err != nil {
				return nil, err
			}
			_, err = p.Eval(&p, e)
			f = &p.File
		}
		prog.Files[src.URL] = f
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
	obj *lit.Obj
}

func FindModEnv(env exp.Env) *ModEnv {
	for ; env != nil; env = env.Parent() {
		if me, _ := env.(*ModEnv); me != nil {
			return me
		}
	}
	return nil
}
func NewModEnv(par exp.Env, file *File, src ast.Src) *ModEnv {
	obj := &lit.Obj{Typ: typ.Type{Kind: knd.Mod | knd.Obj, Body: &typ.ParamBody{}}}
	m := &Mod{File: file, Decl: &exp.Lit{Res: obj.Typ, Val: obj, Src: src}}
	file.Refs = append(file.Refs, exp.ModRef{Pub: true, Mod: m})
	return &ModEnv{Par: par, Mod: m, obj: obj}
}

func (e *ModEnv) Parent() exp.Env { return e.Par }

func (e *ModEnv) SetName(name string) {
	if o := e.obj; o != nil {
		e.Mod.Name = name
		o.Typ.Ref = name
		e.Mod.Decl.Res = e.obj.Typ
	}
}
func (e *ModEnv) AddDecl(name string, v lit.Val) {
	if o := e.obj; o != nil {
		pb := o.Typ.Body.(*typ.ParamBody)
		pb.Params = append(pb.Params, typ.P(name, v.Type()))
		o.Vals = append(o.Vals, v)
	}
}

func (e *ModEnv) Lookup(s *exp.Sym, k string, eval bool) (exp.Exp, error) {
	if o := e.obj; o != nil {
		v, err := lit.Select(o, k)
		if err == nil {
			return exp.LitVal(v), nil
		}
	}
	return e.Par.Lookup(s, k, eval)
}

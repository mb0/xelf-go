package mod

import (
	"fmt"
	"strings"

	"xelf.org/xelf/cor"
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
	var src *Src
	for _, l := range le.Loaders {
		src, err = l.LoadSrc(loc, base)
		if err != nil {
			if err == ErrFileNotFound {
				continue
			}
			return nil, fmt.Errorf("module source load error for %s:\n%v", loc.URL, err)
		}
		if prog.Files == nil {
			prog.Files = make(map[string]*File)
		} else if f = prog.Files[src.URL]; f != nil {
			return f, nil
		}
		if prog.Birth == nil {
			prog.Birth = make(map[string]struct{})
		} else if _, ok := prog.Birth[src.URL]; ok {
			return nil, fmt.Errorf("module load recursion detected for %s", src.URL)
		}
		prog.Birth[src.URL] = struct{}{}
		if src.Setup != nil {
			f, err = src.Setup(prog, src)
		} else {
			var e exp.Exp
			e, err = exp.ParseAll(src.Raw)
			if err != nil {
				break
			}
			// shallow copy the loader for every loaded file
			p := *prog
			p.File = File{URL: src.URL}
			e, err = p.Resl(&p, e, typ.Void)
			if err != nil {
				break
			}
			_, err = p.Eval(&p, e)
			f = &p.File
		}
		if err != nil {
			break
		}
		delete(prog.Birth, src.URL)
		prog.Files[src.URL] = f
		return f, nil
	}
	if err == nil {
		err = ErrFileNotFound
	}
	return nil, fmt.Errorf("module load failed for %s:\n%v", loc, err)
}

func (le *LoaderEnv) Parent() exp.Env { return le.Par }

func (le *LoaderEnv) Lookup(s *exp.Sym, p cor.Path, eval bool) (lit.Val, error) {
	// we return the mod and use spec only here so we can expect a loader env in those specs
	switch p.Plain() {
	case "module":
		return exp.NewSpecRef(Module), nil
	case "import":
		return exp.NewSpecRef(Import), nil
	case "export":
		return exp.NewSpecRef(Export), nil
	}
	return le.Par.Lookup(s, p, eval)
}

// ModEnv encapsulates a module environment.
type ModEnv struct {
	Par exp.Env
	Mod *Mod
}

func FindModEnv(env exp.Env) *ModEnv {
	for ; env != nil; env = env.Parent() {
		if me, _ := env.(*ModEnv); me != nil {
			return me
		}
	}
	return nil
}
func NewModEnv(par exp.Env, file *File) *ModEnv {
	return &ModEnv{Par: par, Mod: &Mod{File: file, Decl: &lit.Obj{
		Typ: typ.Type{Kind: knd.Mod | knd.Obj, Body: &typ.ParamBody{}},
	}}}
}

func (e *ModEnv) Parent() exp.Env { return e.Par }

func (e *ModEnv) SetName(name string) {
	if m := e.Mod; m != nil {
		e.Mod.Name = name
		e.Mod.Decl.Typ.Ref = name
	}
}
func (e *ModEnv) AddDecl(name string, v lit.Val) error {
	if m := e.Mod; m != nil {
		if name == "" || !cor.IsName(name) {
			return fmt.Errorf("invalid module declaration name %q", name)
		}
		p := typ.P(name, v.Type())
		pb := m.Decl.Typ.Body.(*typ.ParamBody)
		for _, o := range pb.Params {
			if p.Key == o.Key {
				return fmt.Errorf("module declaration name %q is not unique", name)
			}
		}
		pb.Params = append(pb.Params, p)
		m.Decl.Vals = append(m.Decl.Vals, v)
	}
	return nil
}

// Publish checks and publishes the module to the file or returns an error.
// It must be called once at the end of module setup or evaluation.
func (e *ModEnv) Publish() error {
	if m := e.Mod; m != nil {
		if m.Name == "" || !cor.IsKey(m.Name) {
			return fmt.Errorf("invalid module name %q", m.Name)
		}
		return m.File.AddRefs(exp.ModRef{Pub: true, Mod: m})
	}
	return nil
}

func (e *ModEnv) Lookup(s *exp.Sym, path cor.Path, eval bool) (lit.Val, error) {
	if m := e.Mod; m != nil {
		p := path
		if len(p) > 1 && p[0].Key == m.Name && strings.HasPrefix(s.Sym, m.Name) {
			p = p[1:]
		}
		v, err := lit.SelectPath(m.Decl, p)
		if err == nil && v != nil {
			return v, nil
		}
		if len(p) != len(path) {
			return nil, exp.ErrSymNotFound
		}
	}
	return e.Par.Lookup(s, path, eval)
}

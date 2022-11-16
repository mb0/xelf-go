package mod

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var ModSpec = &modSpec{impl("<form@mod name:sym tags?:tupl|tag void>")}

type modSpec struct{ exp.SpecBase }

func (s *modSpec) Value() lit.Val { return s }
func (s *modSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, _ typ.Type) (_ exp.Exp, err error) {
	if c.Env != nil {
		return c, nil
	}
	name := c.Args[0].String()
	m := &Mod{File: &p.File, Name: name}
	c.Env = &ModEnv{Par: env, Mod: m}
	// eval elements to build the result type and value
	tags := c.Args[1].(*exp.Tupl)
	// create module type
	pb := &typ.ParamBody{Params: make([]typ.Param, 0, len(tags.Els))}
	val := &lit.Obj{Vals: make([]lit.Val, 0, len(tags.Els)),
		Typ: typ.Type{Kind: knd.Mod | knd.Obj, Ref: name, Body: pb},
	}
	m.Decl = &exp.Lit{Res: val.Typ, Val: val, Src: c.Src}
	for _, el := range tags.Els {
		tag := el.(*exp.Tag)
		if tag.Exp == nil {
			return nil, ast.ErrReslSpec(tag.Src, c.Sig.Ref, fmt.Errorf("empty tag not allowed"))
		}
		te, err := p.Resl(c.Env, tag.Exp, typ.Void)
		if err != nil {
			return nil, err
		}
		tl, err := p.Eval(c.Env, te)
		if err != nil {
			return nil, err
		}
		pb.Params = append(pb.Params, typ.P(tag.Tag, te.Resl()))
		val.Vals = append(val.Vals, tl.Val)
		tag.Exp = tl
	}
	p.File.Refs = append(p.File.Refs, ModRef{Pub: true, Mod: m})
	return c, nil
}

func (s *modSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	return &exp.Lit{Val: typ.Void, Src: c.Src}, nil
}

var Use = &useSpec{impl("<form@use mods:<tupl|exp|alt str tag|str> void>"), false}
var Export = &useSpec{impl("<form@export mods:<tupl|exp|alt str tag|str> void>"), true}

type useSpec struct {
	exp.SpecBase
	export bool
}

func (s *useSpec) Value() lit.Val { return s }
func (s *useSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, _ typ.Type) (_ exp.Exp, err error) {
	if c.Env != nil {
		return c, nil
	}
	// lookup the loader env
	le := FindLoaderEnv(p.Root)
	top, _ := c.Args[0].(*exp.Tupl)
	for _, el := range top.Els {
		// get alias and path from argument
		var ref ModRef
		if t, ok := el.(*exp.Tag); ok {
			ref.Alias = t.Tag
			el = t.Exp
		}
		if l, ok := el.(*exp.Lit); ok {
			ref.Path = l.Value().String()
		} else {
			return nil, fmt.Errorf("unexpected use argument %T", el)
		}
		// load module using loader env
		f, err := le.LoadFile(p, ref.Path)
		if err != nil {
			return nil, fmt.Errorf("could not load module %q: %v", ref.Path, err)
		}
		// register modules in parent loader or mod env local
		if ref.Alias != "" {
			// TODO select module from decls
			fst, n := getPub(f.Refs)
			if n != 1 {
				return nil, fmt.Errorf("alias works only with single module units %q", ref.Path)
			}
			ref.Mod = fst.Mod
			ref.Pub = s.export
			p.File.Refs = append(p.File.Refs, ref)
		} else {
			for _, m := range f.Refs {
				if !m.Pub {
					continue
				}
				m.Path = ref.Path
				if ref.Alias != "" {
					m.Alias = ref.Alias
				}
				m.Pub = s.export
				p.File.Refs = append(p.File.Refs, m)
			}
		}
	}
	c.Env = env
	// keep the call around for printing
	return c, nil
}
func (s *useSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	return &exp.Lit{Val: typ.Void, Src: c.Src}, nil
}

func getPub(refs exp.ModRefs) (fst exp.ModRef, n int) {
	for _, r := range refs {
		if r.Pub {
			if n++; n == 1 {
				fst = r
			}
		}
	}
	return fst, n
}

func impl(sig string) exp.SpecBase {
	t, err := typ.Parse(sig)
	if err != nil {
		panic(fmt.Errorf("impl sig %s: %v", sig, err))
	}
	return exp.SpecBase{Decl: t}
}

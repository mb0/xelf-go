package mod

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
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
	me := NewModEnv(env, &p.File, c.Src)
	me.Name(name)
	c.Env = me
	// eval elements to build the result type and value
	tags := c.Args[1].(*exp.Tupl)
	// create module type
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
		me.Add(tag.Tag, tl.Val)
		tag.Exp = tl
	}
	me.Pub()
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
		loc := ParseLoc(ref.Path)
		f, err := le.LoadFile(p, loc)
		if err != nil {
			return nil, fmt.Errorf("could not load module %q: %v", ref.Path, err)
		}
		refs := filterRefs(f.Refs, loc.Frag())
		if ref.Alias != "" {
			if len(refs) > 1 {
				refs = filterRefs(refs, ref.Alias)
			}
		}
		if len(refs) == 0 {
			return nil, fmt.Errorf("no modules found for %s", ref.Path)
		}
		// register modules in parent loader or mod env local
		for _, m := range refs {
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
	c.Env = env
	// keep the call around for printing
	return c, nil
}

func (s *useSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	return &exp.Lit{Val: typ.Void, Src: c.Src}, nil
}

func matchRef(m exp.ModRef, s string) bool {
	return s == "" || m.Alias != "" && m.Alias == s || m.Name == s
}
func filterRefs(refs []exp.ModRef, find string) (pub []exp.ModRef) {
	for _, m := range refs {
		if m.Pub && matchRef(m, find) {
			pub = append(pub, m)
		}
	}
	return pub
}

func impl(sig string) exp.SpecBase {
	t, err := typ.Parse(sig)
	if err != nil {
		panic(fmt.Errorf("impl sig %s: %v", sig, err))
	}
	return exp.SpecBase{Decl: t}
}

package mod

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Module = &moduleSpec{impl("<form@module name:sym tags?:tupl|exp none>")}

type moduleSpec struct{ exp.SpecBase }

func (s *moduleSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, _ typ.Type) (_ exp.Exp, err error) {
	if c.Env != nil {
		return c, nil
	}
	name := c.Args[0].String()
	me := NewModEnv(env, &p.File)
	me.SetName(name)
	c.Env = me
	// eval elements to build the result type and value
	tags := c.Args[1].(*exp.Tupl)
	// create module type
	for _, el := range tags.Els {
		tag, _ := el.(*exp.Tag)
		if tag != nil {
			el = tag.Exp
		}
		if el == nil {
			return nil, ast.ErrReslSpec(tag.Src, c.Sig.Ref,
				fmt.Errorf("empty module declaration not allowed"))
		}
		el, err = p.Resl(c.Env, el, typ.Void)
		if err != nil {
			return nil, err
		}
		val, err := p.Eval(c.Env, el)
		if err != nil {
			return nil, err
		}
		vt := val.Type()
		var decl string
		if tag != nil {
			tag.Exp = exp.LitVal(val)
			decl = tag.Tag
		} else if vt.Kind&knd.Typ != 0 {
			t, err := typ.ToType(val)
			if err != nil || t.Ref == "" {
				return nil, fmt.Errorf("expected named type got %s", val)
			}
			decl = t.Ref
			t.Ref = me.Mod.Name + "." + decl
			val = t
		} else if vt.Kind&knd.Spec != 0 {
			t := val.Type()
			if t.Ref == "" {
				return nil, fmt.Errorf("expected named spec got %s", val)
			}
			decl = t.Ref
			t.Ref = me.Mod.Name + "." + decl
			val = &exp.SpecRef{Spec: val.Value().(exp.Spec), Decl: t}
		} else {
			return nil, fmt.Errorf("unexpected module declaration %s", val)
		}
		err = me.AddDecl(decl, val)
		if err != nil {
			return nil, err
		}
	}
	return c, me.Publish()
}

func (s *moduleSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	return lit.Null{}, nil
}

var Import = &importSpec{impl("<form@import mods:<tupl|exp|alt str tag|str> none>"), false}
var Export = &importSpec{impl("<form@export mods:<tupl|exp|alt str tag|str> none>"), true}

type importSpec struct {
	exp.SpecBase
	export bool
}

func (s *importSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, _ typ.Type) (_ exp.Exp, err error) {
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
			return nil, err
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
			err = p.File.AddRefs(m)
			if err != nil {
				return nil, err
			}
		}
	}
	c.Env = env
	// keep the call around for printing
	return c, nil
}

func (s *importSpec) Eval(p *exp.Prog, c *exp.Call) (lit.Val, error) {
	return lit.Null{}, nil
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

var impl = exp.MustSpecBase

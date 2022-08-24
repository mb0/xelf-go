package lib

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

var Dyn = &dynSpec{impl("<form@dyn tupl|exp @>")}

type dynSpec struct{ exp.SpecBase }

func (s *dynSpec) Value() lit.Val { return s }
func (s *dynSpec) Resl(p *exp.Prog, env exp.Env, c *exp.Call, h typ.Type) (exp.Exp, error) {
	if c.Env == nil {
		c.Env = env
	}
	d := c.Args[0].(*exp.Tupl)
	if len(d.Els) == 0 {
		return nil, ast.ErrEval(c.Src, "empty call is unexpected at this point", nil)
	}
	// TODO use exp type with hint as result arg
	fst, err := p.Resl(env, d.Els[0], typ.Void)
	if err != nil {
		fst := d.Els[0]
		src := fst.Source()
		return nil, ast.ErrEval(src, fmt.Sprintf("dyn resl failed for %s", fst), err)
	}
	if fst.Kind()&(knd.Sym|knd.Call) != 0 {
		// TODO lets check the resolved type of fst
		// we may know a sugar spec for lit res or impossible specs, if we expect a spec
		// we need to evaluate the argument before we can determine the spec
		return c, nil
	}
	var spec exp.Spec
	args := d.Els
	switch a := fst.(type) {
	case *exp.Lit:
		spec, args = litSpec(a, args)
	case *exp.Tag:
		got, err := env.Resl(p, &exp.Sym{Sym: ":"}, ":", false)
		if err != nil {
			break
		}
		spec = got.(*exp.Lit).Val.(exp.Spec)
		tag := fst.(*exp.Tag)
		src := tag.Src
		src.End = src.Pos
		src.End.Byte += int32(len(tag.Tag))
		nargs := append(make([]exp.Exp, 0, len(args)+1),
			&exp.Lit{Res: typ.Sym, Val: lit.Str(tag.Tag), Src: src},
			tag.Exp,
		)
		args = append(nargs, args[1:]...)
	}
	if spec == nil {
		return nil, fmt.Errorf("no spec for %[1]T %[1]s", d.Els[0])
	}
	sig, _ := p.Sys.Inst(spec.Type())
	args, err = exp.LayoutSpec(sig, args)
	if err != nil {
		return nil, fmt.Errorf("layout error for %s %s: %v", sig.Type(), args, err)
	}
	cc := &exp.Call{Sig: sig, Spec: spec, Args: args, Src: d.Src}
	return spec.Resl(p, env, cc, h)
}
func (s *dynSpec) Eval(p *exp.Prog, c *exp.Call) (*exp.Lit, error) {
	d := c.Args[0].(*exp.Tupl)
	if len(d.Els) == 0 {
		return nil, ast.ErrEval(c.Src, "empty call is unexpected at this point", nil)
	}
	// TODO use exp type with hint as result arg
	a, err := p.Eval(c.Env, d.Els[0])
	if err != nil {
		return nil, ast.ErrEval(a.Source(), fmt.Sprintf("dyn eval failed for %s", a), err)
	}
	spec, args := litSpec(a, d.Els)
	if spec == nil {
		return nil, fmt.Errorf("no spec for %[1]T %[1]s %s", a, a.Res)
	}
	sig, _ := p.Sys.Inst(spec.Type())
	args, err = exp.LayoutSpec(sig, args)
	if err != nil {
		return nil, fmt.Errorf("layout error for %s %s: %v", sig.Type(), args, err)
	}
	cc := &exp.Call{Sig: sig, Spec: spec, Args: args, Src: d.Src}
	ce, err := spec.Resl(p, c.Env, cc, typ.Void)
	if err != nil {
		return nil, ast.ErrEval(a.Source(), fmt.Sprintf("dyn resl call failed for %s", a), err)
	}
	cc = ce.(*exp.Call)
	return cc.Spec.Eval(p, cc)
}

func litSpec(a *exp.Lit, args []exp.Exp) (spec exp.Spec, _ []exp.Exp) {
	t := a.Val.Type()
	k := t.Kind & knd.All
	switch {
	case k == knd.All, k == knd.Data:
	case k&knd.Spec != 0:
		spec, args = a.Val.(exp.Spec), args[1:]
	case k == knd.Typ:
		spec = Make
	case k&knd.Num != 0:
		spec = Add
	case k&knd.Str != 0:
		spec = Cat
	case k&knd.Raw != 0:
		spec = Json
	case k&knd.List != 0:
		spec = Append
	case k&knd.Keyr != 0:
		spec = Mut
	}
	return spec, args
}

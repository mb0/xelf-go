package exp

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

type Dyn func(*Prog, Env, *Call, typ.Type) error

func DefaultDyn(env Env) Dyn {
	mut := lookupSpec(env, "mut")
	mak := lookupSpec(env, "make")
	call := lookupSpec(env, "call")
	return func(p *Prog, env Env, a *Call, h typ.Type) error {
		if len(a.Args) == 0 {
			return ast.ErrReslSpec(a.Src, "unexpected empty call", nil)
		}
		fst, err := p.Resl(env, a.Args[0], typ.Void)
		if err != nil {
			return err
		}
		ft := fst.Type()
		if ft.Kind == knd.Tag {
			tag := fst.(*Tag)
			a.Spec = lookupSpec(env, ":")
			if a.Spec == nil {
				return ast.ErrReslSpec(tag.Src,
					"tag call spec must be registered", nil)
			}
			src := tag.Src
			src.End = src.Pos
			src.End.Byte += int32(len(tag.Tag))
			a.Args = append(append(make([]Exp, 0, len(a.Args)+1),
				LitSrc(lit.Wrap(lit.Str(tag.Tag).Mut(), typ.Sym), src),
				tag.Exp,
			), a.Args[1:]...)
			return nil
		}
		switch rt := typ.Res(ft); rt.Kind & knd.All {
		case knd.Spec, knd.Form, knd.Func:
			if ft.Kind == knd.Lit {
				a.Spec = fst.(*Lit).Value().(Spec)
				a.Args = a.Args[1:]
			} else {
				a.Spec = call
			}
		case knd.Typ:
			a.Spec = mak
		case knd.Void:
		default:
			a.Spec = mut
		}
		if a.Spec == nil {
			return ast.ErrReslSpec(a.Src, fmt.Sprintf("unsupported dyn call for %s", ft), nil)
		}
		return nil
	}
}

func lookupSpec(env Env, key string) (s Spec) {
	if v, _ := LookupKey(env, key); v != nil {
		s, _ = v.Value().(Spec)
	}
	return s
}

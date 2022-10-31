package exp

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Select reads path and returns the selected literal from l or an error.
func Select(l *Lit, path string) (*Lit, error) {
	p, err := cor.ParsePath(path)
	if err != nil {
		return nil, err
	}
	return SelectPath(l, p)
}

// SelectPath returns the selected literal at path from l or an error.
func SelectPath(l *Lit, path cor.Path) (_ *Lit, err error) {
	t, err := typ.SelectPath(l.Res, path)
	if err != nil {
		return nil, err
	}
	v, err := lit.SelectPath(l.Val, path)
	if err != nil {
		return nil, err
	}
	return &Lit{Res: t, Val: v}, nil
}

func SelectLookup(l *Lit, k string, eval bool) (*Lit, error) {
	if l == nil || eval && l.Val == nil {
		return nil, ErrSymNotFound
	}
	p, err := cor.ParsePath(k)
	if err != nil {
		return nil, err
	}
	if l.Val != nil {
		val, err := lit.SelectPath(l.Val, p)
		if err == nil {
			return LitVal(val), nil
		} else if eval {
			return nil, err
		}
	}
	t, err := typ.SelectPath(l.Res, p)
	if err != nil {
		return nil, err
	}
	return &Lit{Res: t, Val: lit.Null{}}, nil
}

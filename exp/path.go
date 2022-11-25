package exp

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Select lookup reads path and returns the selected literal from l or an error.
// If eval is false we attempt to at least resolve the intended type if no value was found.
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

package exp

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// SelectLookup reads path and returns the selected literal from l or an error.
// If eval is false we attempt to at least resolve the intended type if no value was found.
func SelectLookup(v lit.Val, k string, eval bool) (lit.Val, error) {
	if v == nil {
		return nil, ErrSymNotFound
	}
	p, err := cor.ParsePath(k)
	if err != nil {
		return nil, err
	}
	val, err := lit.SelectPath(v, p)
	if err == nil || eval {
		return val, err
	}
	t, err := typ.SelectPath(v.Type(), p)
	if err != nil {
		return nil, err
	}
	return lit.AnyWrap(t), nil
}

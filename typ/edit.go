package typ

import (
	"fmt"
)

// EditFunc edits a type and returns the result or an error.
// A copy editor expects returned types that differ from the input to be already copied.
type EditFunc func(e *Editor) (Type, error)

// Editor is type context used to edit type. It references the parents type context.
type Editor struct {
	Type
	Parent *Editor
	seen   editMap
}
type editMap map[Body]Type

// Copy edits type t and all offspring types with f and returns the result or an error.
func Edit(t Type, f EditFunc) (Type, error) {
	e := Editor{Type: t}
	return e.edit(f)
}

// sub calls edit with a new child editor on t and returns the result or an error.
func (e *Editor) sub(t Type, f EditFunc) (Type, error) {
	if t.ID > 0 && e.Type.ID == t.ID {
		// TODO panic and investigate
		return t, fmt.Errorf("nested id for %s", t)
	}
	sub := Editor{Type: t, Parent: e, seen: e.seen}
	return sub.edit(f)
}

// edit calls f on its type and recurses on all offspring types and returns the result or an error.
func (e *Editor) edit(f EditFunc) (res Type, err error) {
	old := e.Body
	if old != nil {
		if t, ok := e.seen[old]; ok {
			return t, nil
		}
	}
	if res, err = f(e); err != nil {
		return
	}
	if old != nil && e.seen != nil {
		if len(e.seen) > 127 {
			return res, fmt.Errorf("runaway type edit")
		}
		e.seen[old] = res
	}
	if res.Body == nil {
		return res, nil
	}
	var sub Type
	if e.seen == nil {
		e.seen = make(editMap)
		if old != nil {
			e.seen[old] = res
		}
	}
	switch b := res.Body.(type) {
	case *Type:
		sub, err = e.sub(*b, f)
		if err == nil {
			*b = sub
		}
	case *AltBody:
		for i, a := range b.Alts {
			sub, err = e.sub(a, f)
			if err != nil {
				return
			}
			b.Alts[i] = sub
		}
	case *ParamBody:
		for i, p := range b.Params {
			sub, err = e.sub(p.Type, f)
			if err != nil {
				return
			}
			b.Params[i].Type = sub
		}
	}
	return
}

func Clone(r Type) Type {
	return clone(r, nil)
}
func clone(r Type, stack [][2]Body) Type {
	if r.Body == nil {
		return r
	}
	for _, o := range stack {
		if o[0] == r.Body {
			r.Body = o[1]
			return r
		}
	}
	switch b := r.Body.(type) {
	case *Type:
		n := new(Type)
		stack = append(stack, [2]Body{b, n})
		*n = clone(*b, stack)
		r.Body = n
	case *AltBody:
		n := &AltBody{Alts: make([]Type, len(b.Alts))}
		stack = append(stack, [2]Body{b, n})
		for i, a := range b.Alts {
			n.Alts[i] = clone(a, stack)
		}
		r.Body = n
	case *ParamBody:
		n := &ParamBody{Params: make([]Param, len(b.Params))}
		stack = append(stack, [2]Body{b, n})
		for i, p := range b.Params {
			p.Type = clone(p.Type, stack)
			n.Params[i] = p
		}
		r.Body = n
	}
	return r
}

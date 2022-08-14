package typ

// EditFunc edits a type and returns the result or an error.
// A copy editor expects returned types that differ from the input to be already copied.
type EditFunc func(e *Editor) (Type, error)

// Editor is type context used to edit type. It references the parents type context.
type Editor struct {
	Type
	Parent *Editor
	seen   editMap
	copy   bool
}
type editMap map[*ParamBody]*ParamBody

// Copy edits type t it and all offspring types with f and returns the result or an error.
func Edit(t Type, f EditFunc) (Type, error) {
	e := Editor{Type: t, copy: false}
	return e.edit(f)
}

// Copy copies type t editing it and all offspring types with f and returns the result or an error.
func Copy(t Type, f EditFunc) (Type, error) {
	e := Editor{Type: t, copy: true}
	return e.edit(f)
}

// sub calls edit with a new child editor on t and returns the result or an error.
func (e *Editor) sub(t Type, f EditFunc) (Type, error) {
	sub := Editor{Type: t, Parent: e, seen: e.seen, copy: e.copy}
	return sub.edit(f)
}

// edit calls f on its type and recurses on all offspring types and returns the result or an error.
func (e *Editor) edit(f EditFunc) (res Type, err error) {
	if f == nil {
		res = e.Type
	} else {
		res, err = f(e)
		if err != nil {
			return
		}
	}
	if res.Body == nil {
		return res, nil
	}
	if e.seen == nil {
		e.seen = make(editMap)
	}
	mod := e.Type.Body != res.Body
	var sub Type
	switch b := res.Body.(type) {
	case *ElBody:
		sub, err = e.sub(b.El, f)
		if err == nil {
			if e.copy && !mod {
				b = &ElBody{}
				res.Body, mod = b, true
			}
			b.El = sub
		}
	case *SelBody:
		sub, err = e.sub(b.Sel, f)
		if err == nil {
			if e.copy && !mod {
				b = &SelBody{Path: b.Path}
				res.Body, mod = b, true
			}
			b.Sel = sub
		}
	case *AltBody:
		for i, a := range b.Alts {
			sub, err = e.sub(a, f)
			if err != nil {
				return
			}
			if e.copy && !mod {
				b = &AltBody{Alts: append([]Type{}, b.Alts...)}
				res.Body, mod = b, true
			}
			b.Alts[i] = sub
		}
	case *ParamBody:
		if pb, ok := e.seen[b]; ok {
			if e.copy {
				res.Body = pb
			}
			return res, nil
		}
		if e.copy && !mod {
			old := b
			res.Ref = e.Ref
			b = &ParamBody{Params: append([]Param{}, b.Params...)}
			res.Body, mod = b, true
			e.seen[old] = b
		}
		e.seen[b] = b
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

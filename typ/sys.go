package typ

import (
	"fmt"
	"strings"

	"xelf.org/xelf/knd"
)

// Lookup is the type registry interface used to look up type references.
type Lookup = func(string) (Type, error)

// Sys is the resolution context used to instantiate, bind, update and unify types.
// Type unification is part of sys, because it needs close access to the type variable bindings.
type Sys struct {
	MaxID int32
	Map   map[int32]Type
}

// NewSys creates and returns an empty type system using the given type registry.
func NewSys() *Sys { return &Sys{Map: make(map[int32]Type)} }

// Bind binds type t using an existing id or sets a new id and returns the update type.
func (sys *Sys) Bind(t Type) Type {
	if t.ID <= 0 {
		sys.MaxID++
		t.ID = sys.MaxID
	}
	sys.Map[t.ID] = t
	return t
}

// Get returns the bound type for id or void.
func (sys *Sys) Get(id int32) Type {
	if id <= 0 {
		return Void
	}
	return sys.Map[id]
}

// Update updates all vars and refs in t with the currently bound types and returns the result.
func (sys *Sys) Update(lup Lookup, t Type) Type {
	t, _ = Edit(t, func(e *Editor) (Type, error) {
		if e.Kind&knd.Var != 0 {
			r := sys.Get(e.ID)
			if r != Void {
				return r, nil
			}
		}
		if e.Kind&knd.Ref != 0 && e.Ref != "" {
			r, err := sys.resolveRef(lup, e.Type)
			if err != nil {
				return e.Type, nil
			}
			if e.Kind&knd.None != 0 {
				r.Kind |= knd.None
			}
			return r, nil
		}
		return e.Type, nil
	})
	return t
}

// Inst instantiates all vars in t for sys and returns the result or an error.
func (sys *Sys) Inst(lup Lookup, t Type) (Type, error) {
	return sys.inst(lup, t, make(map[int32]Type))
}
func (sys *Sys) inst(lup Lookup, t Type, m map[int32]Type) (Type, error) {
	var deferSel bool
	res, err := Edit(Clone(t), func(e *Editor) (Type, error) {
		r := e.Type
		if r.ID > 0 {
			if r, ok := m[r.ID]; ok {
				return r, nil
			}
		}
		old := r.ID
		if r.ID != 0 || r.Kind&knd.Var != 0 {
			sys.MaxID++
			r.ID = sys.MaxID
		}
		if r.Kind&knd.Ref != 0 && r.Ref != "" {
			n, err := sys.resolveRef(lup, r)
			if err != nil {
				return r, fmt.Errorf("resolve type ref %q: %v", r.Ref, err)
			} else {
				r = n
			}
		} else if _, ok := r.Body.(*SelBody); ok {
			deferSel = true
			return r, nil
		}
		if old > 0 {
			m[old] = r
		}
		return r, nil
	})
	if !deferSel {
		return res, err
	}
	return Edit(res, func(e *Editor) (Type, error) {
		if b, ok := e.Type.Body.(*SelBody); ok {
			return resolveSel(e, b.Path)
		}
		return e.Type, nil
	})
}

func resolveSel(e *Editor, path string) (Type, error) {
	cur := e
	rest := path
	for rest != "" && rest[0] == '.' {
		for {
			cur = cur.Parent
			if cur == nil {
				return e.Type, fmt.Errorf("selection %s not found in %v", path, e.Type)
			}
			if cur.Type.Kind&(knd.Obj|knd.Spec) != 0 {
				break
			}
		}
		rest = rest[1:]
	}
	p := path[len(path)-len(rest)-1:]
	return Select(cur.Type, p)
}

func (sys *Sys) resolveRef(lup Lookup, t Type) (Type, error) {
	if t.Ref[0] == '.' {
		// skip localref
		return t, nil
	}
	if lup == nil {
		return t, fmt.Errorf("no type lookup configured")
	}
	// try the first part
	ref, rest, sel := strings.Cut(t.Ref, ".")
	n, err := lup(ref)
	if err != nil {
		if !sel {
			return t, err
		}
		// modules are not yet implemented, so check for schema qualified types
		// we can at least expect a flat schema.model structure
		idx := strings.IndexByte(rest, '.')
		if idx >= 0 {
			idx += 1 + len(ref)
			ref, rest, sel = t.Ref[:idx], t.Ref[idx+1:], true
		} else {
			ref, rest, sel = t.Ref, "", false
		}
		n, err = lup(ref)
		if err != nil {
			return t, err
		}
	}
	// normalize ref
	ref = n.Ref
	if sel {
		ref += "." + rest
		if n.Kind&knd.Ref != 0 {
			n = Ref(ref)
		} else {
			// find selection
			n, err = Select(n, rest)
			if err != nil {
				return t, err
			}
			n.Ref = ref
		}
	}
	if t.Kind&knd.None != 0 {
		n.Kind |= knd.None
	}
	return n, nil
}

// Unify unifies type t and h and returns the result or an error.
// Type unification in this context means that we have two types that should describe the same type.
// We then check where these descriptions overlap and use the result instead of the input types.
func (sys *Sys) Unify(lup Lookup, t, h Type) (Type, error) {
	t = sys.Update(lup, t)
	if h == Void {
		return t, nil
	}
	h = sys.Update(lup, h)
	r, err := unify(sys, lup, t, h)
	if err != nil {
		return t, err
	}
	r = unibind(sys, t, h, r)
	return r, nil
}

func unify(sys *Sys, lup Lookup, t, h Type) (Type, error) {
	a := base(t)
	b := base(h)
	kk := a.Kind | b.Kind
	if kk&(knd.Sel|knd.Ref) != 0 {
		return Void, fmt.Errorf("cannot unify sel or ref %s with %s", t, h)
	}
	r := a
	ak := a.Kind &^ (knd.Var | knd.None)
	bk := b.Kind &^ (knd.Var | knd.None)
	if ak == 0 {
		r.Kind = b.Kind
		r.Body = b.Body
		return unibind(sys, a, b, r), nil
	}
	if bk == 0 {
		return unibind(sys, a, b, r), nil
	}
	if kk&knd.Alt != 0 {
		x, y := a, b
		if ak&knd.Alt == 0 {
			x, y = b, a
		}
		// TODO merge alts
		_, _ = x, y
	}
	if ak == bk {
		if t.Ref == h.Ref && equalBody(a.Body, b.Body) {
			return unibind(sys, a, b, r), nil
		}
		if a.Body == nil {
			a, b, r, ak, bk = b, a, b, bk, ak
		}
	Switch:
		switch ab := a.Body.(type) {
		case *ElBody:
			bb, ok := b.Body.(*ElBody)
			if ak&knd.Tupl != 0 {
				if !ok {
					return unibind(sys, a, b, r), nil
				}
			}
			if ok {
				el, err := sys.Unify(lup, ab.El, bb.El)
				if err != nil {
					return Void, err
				}
				r.Body = &ElBody{El: el}
			}
			return unibind(sys, a, b, r), nil
		case *ParamBody:
			r.Ref = "_"
			bb, ok := b.Body.(*ParamBody)
			if ak&knd.Tupl != 0 {
				if !ok {
					return unibind(sys, a, b, r), nil
				}
			}
			if len(ab.Params) > len(bb.Params) {
				break
			}
			ps := make([]Param, 0, len(ab.Params))
			for i, p := range ab.Params {
				op := bb.Params[i]
				if p.Name != op.Name || p.Key != op.Key ||
					p.Type.Kind&^knd.None != op.Type.Kind&^knd.None ||
					(p.Type.Body == nil) != (op.Type.Body == nil) {
					break Switch
				}
				if p.Type.Body != nil && !p.Type.Body.EqualHist(op.Type.Body, nil) {
					break Switch
				}
				ps = append(ps, p)
			}
			r.Body = &ParamBody{Params: ps}
			return unibind(sys, a, b, r), nil
		}
	} else {
		k := a.Kind & knd.Any
		if k == 0 {
			k = b.Kind & knd.Any
		} else {
			k = k & b.Kind
		}
		if k != 0 {
			if equalBody(a.Body, b.Body) {
				r.Kind = k
				return unibind(sys, a, b, r), nil
			}
			if a.Body == nil {
				r.Kind = k
				r.Body = b.Body
				return unibind(sys, a, b, r), nil
			}
			if b.Body == nil {
				r.Kind = k
				return unibind(sys, a, b, r), nil
			}
		}
	}
	return Void, fmt.Errorf("cannot unify %s with %s", t, h)
}

func equalBody(a, b Body) bool {
	if a == nil {
		return b == nil
	}
	return a.EqualHist(b, nil)
}

func base(t Type) Type {
	for t.Kind&knd.Exp == knd.Exp || t.Kind&knd.Exp != 0 && t.Kind&knd.Tupl == 0 {
		b, ok := t.Body.(*ElBody)
		if ok && b.El != Void {
			t = b.El
		} else {
			t = Any
		}
	}
	return t
}

func unibind(sys *Sys, a, b, r Type) Type {
	if a.ID > 0 {
		if r.ID <= 0 {
			r.ID = a.ID
		}
		sys.Map[a.ID] = r
	}
	if b.ID > 0 {
		if r.ID <= 0 {
			r.ID = b.ID
		}
		sys.Map[b.ID] = r
	}
	if r.ID > 0 {
		sys.Map[r.ID] = r
		if r.Kind == 0 || r.Kind.IsAlt() {
			r.Kind |= knd.Var
		}
	}
	return r
}

// Free appends all unbound type variables in t to res and returns the result.
func (c *Sys) Free(lup Lookup, t Type, res []Type) []Type {
	if t.Kind&(knd.Var|knd.Ref|knd.Sel) != 0 {
		if t.ID > 0 {
			// skip if we already collect this type
			for _, r := range res {
				if r.ID == t.ID {
					return res
				}
			}
			// update from context
			t = c.Update(lup, t)
		}
		// append if still a free type
		if t.Kind&(knd.Var|knd.Ref|knd.Sel) != 0 {
			res = append(res, t)
		}
	}
	switch b := t.Body.(type) {
	case *ElBody:
		res = c.Free(lup, b.El, res)
	case *AltBody:
		for _, a := range b.Alts {
			res = c.Free(lup, a, res)
		}
	case *ParamBody:
		for _, p := range b.Params {
			res = c.Free(lup, p.Type, res)
		}
	}
	return res
}

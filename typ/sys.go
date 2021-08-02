package typ

import (
	"fmt"

	"xelf.org/xelf/knd"
)

// Reg is the type registry interface used to resolve type references.
// The Reg of package lit is the only noteworthy implementation.
type Reg interface{ RefType(string) (Type, error) }

// Sys is the resolution context used to instantiate, bind, update and unify types.
// Type unification is part of sys, because it needs close access to the type variable bindings.
type Sys struct {
	MaxID int32
	Map   map[int32]Type
	Reg   Reg
}

// NewSys creates and returns an empty type system using the given type registry.
func NewSys(reg Reg) *Sys {
	return &Sys{Map: make(map[int32]Type), Reg: reg}
}

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
func (sys *Sys) Update(t Type) Type { t, _ = Edit(t, sys.update); return t }
func (sys *Sys) update(e *Editor) (Type, error) {
	if e.Kind&knd.Var != 0 {
		r := sys.Get(e.ID)
		if r != Void {
			return r, nil
		}
	}
	if e.Kind&knd.Ref != 0 {
		b, ok := e.Type.Body.(*RefBody)
		if ok {
			r, err := sys.Reg.RefType(b.Ref)
			if err != nil {
				return e.Type, nil
			}
			if e.Kind&knd.None != 0 {
				r.Kind |= knd.None
			}
			return r, nil
		}
	}
	return e.Type, nil
}

// Inst instantiates all vars in t for sys and returns the result or an error.
func (sys *Sys) Inst(t Type) (Type, error) { return sys.inst(t, make(map[int32]Type)) }
func (sys *Sys) inst(t Type, m map[int32]Type) (Type, error) {
	return Copy(t, func(e *Editor) (Type, error) {
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
		switch b := r.Body.(type) {
		case *RefBody:
			if sys.Reg != nil {
				n, err := sys.Reg.RefType(b.Ref)
				if err != nil {
					break // ignore error here
				}
				if r.Kind&knd.None != 0 {
					n.Kind |= knd.None
				}
				return n, nil
			}
		case *SelBody:
			// TODO resolve path or think about type selections
			var par *Editor
			for par = e.Parent; par != nil; par = par.Parent {
				if par.Type.Kind&(knd.Strc|knd.Spec) != 0 {
					_, ok := par.Type.Body.(*ParamBody)
					if ok {
						break
					}
				}
			}
			return Select(par.Type, b.Path)
		}
		if old > 0 {
			m[old] = r
		}
		return r, nil
	})
}

// Unify unifies type t and h and returns the result or an error.
// Type unification in this context means that we have two types that should describe the same type.
// We then check where these descriptions overlap and use the result instead of the input types.
func (sys *Sys) Unify(t, h Type) (Type, error) {
	t = sys.Update(t)
	if h == Void {
		return t, nil
	}
	h = sys.Update(h)
	r, err := unify(sys, t, h)
	if err != nil {
		return t, err
	}
	r = unibind(sys, t, h, r)
	return r, nil
}

func unify(sys *Sys, t, h Type) (Type, error) {
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
		if equalBody(a.Body, b.Body) {
			return unibind(sys, a, b, r), nil
		}
		switch ab := a.Body.(type) {
		case *ElBody:
			bb, ok := b.Body.(*ElBody)
			if ok {
				el, err := sys.Unify(ab.El, bb.El)
				if err != nil {
					return Void, err
				}
				r.Body = &ElBody{El: el}
			}
			return unibind(sys, a, b, r), nil
		case *ParamBody:
			_, ok := b.Body.(*ParamBody)
			if ak&knd.Tupl != 0 {
				if !ok {
					return unibind(sys, a, b, r), nil
				}
			}
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
	return a.Equal(b)
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
func (c *Sys) Free(t Type, res []Type) []Type {
	if t.Kind&(knd.Var|knd.Ref|knd.Sel) != 0 {
		if t.ID > 0 {
			// skip if we already collect this type
			for _, r := range res {
				if r.ID == t.ID {
					return res
				}
			}
			// update from context
			t = c.Update(t)
		}
		// append if still a free type
		if t.Kind&(knd.Var|knd.Ref|knd.Sel) != 0 {
			res = append(res, t)
		}
	}
	switch b := t.Body.(type) {
	case *ElBody:
		res = c.Free(b.El, res)
	case *AltBody:
		for _, a := range b.Alts {
			res = c.Free(a, res)
		}
	case *ParamBody:
		for _, p := range b.Params {
			res = c.Free(p.Type, res)
		}
	}
	return res
}

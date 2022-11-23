package lit

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// Reg is a registry for custom mutable values and provides api to work proxies in general. Reg uses
// the global reflection cache by default, to support self referential types and improve efficiency.
type Reg struct {
	refs  map[string]Mut
	Cache *Cache
}

func (reg *Reg) init() {
	if reg.Cache == nil {
		reg.Cache = DefaultCache
	}
}

// SetRef registers type and optionally a mutable implementation for ref.
func (reg *Reg) SetRef(ref string, mut Mut) {
	if reg.refs == nil {
		reg.refs = make(map[string]Mut)
	}
	reg.refs[cor.Keyed(ref)] = mut
}

func (reg *Reg) Each(f func(string, Mut) error) error {
	for ref, mut := range reg.refs {
		if err := f(ref, mut); err != nil {
			return err
		}
	}
	return nil
}

// Zero returns a mutable zero value for t or an error.
func (reg *Reg) Zero(t typ.Type) (m Mut, err error) {
	reg.init()
	k := t.Kind & knd.All
	if t.Ref != "" {
		if m := reg.refs[cor.Keyed(t.Ref)]; m != nil {
			return m.New()
		}
	}
	if k == knd.List {
		if e := typ.ContEl(t); e.Ref != "" {
			if m := reg.refs[cor.Keyed(e.Ref)]; m != nil {
				if s, ok := m.(interface{ Slice() Mut }); ok {
					return s.Slice(), nil
				}
			}
		}
	}
	switch k {
	case knd.Typ:
		t = typ.El(t)
		m = &t
	default:
		m = Zero(t)
		if m == nil {
			return newAnyPrx(reg, t), nil
		}
	}
	if t.Kind&knd.None != 0 {
		m = &OptMut{m, nil, true}
	}
	return m, nil
}

// AddFrom updates the registry and reflect cache with entries from o.
func (reg *Reg) AddFrom(o *Reg) {
	if reg.refs == nil {
		reg.refs = make(map[string]Mut, len(o.refs))
	}
	for ref, mut := range o.refs {
		if _, ok := reg.refs[ref]; !ok {
			reg.refs[ref] = mut
		}
	}
	if reg.Cache == nil {
		reg.Cache = o.Cache
	} else {
		reg.Cache.AddFrom(o.Cache)
	}
}

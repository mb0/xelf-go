package lit

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type refInfo struct {
	Type typ.Type
	Mut  Mut
}

// Reg is a registry for custom mutable values and provides api to work proxies in general. Reg uses
// the global reflection cache by default, to support self referential types and improve efficiency.
type Reg struct {
	refs  map[string]refInfo
	Cache *Cache
}

func (reg *Reg) init() {
	if reg.Cache == nil {
		reg.Cache = DefaultCache
	}
}

// SetRef registers type and optionally a mutable implementation for ref.
func (reg *Reg) SetRef(ref string, t typ.Type, mut Mut) {
	if reg.refs == nil {
		reg.refs = make(map[string]refInfo)
	}
	t.Ref = ref
	ref = cor.Keyed(ref)
	reg.refs[ref] = refInfo{t, mut}
}

func (reg *Reg) Each(f func(string, typ.Type, Mut) error) error {
	for ref, info := range reg.refs {
		err := f(ref, info.Type, info.Mut)
		if err != nil {
			return err
		}
	}
	return nil
}

// Zero returns a mutable zero value for t or an error.
func (reg *Reg) Zero(t typ.Type) (m Mut, err error) {
	reg.init()
	if t.Kind&knd.Idxr == knd.List {
		n := typ.ContEl(t).Ref
		if n != "" {
			nfo := reg.refs[n]
			if nfo.Mut != nil {
				if s, ok := nfo.Mut.(interface{ Slice() Mut }); ok {
					return s.Slice(), nil
				}
			}
		}
	} else {
		if t.Ref != "" {
			nfo := reg.refs[t.Ref]
			if nfo.Mut != nil {
				return nfo.Mut.New()
			}
		}
	}
	k := t.Kind & knd.All
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

// AddFrom updates the registry with entries from o.
func (reg *Reg) AddFrom(o *Reg) {
	for ref, r := range o.refs {
		if ri, ok := reg.refs[ref]; !ok || ri.Type == typ.Void || ri.Mut == nil {
			reg.SetRef(ref, r.Type, reg.copyMut(r.Mut))
		}
	}
	reg.Cache = o.Cache
}

func (reg *Reg) copyMut(p Mut) Mut {
	if p != nil {
		p, _ = p.New()
		if wr, ok := p.(interface{ WithReg(*Reg) }); ok {
			wr.WithReg(reg)
		}
	}
	return p
}

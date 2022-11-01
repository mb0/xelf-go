package lit

import (
	"fmt"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// Reg is a registry context for type references, reflected types and proxies. Many functions and
// container literals have an optional registry to aid in value conversion and construction.
type Reg struct {
	refs  map[string]refInfo
	Cache *Cache
}

func (reg *Reg) init() {
	if reg.Cache == nil {
		reg.Cache = DefaultCache
	}
}

type refInfo struct {
	Type typ.Type
	Mut  Mut
}
type typInfo struct {
	typ.Type
	*params
}
type params struct {
	ps  []typ.Param
	idx [][]int
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

// LookupType returns a type for ref or an error.
func (reg *Reg) LookupType(ref string) (typ.Type, error) {
	nfo, ok := reg.refs[cor.Keyed(ref)]
	if ok && nfo.Type != typ.Void {
		return nfo.Type, nil
	}
	return typ.Void, fmt.Errorf("no type found named %s", ref)
}

// Zero returns a zero mutable value for t or an error.
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
	if k.Count() != 1 {
		switch {
		case k&knd.Num != 0 && k&^knd.Num == 0:
			m = new(Num)
		case k&knd.Char != 0 && k&^knd.Char == 0:
			m = new(Char)
		case k&knd.Keyr != 0 && k&^knd.Keyr == 0:
			m = &Keyed{}
		case k&knd.Idxr != 0 && k&^knd.Idxr == 0:
			m = &Vals{}
		default:
			return newAnyPrx(reg, t), nil
		}
	} else {
		switch k {
		case knd.Typ:
			t = typ.El(t)
			m = &t
		case knd.Bool:
			m = new(Bool)
		case knd.Int:
			m = new(Int)
		case knd.Real:
			m = new(Real)
		case knd.Str:
			m = new(Str)
		case knd.Raw:
			m = new(Raw)
		case knd.UUID:
			m = new(UUID)
		case knd.Time:
			m = new(Time)
		case knd.Span:
			m = new(Span)
		case knd.List:
			m = &List{Reg: reg, El: typ.ContEl(t)}
		case knd.Dict:
			m = &Dict{Reg: reg, El: typ.ContEl(t)}
		case knd.Obj:
			m, err = NewObj(reg, t)
			if err != nil {
				return nil, err
			}
		default:
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

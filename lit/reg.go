package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// Reg is a registry context for type references, reflected types and proxies. Many functions and
// container literals have an optional registry to aid in value conversion and construction.
type Reg struct {
	refs  map[string]refInfo
	proxy map[reflect.Type]Prx
	param map[reflect.Type]typInfo
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
	reg.refs[ref] = refInfo{t, mut}
}

// RefType returns a type for ref or an error.
func (reg *Reg) RefType(ref string) (typ.Type, error) {
	nfo, ok := reg.refs[cor.Keyed(ref)]
	if ok && nfo.Type != typ.Void {
		return nfo.Type, nil
	}
	return typ.Void, fmt.Errorf("no type found named %s", ref)
}

// Zero returns a zero mutable value for t or an error.
func (reg *Reg) Zero(t typ.Type) (m Mut, err error) {
	if t.Kind&knd.List != 0 {
		n := typ.Name(typ.El(t))
		if n != "" {
			nfo := reg.refs[n]
			if nfo.Mut != nil {
				if s, ok := nfo.Mut.(interface{ Slice() Mut }); ok {
					return s.Slice(), nil
				}
			}
		}
	} else {
		n := typ.Name(t)
		if n != "" {
			nfo := reg.refs[n]
			if nfo.Mut != nil {
				return nfo.Mut.New()
			}
		}
	}
	k := t.Kind & knd.All
	if k.Count() != 1 {
		switch {
		case k&knd.Num != 0 && k&^knd.Num == 0:
			m = new(Int)
		case k&knd.Str != 0 && k&^knd.Char == 0:
			m = new(Str)
		case k&knd.List != 0 && k&^knd.Idxr == 0:
			m = &List{Reg: reg}
		case k&knd.Dict != 0 && k&^knd.Keyr == 0:
			m = &Dict{Reg: reg}
		default:
			var any interface{}
			return &AnyPrx{proxy{reg, t, reflect.ValueOf(&any)}, Null{}}, nil
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
			m = &List{Reg: reg}
		case knd.Dict:
			m = &Dict{Reg: reg}
		case knd.Rec, knd.Obj:
			m, err = NewStrc(reg, t)
			if err != nil {
				return nil, err
			}
		default:
			var any interface{}
			return &AnyPrx{proxy{reg, t, reflect.ValueOf(&any)}, Null{}}, nil
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
	for rt, p := range o.proxy {
		if _, ok := reg.proxy[rt]; !ok && p != nil {
			reg.setProxy(rt, reg.copyMut(p).(Prx))
		}
	}
	for rt, p := range o.param {
		if _, ok := reg.param[rt]; !ok {
			reg.setParam(rt, p)
		}
	}
}
func (reg *Reg) setParam(rt reflect.Type, nfo typInfo) {
	if reg.param == nil {
		reg.param = make(map[reflect.Type]typInfo)
	}
	reg.param[rt] = nfo
}
func (reg *Reg) setProxy(rt reflect.Type, prx Prx) {
	if reg.proxy == nil {
		reg.proxy = make(map[reflect.Type]Prx)
	}
	reg.proxy[rt] = prx
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

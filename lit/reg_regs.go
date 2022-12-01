package lit

import (
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

var global = NewRegs()

// Regs embeds both PrxReg and MutReg to make it easier to provide and pass around.
// It implements the Reg interface by embedding PrxReg.
type Regs struct {
	*PrxReg
	MutReg
}

// NewRegs returns a new instance with a pristine PrxReg.
func NewRegs() *Regs { return &Regs{PrxReg: &PrxReg{}} }

// GlobalRegs returns the global default registry.
func GlobalRegs() *Regs { g := *global; return &g }

// Default applies sensible defaults to rs and returns it.
func DefaultRegs(rs *Regs) *Regs {
	if rs == nil {
		rs = &Regs{}
	}
	if rs.PrxReg == nil {
		rs.PrxReg = global.PrxReg
	}
	return rs
}

// Update updates the registry and reflect cache with entries from o.
func UpdateRegs(rs, o *Regs) {
	if o == nil || rs == o {
		return
	}
	o.MutReg.AddFrom(o.MutReg)
	if rs.PrxReg == nil {
		rs.PrxReg = &PrxReg{}
	}
	rs.PrxReg.AddFrom(o.PrxReg)
}

// MutReg stores a map of custom mutable implementations by reference.
type MutReg map[string]Mut

// Zero returns a mutable zero value for t. It provides registered mutables for name types and falls
// back to the default zero function.
func (mr MutReg) Zero(t typ.Type) Mut {
	if mr != nil {
		k := t.Kind & knd.All
		if t.Ref != "" {
			if m := mr[cor.Keyed(t.Ref)]; m != nil {
				return m.New()
			}
		}
		if k == knd.List {
			if e := typ.ContEl(t); e.Ref != "" {
				if m := mr[cor.Keyed(e.Ref)]; m != nil {
					if s, ok := m.(interface{ Slice() Mut }); ok {
						return s.Slice()
					}
				}
			}
		}
	}
	return Zero(t)
}
func (mr MutReg) ZeroWrap(t typ.Type) Mut {
	v := mr.Zero(t)
	if v.Type() == t {
		return v
	}
	return Wrap(v, t)
}

func (mr MutReg) Each(f func(string, Mut) error) error {
	for ref, mut := range mr {
		if err := f(ref, mut); err != nil {
			return err
		}
	}
	return nil
}

// SetRef registers type and optionally a mutable implementation for ref.
func (mr MutReg) SetRef(ref string, mut Mut) error {
	mr[cor.Keyed(ref)] = mut
	return nil
}

// AddFrom updates the registry and reflect cache with entries from o.
func (mr *MutReg) AddFrom(o MutReg) {
	if o != nil {
		if *mr == nil {
			*mr = make(MutReg)
		}
		o.Each(mr.SetRef)
	}
}

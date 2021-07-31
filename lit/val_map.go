package lit

import (
	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type Map struct {
	Reg *Reg
	El  typ.Type
	M   map[string]Val
}

func (h *Map) Type() typ.Type               { return typ.DictOf(h.El) }
func (h *Map) Nil() bool                    { return h == nil }
func (h *Map) Zero() bool                   { return len(h.M) == 0 }
func (h *Map) Value() Val                   { return h }
func (h *Map) MarshalJSON() ([]byte, error) { return bfr.JSON(h) }
func (h *Map) UnmarshalJSON(b []byte) error { return unmarshal(b, h) }
func (h *Map) String() string               { return bfr.String(h) }
func (h *Map) Print(p *bfr.P) (err error) {
	p.Byte('{')
	var i int
	for k, v := range h.M {
		if i > 0 {
			p.Sep()
		}
		i++
		p.RecordKey(k)
		if err = v.Print(p); err != nil {
			return err
		}
	}
	return p.Byte('}')
}

func (h *Map) New() (Mut, error) { return &Map{h.Reg, h.El, nil}, nil }
func (h *Map) WithReg(reg *Reg)  { h.Reg = reg }
func (h *Map) Ptr() interface{}  { return h }
func (h *Map) Parse(a ast.Ast) error {
	if isNull(a) {
		h.M = nil
		return nil
	}
	if a.Kind != knd.Dict {
		return ast.ErrExpect(a, knd.Dict)
	}
	if h.M == nil {
		h.M = make(map[string]Val, len(a.Seq))
	} else if len(h.M) > 0 {
		for k := range h.M {
			delete(h.M, k)
		}
	}
	for _, e := range a.Seq {
		key, val, err := ast.UnquotePair(e)
		if err != nil {
			return err
		}
		el, err := h.Reg.parseMutNull(val)
		if err != nil {
			return err
		}
		h.M[key] = el
	}
	return nil
}
func (h *Map) Assign(p Val) error {
	// TODO check types
	switch o := p.(type) {
	case nil:
		h.M = nil
	case Null:
		h.M = nil
	case Keyr:
		res := make(map[string]Val, o.Len())
		o.IterKey(func(k string, v Val) error {
			res[k] = v
			return nil
		})
		h.M = res
	default:
		return ErrAssign
	}
	return nil
}

func (h *Map) Len() int { return len(h.M) }
func (h *Map) Keys() []string {
	if len(h.M) == 0 {
		return nil
	}
	res := make([]string, 0, len(h.M))
	for k := range h.M {
		res = append(res, k)
	}
	return res
}
func (h *Map) Key(k string) (Val, error) {
	v := h.M[k]
	// TODO think about zero values of keyr go uses zero value and js undefined values
	return v, nil
}
func (h *Map) SetKey(k string, el Val) error {
	if el == nil {
		el = Null{}
	}
	if h.El != typ.Void {
		// TODO check and conevert el
	}
	h.M[k] = el
	return nil
}
func (h *Map) IterKey(it func(string, Val) error) error {
	for k, v := range h.M {
		if err := it(k, v); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}

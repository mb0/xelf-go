package lit

import (
	"fmt"
	"sort"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type Map struct {
	Typ typ.Type
	M   map[string]Val
}

func NewMap(el typ.Type) *Map {
	return &Map{Typ: typ.DictOf(el)}
}
func (h *Map) Type() typ.Type {
	if h.Typ == typ.Void {
		h.Typ = typ.Dict
	}
	return h.Typ
}
func (h *Map) Nil() bool  { return h == nil }
func (h *Map) Zero() bool { return len(h.M) == 0 }
func (h *Map) Mut() Mut   { return h }
func (h *Map) Value() Val { return h }
func (h *Map) As(t typ.Type) (Val, error) {
	if h.Typ.AssignableTo(t) {
		h.Typ = t
		return h, nil
	}
	if ok := h.Typ.ConvertibleTo(t); ok {
		neu := typ.ContEl(t)
		for _, el := range h.M {
			if !el.Type().ConvertibleTo(neu) {
				ok = false
				break
			}
		}
		if ok {
			// TODO convert els
			h.Typ = t
			return h, nil
		}
	}
	// TODO obj type
	return nil, fmt.Errorf("cannot convert %T from %s to %s", h, h.Type(), t)
}

func (h *Map) MarshalJSON() ([]byte, error) { return bfr.JSON(h) }
func (h *Map) UnmarshalJSON(b []byte) error { return unmarshal(b, h) }
func (h *Map) String() string               { return bfr.String(h) }
func (h *Map) Print(p *bfr.P) (err error) {
	keys := h.Keys()
	sort.Strings(keys)
	p.Byte('{')
	for i, k := range keys {
		if i > 0 {
			p.Sep()
		}
		v, ok := h.M[k]
		if !p.JSON && (!ok || v.Nil()) {
			p.RecordKeyTag(k, ';')
			continue
		}
		p.RecordKey(k)
		if err = v.Print(p); err != nil {
			return err
		}
	}
	return p.Byte('}')
}

func (h *Map) New() Mut         { return &Map{h.Typ, nil} }
func (h *Map) Ptr() interface{} { return h }
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
		el, err := parseMutNull(val)
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
	if el == nil { // if el is explicitly nil delete the value
		delete(h.M, k)
		return nil
	}
	if h.M == nil {
		h.M = make(map[string]Val)
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

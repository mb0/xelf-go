package lit

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type KeyVal struct {
	Key string
	Val
}
type Keyed []KeyVal

func (d Keyed) Type() typ.Type { return typ.Keyr }
func (d *Keyed) Nil() bool     { return d == nil }
func (d *Keyed) Zero() bool    { return d == nil || len(*d) == 0 }
func (d *Keyed) Mut() Mut      { return d }
func (d *Keyed) Value() Val    { return d }
func (d *Keyed) As(t typ.Type) (Val, error) {
	if typ.Keyr.AssignableTo(t) {
		return &Dict{Typ: t, Keyed: *d}, nil
	}
	if ok := typ.Keyr.ConvertibleTo(t); ok {
		neu := typ.ContEl(t)
		for _, kv := range *d {
			if !kv.Type().ConvertibleTo(neu) {
				ok = false
				break
			}
		}
		if ok {
			return &Dict{Typ: t, Keyed: *d}, nil
		}
	}
	// TODO obj type
	return nil, fmt.Errorf("cannot convert %T from %s to %s", d, d.Type(), t)
}

func (d *Keyed) UnmarshalJSON(b []byte) error { return unmarshal(b, d) }
func (d Keyed) MarshalJSON() ([]byte, error)  { return bfr.JSON(d) }
func (d Keyed) String() string                { return bfr.String(d) }
func (d Keyed) Print(p *bfr.P) (err error) {
	p.Byte('{')
	for i, v := range d {
		if i > 0 {
			p.Sep()
		}
		if !p.JSON && v.Val.Nil() {
			p.RecordKeyTag(v.Key, ';')
		} else {
			p.RecordKey(v.Key)
			if err = v.Val.Print(p); err != nil {
				return err
			}
		}
	}
	return p.Byte('}')
}

func (d *Keyed) New() Mut         { return &Keyed{} }
func (d *Keyed) Ptr() interface{} { return d }
func (d *Keyed) Parse(a ast.Ast) error {
	if isNull(a) {
		*d = nil
		return nil
	}
	if a.Kind != knd.Keyr {
		return ast.ErrExpect(a, knd.Keyr)
	}
	kvs := make([]KeyVal, 0, len(a.Seq))
	for _, e := range a.Seq {
		key, val, err := ast.UnquotePair(e)
		if err != nil {
			return err
		}
		el, err := parseMutNull(val)
		if err != nil {
			return err
		}
		kvs = append(kvs, KeyVal{key, el})
	}
	*d = kvs
	return nil
}
func (d *Keyed) Assign(p Val) error {
	// TODO check types
	switch o := Unwrap(p).(type) {
	case nil:
		*d = nil
	case Null:
		*d = nil
	case Keyr:
		// TODO check types
		res := make([]KeyVal, 0, o.Len())
		o.IterKey(func(k string, v Val) error {
			res = append(res, KeyVal{k, Unwrap(v)})
			return nil
		})
		*d = res
	default:
		return fmt.Errorf("%w %T to %T", ErrAssign, p, d)
	}
	return nil
}

func (d *Keyed) Len() int {
	return len(*d)
}
func (d *Keyed) Idx(i int) (Val, error) {
	if i < 0 || i >= len(*d) {
		return nil, ErrIdxBounds
	}
	return (*d)[i].Val, nil
}
func (d *Keyed) SetIdx(i int, el Val) error {
	if i < 0 {
		return ErrIdxBounds
	}
	if el == nil {
		el = Null{}
	}
	s := *d
	if i >= len(s) {
		if i < cap(s) {
			n := s[:i+1]
			for j := len(s) - 1; j < i; j++ {
				n[j] = KeyVal{}
			}
		} else {
			n := make([]KeyVal, (i+1)*3/2)
			copy(n, s)
			s = n
		}
	}
	s[i].Val = el
	*d = s
	return nil
}
func (d *Keyed) IterIdx(it func(int, Val) error) error {
	for i, kv := range *d {
		if err := it(i, kv.Val); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (d *Keyed) Keys() []string {
	if d == nil || len(*d) == 0 {
		return nil
	}
	res := make([]string, 0, len(*d))
	for _, v := range *d {
		res = append(res, v.Key)
	}
	return res
}
func (d *Keyed) Key(k string) (Val, error) {
	if d != nil {
		for _, v := range *d {
			if k == v.Key {
				return v.Val, nil
			}
		}
	}
	return Null{}, ErrKeyNotFound
}
func (d *Keyed) SetKey(k string, el Val) error {
	s := *d
	if el == nil { // if el is explicitly nil delete the value
		for i, v := range s {
			if k == v.Key {
				*d = append(s[:i], s[i+1:]...)
				return nil
			}
		}
		return nil
	}
	el = Unwrap(el)
	for i, v := range s {
		if k == v.Key {
			s[i].Val = el
			return nil
		}
	}
	*d = append(s, KeyVal{Key: k, Val: el})
	return nil
}
func (d *Keyed) IterKey(it func(string, Val) error) error {
	for _, k := range *d {
		if err := it(k.Key, k.Val); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}

type Dict struct {
	Typ typ.Type
	Keyed
}

func NewDict(el typ.Type, ks ...KeyVal) *Dict {
	return &Dict{typ.DictOf(el), ks}
}
func (d *Dict) Type() typ.Type {
	if d.Typ == typ.Void {
		d.Typ = typ.Dict
	}
	return d.Typ
}
func (d *Dict) Nil() bool  { return d == nil }
func (d *Dict) Mut() Mut   { return d }
func (d *Dict) Value() Val { return d }
func (d *Dict) As(t typ.Type) (Val, error) {
	if typ.Keyr.AssignableTo(t) {
		d.Typ = t
		return d, nil
	}
	if ok := d.Typ.ConvertibleTo(t); ok {
		neu := typ.ContEl(t)
		if ok {
			d.Typ = t
			return d, nil
		}
		for _, kv := range d.Keyed {
			if !kv.Type().ConvertibleTo(neu) {
				ok = false
				break
			}
		}
		if ok {
			d.Typ = t
			return d, nil
		}
	}
	// TODO obj type
	return nil, fmt.Errorf("cannot convert %T from %s to %s", d, d.Type(), t)
}
func (d *Dict) New() Mut         { return &Dict{d.Typ, nil} }
func (d *Dict) Ptr() interface{} { return d }
func (d *Dict) Key(k string) (Val, error) {
	if d != nil {
		return d.Keyed.Key(k)
	}
	return Null{}, ErrKeyNotFound
}

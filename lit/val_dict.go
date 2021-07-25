package lit

import (
	"bytes"
	"fmt"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

type KeyVal struct {
	Key string
	Val
}

type Dict struct {
	El    typ.Type
	Keyed []KeyVal
}

func (d *Dict) Type() typ.Type               { return typ.DictOf(d.El) }
func (d *Dict) Nil() bool                    { return d == nil }
func (d *Dict) Zero() bool                   { return len(d.Keyed) == 0 }
func (d *Dict) Value() Val                   { return d }
func (d *Dict) MarshalJSON() ([]byte, error) { return bfr.JSON(d) }
func (d *Dict) String() string               { return bfr.String(d) }
func (d *Dict) Print(p *bfr.P) (err error) {
	p.Byte('{')
	for i, v := range d.Keyed {
		if i > 0 {
			p.Sep()
		}
		p.RecordKey(v.Key)
		if err = v.Val.Print(p); err != nil {
			return err
		}
	}
	return p.Byte('}')
}

func (d *Dict) New() Mut         { return &Dict{d.El, nil} }
func (d *Dict) Ptr() interface{} { return d }
func (d *Dict) Assign(p Val) error {
	// TODO check types
	switch o := p.(type) {
	case nil:
		d.Keyed = nil
	case Null:
		d.Keyed = nil
	case Keyr:
		// TODO check types
		res := make([]KeyVal, 0, o.Len())
		o.IterKey(func(k string, v Val) error {
			res = append(res, KeyVal{k, Unwrap(v)})
			return nil
		})
		d.Keyed = res
	default:
		return ErrAssign
	}
	return nil
}

func (d *Dict) Len() int {
	return len(d.Keyed)
}
func (d *Dict) Idx(i int) (Val, error) {
	if i < 0 || i >= len(d.Keyed) {
		return nil, ErrIdxBounds
	}
	return d.Keyed[i].Val, nil
}
func (d *Dict) SetIdx(i int, el Val) error {
	if i < 0 {
		return ErrIdxBounds
	}
	if el == nil {
		el = Null{}
	}
	if d.El != typ.Void && d.El != typ.Data {
		// TODO check and convert el
	}
	if i >= len(d.Keyed) {
		if i < cap(d.Keyed) {
			n := d.Keyed[:i+1]
			for j := len(d.Keyed) - 1; j < i; j++ {
				n[j] = KeyVal{}
			}
		} else {
			n := make([]KeyVal, (i+1)*3/2)
			copy(n, d.Keyed)
			d.Keyed = n
		}
	}
	d.Keyed[i].Val = el
	return nil
}
func (d *Dict) IterIdx(it func(int, Val) error) error {
	for i, kv := range d.Keyed {
		if err := it(i, kv.Val); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (d *Dict) Keys() []string {
	if len(d.Keyed) == 0 {
		return nil
	}
	res := make([]string, 0, len(d.Keyed))
	for _, v := range d.Keyed {
		res = append(res, v.Key)
	}
	return res
}
func (d *Dict) Key(k string) (Val, error) {
	if d != nil {
		for _, v := range d.Keyed {
			if k == v.Key {
				return v.Val, nil
			}
		}
	}
	// TODO think about zero values of keyr go uses zero value and js undefined values
	return Null{}, nil
}
func (d *Dict) SetKey(k string, el Val) error {
	if el == nil {
		el = Null{}
	} else {
		el = Unwrap(el)
	}
	if d.El != typ.Void {
		// TODO check and convert el
	}
	for i, v := range d.Keyed {
		if k == v.Key {
			d.Keyed[i].Val = el
			return nil
		}
	}
	d.Keyed = append(d.Keyed, KeyVal{Key: k, Val: el})
	return nil
}
func (d *Dict) IterKey(it func(string, Val) error) error {
	for _, k := range d.Keyed {
		if err := it(k.Key, k.Val); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (d *Dict) UnmarshalJSON(b []byte) error {
	lit, err := Read(bytes.NewReader(b), "")
	if err != nil {
		return err
	}
	o, ok := lit.Val.(*Dict)
	if !ok {
		return fmt.Errorf("expect dict got %T", lit.Val)
	}
	*d = *o
	return nil
}

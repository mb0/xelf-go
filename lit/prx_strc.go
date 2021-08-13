package lit

import (
	"fmt"
	"log"
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type StrcPrx struct {
	proxy
	*params
}

func (x *StrcPrx) NewWith(v reflect.Value) (Mut, error) {
	return &StrcPrx{x.with(v), x.params}, nil
}
func (x *StrcPrx) New() (Mut, error) { return x.NewWith(x.new()) }

func (x *StrcPrx) Zero() bool {
	if x.Nil() {
		return true
	}
	e := x.elem()
	for _, idx := range x.idx {
		el, err := x.Reg.ProxyValue(e.FieldByIndex(idx).Addr())
		if err != nil {
			log.Printf("inconsistent struct proxy field: %v", err)
			return false
		}
		if !el.Zero() {
			return false
		}
	}
	return true
}
func (x *StrcPrx) Value() Val {
	if x.Nil() {
		return Null{}
	}
	return x
}
func (x *StrcPrx) Parse(a ast.Ast) error {
	if isNull(a) {
		return x.setNull()
	}
	if a.Kind != knd.Dict {
		return ast.ErrExpect(a, knd.Dict)
	}
	rv := x.elem()
	rv.Set(reflect.Zero(rv.Type()))
	for _, e := range a.Seq {
		key, val, err := ast.UnquotePair(e)
		if err != nil {
			return err
		}
		_, idx, _ := x.pkey(key)
		ev := rv.FieldByIndex(idx).Addr()
		el, err := x.Reg.ProxyValue(ev)
		if err != nil {
			return err
		}
		err = el.Parse(val)
		if err != nil {
			return err
		}
	}
	return nil
}
func (x *StrcPrx) Assign(v Val) (err error) {
	if v == nil || v.Nil() {
		return x.setNull()
	}
	switch o := v.Value().(type) {
	case Keyr:
		err = o.IterKey(func(k string, v Val) error {
			return x.SetKey(k, v)
		})
	case Idxr:
		err = o.IterIdx(func(i int, v Val) error {
			return x.SetIdx(i, v)
		})
	default:
		err = fmt.Errorf("%T %s to strc %v", v, v.Type(), ErrAssign)
	}
	return err
}
func (x *StrcPrx) String() string               { return bfr.String(x) }
func (x *StrcPrx) MarshalJSON() ([]byte, error) { return bfr.JSON(x) }
func (x *StrcPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *StrcPrx) Print(p *bfr.P) error {
	if x.Nil() {
		return p.Fmt("null")
	}
	e := x.elem()
	p.Byte('{')
	var n int
	for i, idx := range x.idx {
		addr := e.FieldByIndex(idx).Addr()
		el, err := x.Reg.ProxyValue(addr)
		if err != nil {
			return err
		}
		param := x.ps[i]
		if param.IsOpt() && el.Zero() {
			continue
		}
		if n++; n > 1 {
			p.Sep()
		}
		p.RecordKey(param.Key)
		err = el.Print(p)
		if err != nil {
			return err
		}
	}
	return p.Byte('}')
}
func (x *StrcPrx) Len() int {
	if x.Nil() {
		return 0
	}
	return len(x.ps)
}
func (x *StrcPrx) Idx(i int) (Val, error) {
	_, idx := x.pidx(i)
	if len(idx) == 0 {
		return nil, ErrIdxBounds
	}
	e := x.elem()
	el, err := x.Reg.ProxyValue(e.FieldByIndex(idx).Addr())
	if err != nil {
		return nil, err
	}
	return el, nil
}
func (x *StrcPrx) SetIdx(i int, v Val) error {
	_, idx := x.pidx(i)
	if len(idx) == 0 {
		return ErrIdxBounds
	}
	el, err := x.Reg.ProxyValue(x.elem().FieldByIndex(idx).Addr())
	if err != nil {
		return err
	}
	return el.Assign(v)
}
func (x *StrcPrx) IterIdx(it func(int, Val) error) error {
	if x.Nil() {
		return nil
	}
	e := x.elem()
	for i, idx := range x.idx {
		el, err := x.Reg.ProxyValue(e.FieldByIndex(idx).Addr())
		if err != nil {
			return err
		}
		if err = it(i, el); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (x *StrcPrx) Keys() []string {
	if x.Nil() {
		return nil
	}
	res := make([]string, 0, len(x.ps))
	for _, p := range x.ps {
		res = append(res, p.Key)
	}
	return res
}
func (x *StrcPrx) Key(k string) (Val, error) {
	if x.Nil() {
		return Null{}, nil
	}
	_, idx, _ := x.pkey(k)
	if len(idx) == 0 {
		return nil, fmt.Errorf("prx strc %T %q: %w", x.Ptr(), k, ErrKeyNotFound)
	}
	el, err := x.Reg.ProxyValue(x.elem().FieldByIndex(idx).Addr())
	if err != nil {
		return nil, err
	}
	return el, nil
}
func (x *StrcPrx) SetKey(k string, v Val) error {
	_, idx, _ := x.pkey(k)
	if len(idx) == 0 {
		return fmt.Errorf("prx strc %T %q: %w", x.Ptr(), k, ErrKeyNotFound)
	}
	el, err := x.Reg.ProxyValue(x.elem().FieldByIndex(idx).Addr())
	if err != nil {
		return err
	}
	return el.Assign(v)
}
func (x *StrcPrx) IterKey(it func(string, Val) error) error {
	if x.Nil() {
		return nil
	}
	e := x.elem()
	for i, idx := range x.idx {
		f := e.FieldByIndex(idx)
		el, err := x.Reg.ProxyValue(f.Addr())
		if err != nil {
			return err
		}
		p := x.ps[i]
		if err = it(p.Key, el); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (x *StrcPrx) pidx(i int) (p typ.Param, _ []int) {
	if i < 0 {
		return p, nil
	}
	if i >= len(x.ps) {
		return p, nil
	}
	return x.ps[i], x.idx[i]
}
func (x *StrcPrx) pkey(k string) (p typ.Param, _ []int, _ int) {
	for i, p := range x.ps {
		if p.Key == k {
			return p, x.idx[i], i
		}
	}
	return p, nil, -1
}

package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type ListPrx struct{ proxy }

func (x *ListPrx) NewWith(v reflect.Value) Mut { return &ListPrx{x.with(v)} }

func (x *ListPrx) New() Mut   { return x.NewWith(x.new()) }
func (x *ListPrx) Zero() bool { return x.Nil() || x.Len() == 0 }
func (x *ListPrx) Mut() Mut   { return x }
func (x *ListPrx) Value() Val {
	if x.Nil() && x.isptr() {
		return Null{}
	}
	return x
}
func (x *ListPrx) As(t typ.Type) (Val, error) {
	if x.typ == t {
		return x, nil
	}
	return &ListPrx{x.typed(t)}, nil
}
func (x *ListPrx) Parse(a ast.Ast) error {
	if isNull(a) {
		return x.setNull()
	}
	if a.Kind != knd.Idxr {
		return ast.ErrExpect(a, knd.Idxr)
	}
	rv := x.elem()
	nv := rv.Slice(0, 0)
	et := rv.Type().Elem()
	for _, e := range a.Seq {
		ev := reflect.New(et)
		val, err := x.Reg.ProxyValue(ev)
		if err != nil {
			return err
		}
		err = val.Parse(e)
		if err != nil {
			return err
		}
		nv = reflect.Append(nv, ev.Elem())
	}
	rv.Set(nv)
	return nil
}
func (x *ListPrx) Assign(v Val) (err error) {
	if v == nil || v.Nil() {
		return x.setNull()
	}
	e := x.elem()
	n := e.Slice(0, 0)
	switch o := Unwrap(v).(type) {
	case Idxr:
		err = o.IterIdx(func(i int, el Val) error {
			val, err := Conv(x.Reg, e.Type().Elem(), el)
			if err != nil {
				return err
			}
			n = reflect.Append(n, val)
			return nil
		})
	default:
		err = fmt.Errorf("%T to list %w", v, ErrAssign)
	}
	e.Set(n)
	return
}
func (x *ListPrx) Append(vs ...Val) (err error) {
	e := x.elem()
	et := e.Type().Elem()
	res := make([]reflect.Value, len(vs))
	for i, v := range vs {
		res[i], err = Conv(x.Reg, et, v)
		if err != nil {
			return err
		}
	}
	e.Set(reflect.Append(e, res...))
	return nil
}
func (x *ListPrx) String() string               { return bfr.String(x) }
func (x *ListPrx) MarshalJSON() ([]byte, error) { return bfr.JSON(x) }
func (x *ListPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *ListPrx) Print(p *bfr.P) error {
	if x.Nil() && x.isptr() {
		return p.Fmt("null")
	}
	p.Byte('[')
	e := x.elem()
	n := e.Len()
	if n > 0 {
		p.Indent()
		for i := 0; i < n; i++ {
			if i > 0 {
				p.Sep()
				p.Break()
			}
			el, err := x.Reg.ProxyValue(e.Index(i).Addr())
			if err != nil {
				return err
			}
			err = el.Print(p)
			if err != nil {
				return err
			}
		}
		p.Dedent()
	}
	return p.Byte(']')
}
func (x *ListPrx) Len() int {
	if x.Nil() {
		return 0
	}
	return x.elem().Len()
}
func (x *ListPrx) Idx(i int) (res Val, err error) {
	if i, err = checkIdx(i, x.Len()); err != nil {
		return
	}
	return x.Reg.ProxyValue(x.elem().Index(i).Addr())
}
func (x *ListPrx) SetIdx(i int, v Val) (err error) {
	if i, err = checkIdx(i, x.Len()); err != nil {
		return
	}
	el, err := x.Reg.ProxyValue(x.elem().Index(i).Addr())
	if err != nil {
		return err
	}
	return el.Assign(v)
}
func (x *ListPrx) IterIdx(it func(int, Val) error) error {
	if x.Nil() {
		return nil
	}
	e := x.elem()
	for i, n := 0, e.Len(); i < n; i++ {
		el, err := x.Reg.ProxyValue(e.Index(i).Addr())
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

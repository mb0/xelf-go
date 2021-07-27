package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
)

type ListPrx struct{ proxy }

func (x *ListPrx) NewWith(v reflect.Value) (Mut, error) { return &ListPrx{x.with(v)}, nil }
func (x *ListPrx) New() (Mut, error)                    { return x.NewWith(x.new()) }

func (x *ListPrx) Zero() bool { return x.Len() == 0 }
func (x *ListPrx) Value() Val { return x }
func (x *ListPrx) Parse(a ast.Ast) error {
	rv := x.Reflect()
	if isNull(a) {
		rv.Set(rv.Slice(0, 0))
		return nil
	}
	if a.Kind != knd.List {
		return fmt.Errorf("expect list")
	}
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
	e := x.Reflect()
	n := e.Slice(0, 0)
	switch o := v.(type) {
	case nil:
	case Null:
	case Idxr:
		err = o.IterIdx(func(i int, el Val) error {
			val, err := x.Reg.Conv(e.Type().Elem(), el)
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
func (x *ListPrx) Append(v Val) error {
	e := x.Reflect()
	val, err := x.Reg.Conv(e.Type().Elem(), v)
	if err != nil {
		return err
	}
	e.Set(reflect.Append(e, val))
	return nil
}
func (x *ListPrx) String() string               { return bfr.String(x) }
func (x *ListPrx) MarshalJSON() ([]byte, error) { return bfr.JSON(x) }
func (x *ListPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *ListPrx) Print(p *bfr.P) error {
	p.Byte('[')
	e := x.Reflect()
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
func (x *ListPrx) Len() int { return x.Reflect().Len() }
func (x *ListPrx) Idx(i int) (res Val, err error) {
	e := x.Reflect()
	if i, err = checkIdx(i, e.Len()); err != nil {
		return
	}
	return x.Reg.ProxyValue(e.Index(i).Addr())
}
func (x *ListPrx) SetIdx(i int, v Val) (err error) {
	e := x.Reflect()
	if i, err = checkIdx(i, e.Len()); err != nil {
		return
	}
	el, err := x.Reg.ProxyValue(e.Index(i).Addr())
	if err != nil {
		return err
	}
	return el.Assign(v)
}
func (x *ListPrx) IterIdx(it func(int, Val) error) error {
	e := x.Reflect()
	n := e.Len()
	for i := 0; i < n; i++ {
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

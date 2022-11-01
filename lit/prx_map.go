package lit

import (
	"reflect"
	"sort"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type MapPrx struct{ proxy }

func (x *MapPrx) NewWith(v reflect.Value) (Mut, error) { return &MapPrx{x.with(v)}, nil }
func (x *MapPrx) New() (Mut, error)                    { return x.NewWith(x.new()) }

func (x *MapPrx) Zero() bool { return x.Len() == 0 }
func (x *MapPrx) Value() Val {
	if x.Nil() {
		return Null{}
	}
	return x
}
func (x *MapPrx) Parse(_ typ.Reg, a ast.Ast) error {
	if isNull(a) {
		return x.setNull()
	}
	if a.Kind != knd.Dict {
		return ast.ErrExpect(a, knd.Dict)
	}
	// we need to delete existing map keys to be compatible with dict implementation
	rv := x.elem()
	clearMap(rv)
	et := rv.Type().Elem()
	for _, e := range a.Seq {
		key, val, err := ast.UnquotePair(e)
		if err != nil {
			return err
		}
		ev := reflect.New(et)
		el, err := x.Reg.ProxyValue(ev)
		if err != nil {
			return err
		}
		err = el.Parse(x.Reg, val)
		if err != nil {
			return err
		}
		rv.SetMapIndex(reflect.ValueOf(key), ev.Elem())
	}
	return nil
}
func (x *MapPrx) Assign(v Val) error {
	if v == nil || v.Nil() {
		return x.setNull()
	}
	rv := x.elem()
	clearMap(rv)
	switch o := v.Value().(type) {
	case Null:
	case Keyr:
		// TODO check type
		err := o.IterKey(func(k string, v Val) error {
			return x.SetKey(k, v)
		})
		if err != nil {
			return err
		}
	default:
		return ErrAssign
	}
	return nil
}
func (x *MapPrx) String() string               { return bfr.String(x) }
func (x *MapPrx) MarshalJSON() ([]byte, error) { return bfr.JSON(x) }
func (x *MapPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *MapPrx) Print(p *bfr.P) error {
	if x.Nil() && x.isptr() {
		return p.Fmt("null")
	}
	keys := x.Keys()
	sort.Strings(keys)
	e := x.elem()
	p.Byte('{')
	for i, k := range keys {
		if i > 0 {
			p.Sep()
		}
		p.RecordKey(k)
		el, err := x.entry(k, e.MapIndex(reflect.ValueOf(k)))
		if err != nil {
			return err
		}
		err = el.Print(p)
		if err != nil {
			return err
		}
	}
	return p.Byte('}')
}
func (x *MapPrx) Len() int {
	if x.Nil() {
		return 0
	}
	return x.elem().Len()
}
func (x *MapPrx) Keys() []string {
	if x.Nil() {
		return nil
	}
	e := x.elem()
	res := make([]string, 0, e.Len())
	iter := e.MapRange()
	for iter.Next() {
		res = append(res, iter.Key().String())
	}
	return res
}
func (x *MapPrx) Key(k string) (Val, error) {
	if x.Nil() {
		return Null{}, nil
	}
	rv := x.elem().MapIndex(reflect.ValueOf(k))
	if !rv.IsValid() {
		return Null{}, nil
	}
	return x.entry(k, rv)
}
func (x *MapPrx) SetKey(k string, v Val) (err error) {
	e := x.elem()
	var val reflect.Value
	if v != nil {
		val, err = Conv(x.Reg, e.Type().Elem(), v)
		if err != nil {
			return err
		}
	}
	e.SetMapIndex(reflect.ValueOf(k), val)
	return nil
}
func (x *MapPrx) IterKey(it func(string, Val) error) error {
	if x.Nil() {
		return nil
	}
	e := x.elem()
	for iter := e.MapRange(); iter.Next(); {
		key := iter.Key().String()
		prx, err := x.entry(key, iter.Value())
		if err != nil {
			return err
		}
		if err = it(key, prx); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (x *MapPrx) entry(key string, val reflect.Value) (Mut, error) {
	var special bool
	if special = val.Type().Kind() != reflect.Ptr; special {
		n := reflect.New(val.Type())
		n.Elem().Set(val)
		val = n
	}
	prx, err := x.Reg.ProxyValue(val)
	if err != nil {
		return nil, err
	}
	if special {
		prx = &proxyEntry{m: x, key: key, Mut: prx}
	}
	return prx, nil
}

type proxyEntry struct {
	m   *MapPrx
	key string
	Mut
}

func (x *proxyEntry) Assign(v Val) error {
	err := x.Mut.Assign(v)
	if err != nil {
		return err
	}
	return x.m.SetKey(x.key, x.Mut)
}

func clearMap(v reflect.Value) {
	if v.Len() > 0 {
		var zero reflect.Value
		for _, key := range v.MapKeys() {
			v.SetMapIndex(key, zero)
		}
	}
}

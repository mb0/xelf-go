package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
)

type MapPrx struct{ proxy }

func (x *MapPrx) NewWith(v reflect.Value) (Mut, error) { return &MapPrx{x.with(v)}, nil }
func (x *MapPrx) New() (Mut, error)                    { return x.NewWith(x.new()) }

func (x *MapPrx) Zero() bool { return x.Len() == 0 }
func (x *MapPrx) Value() Val { return x }
func (x *MapPrx) Parse(a ast.Ast) error {
	rv := x.Reflect()
	if isNull(a) {
		clearMap(rv)
		return nil
	}
	if a.Kind != knd.Dict {
		return fmt.Errorf("expect dict")
	}
	// we need to delete existing map keys to be compatible with dict implementation
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
		err = el.Parse(val)
		if err != nil {
			return err
		}
		rv.SetMapIndex(reflect.ValueOf(key), ev.Elem())
	}
	return nil
}
func (x *MapPrx) Assign(v Val) error {
	e := x.Reflect()
	var zero reflect.Value
	for _, k := range e.MapKeys() {
		e.SetMapIndex(k, zero)
	}
	switch o := v.(type) {
	case nil:
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
	e := x.Reflect()
	p.Byte('{')
	iter := e.MapRange()
	for i := 0; iter.Next(); i++ {
		if i > 0 {
			p.Sep()
		}
		key := iter.Key().String()
		p.RecordKey(key)
		el, err := x.entry(key, iter.Value())
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
func (x *MapPrx) Len() int { return x.Reflect().Len() }
func (x *MapPrx) Keys() []string {
	e := x.Reflect()
	res := make([]string, 0, e.Len())
	iter := e.MapRange()
	for iter.Next() {
		res = append(res, iter.Key().String())
	}
	return res
}
func (x *MapPrx) Key(k string) (Val, error) {
	e := x.Reflect()
	rv := e.MapIndex(reflect.ValueOf(k))
	return x.entry(k, rv)
}
func (x *MapPrx) SetKey(k string, v Val) error {
	e := x.Reflect()
	if e.IsNil() {
		e.Set(reflect.MakeMap(e.Type()))
	}
	val, err := x.Reg.Conv(e.Type().Elem(), v)
	if err != nil {
		return err
	}
	e.SetMapIndex(reflect.ValueOf(k), val)
	return nil
}
func (x *MapPrx) IterKey(it func(string, Val) error) error {
	e := x.Reflect()
	iter := e.MapRange()
	for iter.Next() {
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

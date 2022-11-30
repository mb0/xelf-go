package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

func optPrx(m Mut, ptr reflect.Value, opt, null bool) (Mut, error) {
	if opt {
		if ptr.Elem().IsNil() {
			el := reflect.New(ptr.Type().Elem().Elem())
			ptr.Elem().Set(el)
			m = el.Convert(reflect.TypeOf(m)).Interface().(Mut)
		}
		m = &OptPrx{m, ptr, null}
	}
	return m, nil
}

// OptPrx convers a corner case using primary mutable values to null pointers.
// If we proxy (*bool)(nil) we provide a wrapped *Bool as mutable and keep nil for null values.
// This narrow usecase lets us asume we have always a value backing value and ptr.
type OptPrx struct {
	typ.LitMut
	ptr  reflect.Value
	Null bool
}

func (o *OptPrx) Unwrap() Val {
	if o.Null {
		return Null{}
	}
	return o.LitMut
}
func (o *OptPrx) Type() typ.Type { return typ.Opt(o.LitMut.Type()) }
func (o *OptPrx) Nil() bool      { return o.Null || o.LitMut.Nil() }
func (o *OptPrx) Zero() bool     { return o.Null }
func (o *OptPrx) Mut() Mut       { return o }
func (o *OptPrx) Value() Val     { return o.Unwrap().Value() }
func (o *OptPrx) As(t typ.Type) (Val, error) {
	vt := o.LitMut.Type()
	if t == vt { // drop the wrapper
		if o.Null {
			return nil, fmt.Errorf("cannot convert null to %s", t)
		}
		return o.LitMut, nil
	}
	// TODO we want to keep the same ptr connection around through conversions.
	// TODO check for knd.None and be smarter not to rewrap.
	v, err := o.LitMut.As(t)
	if err != nil {
		o.LitMut = v.Mut()
	}
	return o, err
}

func (o *OptPrx) Print(p *bfr.P) error         { return o.Unwrap().Print(p) }
func (o *OptPrx) String() string               { return o.Unwrap().String() }
func (o *OptPrx) MarshalJSON() ([]byte, error) { return o.Unwrap().MarshalJSON() }
func (o *OptPrx) UnmarshalJSON(b []byte) error { return unmarshal(b, o) }

func (o *OptPrx) New() Mut {
	// we have no ptr anymore switch to typ.Wrap
	return Wrap(o.LitMut.New(), o.Type())
}
func (o *OptPrx) Parse(a ast.Ast) error {
	if isNull(a) {
		o.Null = true
	}
	return o.LitMut.Parse(a)
}
func (o *OptPrx) Assign(v Val) error {
	err := o.LitMut.Assign(v)
	if err != nil {
		return err
	}
	var el reflect.Value
	if o.Null = v == nil || v.Nil(); o.Null {
		el = reflect.New(o.ptr.Type().Elem()).Elem()
	} else {
		el = reflect.ValueOf(o.LitMut.Ptr()).Convert(o.ptr.Type().Elem())
	}
	o.ptr.Elem().Set(el)
	return err
}

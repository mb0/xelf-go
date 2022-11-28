package lit

import (
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

type OptMut struct {
	typ.LitMut
	ptr  *reflect.Value
	Null bool
}

func (o *OptMut) Unwrap() Val {
	if o.Nil() {
		return Null{}
	}
	return o.LitMut
}

func (o *OptMut) Type() typ.Type { return typ.Opt(o.LitMut.Type()) }
func (o *OptMut) Nil() bool      { return o == nil || o.Null }
func (o *OptMut) Zero() bool     { return o.Null || o.LitMut.Zero() }
func (o *OptMut) Mut() Mut       { return o }
func (o *OptMut) Value() Val     { return o.Unwrap().Value() }
func (o *OptMut) String() string { return o.Unwrap().String() }

func (o *OptMut) MarshalJSON() ([]byte, error) { return o.Unwrap().MarshalJSON() }
func (o *OptMut) UnmarshalJSON(b []byte) error { return unmarshal(b, o) }
func (o *OptMut) Print(p *bfr.P) error         { return o.Unwrap().Print(p) }

func (o *OptMut) New() Mut { return &OptMut{o.LitMut.New(), nil, true} }
func (o *OptMut) Parse(a ast.Ast) error {
	if isNull(a) {
		o.Null = true
	}
	return o.LitMut.Parse(a)
}
func (o *OptMut) Assign(v Val) error {
	switch v.(type) {
	case nil:
		o.Null = true
	case Null:
		o.Null = true
	default:
		o.Null = v.Nil()
	}
	if o.Null {
		v = Null{}
		if o.ptr != nil {
			o.ptr.Elem().Set(reflect.New(o.ptr.Type().Elem()).Elem())
		}
	}
	err := o.LitMut.Assign(v)
	if err != nil {
		return err
	}
	if !o.Null && o.ptr != nil {
		o.ptr.Elem().Set(reflect.ValueOf(o.LitMut.Ptr()).Convert(o.ptr.Type().Elem()))
	}
	return nil
}

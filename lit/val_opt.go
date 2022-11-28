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

func (o *OptMut) Type() typ.Type { return typ.Opt(o.LitMut.Type()) }
func (o *OptMut) Nil() bool      { return o.Null }
func (o *OptMut) Zero() bool     { return o.Null || o.LitMut.Zero() }
func (o *OptMut) Mut() Mut       { return o }
func (o *OptMut) Value() Val {
	if o.Null {
		return Null{}
	}
	return o.LitMut.Value()
}
func (o *OptMut) String() string {
	if o.Null {
		return "null"
	}
	return o.LitMut.String()
}
func (o *OptMut) MarshalJSON() ([]byte, error) {
	if o.Null {
		return []byte("null"), nil
	}
	return o.LitMut.MarshalJSON()
}
func (o *OptMut) UnmarshalJSON(b []byte) error { return unmarshal(b, o) }

func (o *OptMut) Print(p *bfr.P) error {
	if o.Null {
		return p.Fmt("null")
	}
	return o.LitMut.Print(p)
}
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

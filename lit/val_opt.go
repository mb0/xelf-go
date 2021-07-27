package lit

import (
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

type OptMut struct {
	Mut
	ptr  *reflect.Value
	null bool
}

func (o *OptMut) Type() typ.Type { return typ.Opt(o.Mut.Type()) }
func (o *OptMut) Nil() bool      { return o.null }
func (o *OptMut) Zero() bool     { return o.null || o.Mut.Zero() }
func (o *OptMut) Value() Val {
	if o.null {
		return Null{}
	}
	return o.Mut.Value()
}
func (o *OptMut) String() string {
	if o.null {
		return "null"
	}
	return o.Mut.String()
}
func (o *OptMut) MarshalJSON() ([]byte, error) {
	if o.null {
		return []byte("null"), nil
	}
	return o.Mut.MarshalJSON()
}
func (o *OptMut) UnmarshalJSON(b []byte) error { return unmarshal(b, o) }

func (o *OptMut) Print(p *bfr.P) error {
	if o.null {
		return p.Fmt("null")
	}
	return o.Mut.Print(p)
}
func (o *OptMut) New() (Mut, error) {
	mut, err := o.Mut.New()
	if err != nil {
		return nil, err
	}
	return &OptMut{mut, nil, true}, nil
}
func (o *OptMut) Parse(a ast.Ast) error {
	if isNull(a) {
		o.null = true
	}
	return o.Mut.Parse(a)
}
func (o *OptMut) Assign(v Val) error {
	switch v.(type) {
	case nil:
		o.null = true
	case Null:
		o.null = true
	default:
		o.null = v.Nil()
	}
	if o.null {
		v = Null{}
		if o.ptr != nil {
			o.ptr.Elem().Set(reflect.New(o.ptr.Type().Elem()).Elem())
		}
	}
	err := o.Mut.Assign(v)
	if err != nil {
		return err
	}
	if !o.null && o.ptr != nil {
		o.ptr.Elem().Set(reflect.ValueOf(o.Mut.Ptr()).Convert(o.ptr.Type().Elem()))
	}
	return nil
}

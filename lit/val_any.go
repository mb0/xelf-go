package lit

import (
	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

type AnyMut struct {
	Typ typ.Type
	val Val
}

func (o *AnyMut) Type() typ.Type {
	if o.val != nil {
		return o.val.Type()
	}
	return o.Typ
}
func (o *AnyMut) Unwrap() Val {
	if o == nil || o.val == nil {
		return Null{}
	}
	return o.val
}
func (o *AnyMut) Nil() bool  { return o == nil || o.val == nil || o.val.Nil() }
func (o *AnyMut) Zero() bool { return o == nil || o.val == nil || o.val.Zero() }
func (o *AnyMut) Mut() Mut   { return o }
func (o *AnyMut) Value() Val { return o.Unwrap().Value() }

func (o *AnyMut) String() string               { return o.Unwrap().String() }
func (o *AnyMut) MarshalJSON() ([]byte, error) { return o.Unwrap().MarshalJSON() }
func (o *AnyMut) UnmarshalJSON(b []byte) error { return unmarshal(b, o) }
func (o *AnyMut) Print(p *bfr.P) error         { return o.Unwrap().Print(p) }

func (o *AnyMut) New() Mut { return &AnyMut{Typ: o.Typ} }
func (o *AnyMut) Ptr() interface{} {
	if m, ok := o.val.(Mut); ok {
		return m.Ptr()
	}
	return o
}
func (o *AnyMut) Parse(a ast.Ast) (err error) {
	o.val, err = ParseMut(a)
	return err
}
func (o *AnyMut) Assign(v Val) error {
	switch v.(type) {
	case nil:
		o.val = nil
	case Null:
		o.val = nil
	default:
		o.val = v
	}
	return nil
}

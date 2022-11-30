package typ

import (
	"bytes"
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
)

// Wrap is versatile literal value wrapper. It can generalize the type of a mutable value.
// It does also provide automatic optional behaviour, a mutable 'any' value and lazy value support.
type Wrap struct {
	Typ Type
	OK  bool
	Val LitMut
}

func (w *Wrap) Unwrap() LitVal {
	if w.Val != nil {
		return w.Val
	}
	return Null{}
}
func (w *Wrap) Type() Type    { return w.Typ }
func (w *Wrap) Nil() bool     { return !w.OK || w.Val == nil || w.Val.Nil() }
func (w *Wrap) Zero() bool    { return !w.OK || w.Val == nil || w.Val.Zero() }
func (w *Wrap) Mut() LitMut   { return w }
func (w *Wrap) Value() LitVal { return w.Unwrap().Value() }
func (w Wrap) As(t Type) (LitVal, error) {
	// TODO check typ against wrap type first
	w.Typ = t
	if w.Val != nil {
		// TODO convert val?
		// TODO can we return val as-is?
	}
	return &w, nil
}
func (w *Wrap) Print(p *bfr.P) error         { return w.Unwrap().Print(p) }
func (w *Wrap) String() string               { return w.Unwrap().String() }
func (w *Wrap) MarshalJSON() ([]byte, error) { return w.Unwrap().MarshalJSON() }
func (w *Wrap) UnmarshalJSON(b []byte) error {
	a, err := ast.Read(bytes.NewReader(b), "")
	if err != nil {
		return err
	}
	return w.Parse(a)
}
func (w *Wrap) Parse(a ast.Ast) error {
	if w.Val == nil {
		w.Val = &AstVal{a}
	} else if err := w.Val.Parse(a); err != nil {
		return err
	}
	w.OK = !w.Val.Nil()
	return nil
}

func (w Wrap) New() (v LitMut) {
	if w.Val != nil {
		v = w.Val.New()
	}
	return &Wrap{Typ: w.Typ, Val: v}
}
func (w *Wrap) Ptr() interface{} {
	if w.Val != nil {
		return w.Val.Ptr()
	}
	return nil
}
func (w *Wrap) Assign(v LitVal) error {
	switch v.(type) {
	case nil:
		w.OK = false
		v = Null{}
	case Null:
		w.OK = false
	default:
		w.OK = !v.Nil()
	}
	if w.Val != nil {
		return w.Val.Assign(v)
	}
	if m, ok := v.(LitMut); ok {
		w.Val = m
	} else if v.Nil() {
		return nil // ignore null value
	}
	return fmt.Errorf("cannot assign immutable value to undefined wrapper, see lit.AnyWrap")
}

const null = "null"

var WrapNull = func(t Type) *Wrap { return &Wrap{Typ: t} }

type Null struct{}

func (Null) Type() Type                   { return None }
func (Null) Nil() bool                    { return true }
func (Null) Zero() bool                   { return true }
func (Null) Value() LitVal                { return Null{} }
func (Null) As(t Type) (LitVal, error)    { return WrapNull(t), nil }
func (Null) Mut() LitMut                  { return WrapNull(None) }
func (Null) String() string               { return null }
func (Null) Print(p *bfr.P) error         { return p.Fmt(null) }
func (Null) MarshalJSON() ([]byte, error) { return []byte(null), nil }
func (Null) Len() int                     { return 0 }

type AstVal struct{ ast.Ast }

func (v *AstVal) Nil() bool {
	return v.Kind <= knd.None || v.Kind == knd.Sym && v.Raw == null
}
func (v *AstVal) Type() Type    { return Any }
func (v *AstVal) Zero() bool    { return v.Nil() }
func (v *AstVal) Mut() LitMut   { return v }
func (v *AstVal) Value() LitVal { return v }
func (v *AstVal) As(t Type) (LitVal, error) {
	return v, fmt.Errorf("cannot redefine ast value type")
}
func (v *AstVal) Print(p *bfr.P) error         { return v.Ast.Print(p) }
func (v *AstVal) String() string               { return bfr.String(&v.Ast) }
func (v *AstVal) MarshalJSON() ([]byte, error) { return bfr.JSON(&v.Ast) }
func (v *AstVal) UnmarshalJSON(b []byte) (err error) {
	v.Ast, err = ast.Read(bytes.NewReader(b), "")
	return err
}
func (v *AstVal) New() LitMut           { return &AstVal{} }
func (v *AstVal) Ptr() interface{}      { return nil }
func (v *AstVal) Parse(a ast.Ast) error { v.Ast = a; return nil }
func (v *AstVal) Assign(o LitVal) error {
	return fmt.Errorf("cannot assign to lazy ast value")
}

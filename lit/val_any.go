package lit

import (
	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

func AnyWrap(t typ.Type) *typ.Wrap {
	return Wrap(&Any{}, t)
}

func Wrap(m Mut, t typ.Type) *typ.Wrap {
	return &typ.Wrap{Typ: t, Val: m}
}

type Any struct{ Val }

func (w *Any) Unwrap() Val {
	if w.Val == nil {
		return Null{}
	}
	return w.Val
}
func (w *Any) Type() typ.Type {
	if w.Val == nil || w.Val.Type() == typ.None {
		return typ.Any
	}
	return w.Unwrap().Type()
}
func (w *Any) Nil() bool  { return w.Val == nil || w.Val.Nil() }
func (w *Any) Zero() bool { return w.Val == nil || w.Val.Zero() }
func (w *Any) Mut() Mut   { return w }
func (w *Any) Value() Val { return w.Unwrap().Value() }
func (w *Any) As(t typ.Type) (Val, error) {
	if w.Val != nil {
		return w.Val.As(t)
	}
	return &typ.Wrap{Typ: t, Val: w}, nil
}

func (w *Any) Print(p *bfr.P) error         { return w.Unwrap().Print(p) }
func (w *Any) String() string               { return w.Unwrap().String() }
func (w *Any) MarshalJSON() ([]byte, error) { return w.Unwrap().MarshalJSON() }
func (w *Any) UnmarshalJSON(b []byte) error { return unmarshal(b, w) }

func (w *Any) New() Mut { return &Any{} }
func (w *Any) Ptr() interface{} {
	if m, ok := w.Val.(Mut); ok {
		return m.Ptr()
	}
	return nil
}
func (w *Any) Parse(a ast.Ast) (err error) {
	w.Val, err = ParseMut(a)
	return err
}
func (w *Any) Assign(v Val) error {
	switch v.(type) {
	case nil:
		w.Val = nil
	case Null:
		w.Val = nil
	default:
		w.Val = v
	}
	return nil
}

func init() { typ.WrapNull = AnyWrap }

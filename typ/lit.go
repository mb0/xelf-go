package typ

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
)

var (
	ErrAssign    = fmt.Errorf("cannot assign")
	ErrIdxBounds = fmt.Errorf("index out of bounds")
)

// LitVal is the common interface of all literal values see lit.Val for more information.
// This interface does in principle belong to the lit package.
type LitVal interface {
	Type() Type
	Nil() bool
	Zero() bool
	Value() LitVal
	String() string
	Print(*bfr.P) error
	MarshalJSON() ([]byte, error)
}

// LitMut is the common interface of all mutable literal values see lit.Mut for more information.
// Mutable values should have an UnmarshalJSON method unless the base type is natively supported.
// This interface does in principle belong to the lit package.
type LitMut interface {
	LitVal
	New() (LitMut, error)
	Ptr() interface{}
	Assign(LitVal) error
	Parse(ast.Ast) error
}

func (Type) Type() Type            { return Typ }
func (Type) Nil() bool             { return false }
func (t Type) Zero() bool          { return t == Void }
func (t Type) Value() LitVal       { return t }
func (*Type) New() (LitMut, error) { return new(Type), nil }
func (t *Type) Ptr() interface{}   { return t }
func (t *Type) Assign(p LitVal) error {
	if p == nil || p.Nil() {
		*t = Void
		return nil
	}
	if n, err := ToType(p); err != nil {
		return err
	} else {
		*t = n
	}
	return nil
}
func (t *Type) Parse(a ast.Ast) error {
	r, err := ParseAst(a)
	if err != nil {
		return err
	}
	*t = r
	return nil
}

func ToType(v LitVal) (t Type, err error) {
	switch v := v.(type) {
	case nil:
	case Type:
		t = v
	case *Type:
		if v != nil {
			t = *v
		}
	default:
		switch v := v.Value().(type) {
		case Type:
			t = v
		default:
			err = fmt.Errorf("not a type value %[1]T %[1]s", v)
		}
	}
	return
}

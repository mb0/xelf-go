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
	// Nil returns whether this is a null value.
	Nil() bool
	// Zero returns whether this is a zero value.
	Zero() bool
	// Value returns a simple value restricted to these types: Null, Bool, Int, Real, Str, Raw,
	// UUID, Time, Span, Type, Idxr, Keyr and *SpecRef.
	Value() LitVal
	// Mut returns the effective mutable itself or a new mutable for this value.
	Mut() LitMut
	// String returns a string content for char literals and xelf format for other literals.
	// Use bfr.String(v) to get quoted char literals.
	String() string
	// Print writes this literal to the given printer or returns an error.
	Print(*bfr.P) error
	// MarshalJSON returns the literal as json bytes
	MarshalJSON() ([]byte, error)
}

// LitMut is the common interface of all mutable literal values see lit.Mut for more information.
// Mutable values should have an UnmarshalJSON method unless the base type is natively supported.
// This interface does in principle belong to the lit package.
type LitMut interface {
	LitVal
	// New returns a fresh mutable value of the same type or an error.
	New() LitMut
	// Ptr returns a pointer to the underlying value for interfacing with other go tools.
	Ptr() interface{}
	// Assign assigns the given value to this mutable or returns an error.
	Assign(LitVal) error
	// Parse reads the given ast into this mutable or returns an error.
	// The registry parameter is strictly optional, proxies should bring a registry if required.
	Parse(ast.Ast) error
}

func (Type) Type() Type          { return Typ }
func (Type) Nil() bool           { return false }
func (t Type) Zero() bool        { return t == Void }
func (t Type) Value() LitVal     { return t }
func (t Type) Mut() LitMut       { return &t }
func (*Type) New() LitMut        { return new(Type) }
func (t *Type) Ptr() interface{} { return t }
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
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case Type:
		t = v
	case *Type:
		t = *v
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

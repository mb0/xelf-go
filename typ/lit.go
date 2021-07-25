package typ

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/bfr"
)

var ErrAssign = fmt.Errorf("cannot assign")

// LitVal is the common interface of all literal values.
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

// LitMut is the common interface of all mutable literal values.
// This interface does in principle belong to the lit package.
// Mutable values should have an UnmarshalJSON method unless the base type is natively supported.
type LitMut interface {
	LitVal
	New() LitMut
	Ptr() interface{}
	Assign(LitVal) error
}

func (Type) Type() Type          { return Typ }
func (Type) Nil() bool           { return false }
func (t Type) Zero() bool        { return t == Void }
func (t Type) Value() LitVal     { return t }
func (*Type) New() LitMut        { return new(Type) }
func (t *Type) Ptr() interface{} { return t }
func (t *Type) Assign(p LitVal) error {
	if n, err := ToType(p); err != nil {
		return err
	} else {
		*t = n
	}
	return nil
}

var ptrType = reflect.TypeOf((*Type)(nil))

func (t *Type) Reflect() reflect.Value { return reflect.ValueOf(t) }
func (*Type) NewWith(ptr reflect.Value) (LitMut, error) {
	t := ptr.Type()
	if t != ptrType {
		if !t.ConvertibleTo(ptrType) {
			return nil, fmt.Errorf("cannot convert %s to *typ.Type", t)
		}
		ptr = ptr.Convert(ptrType)
	}
	return ptr.Interface().(*Type), nil
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

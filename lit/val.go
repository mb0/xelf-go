package lit

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/typ"
)

var (
	ptrBool = reflect.TypeOf((*Bool)(nil))
	ptrInt  = reflect.TypeOf((*Int)(nil))
	ptrReal = reflect.TypeOf((*Real)(nil))
	ptrStr  = reflect.TypeOf((*Str)(nil))
	ptrRaw  = reflect.TypeOf((*Raw)(nil))
	ptrUUID = reflect.TypeOf((*UUID)(nil))
	ptrTime = reflect.TypeOf((*Time)(nil))
	ptrSpan = reflect.TypeOf((*Span)(nil))
)

type (
	Bool bool
	Int  int64
	Real float64
	Str  string
	Raw  []byte
	UUID [16]byte
	Time time.Time
	Span time.Duration
)

func (Bool) Type() typ.Type { return typ.Bool }
func (Int) Type() typ.Type  { return typ.Int }
func (Real) Type() typ.Type { return typ.Real }
func (Str) Type() typ.Type  { return typ.Str }
func (Raw) Type() typ.Type  { return typ.Raw }
func (UUID) Type() typ.Type { return typ.UUID }
func (Time) Type() typ.Type { return typ.Time }
func (Span) Type() typ.Type { return typ.Span }

func (v Bool) Nil() bool { return false }
func (i Int) Nil() bool  { return false }
func (r Real) Nil() bool { return false }
func (s Str) Nil() bool  { return false }
func (r Raw) Nil() bool  { return false }
func (u UUID) Nil() bool { return false }
func (t Time) Nil() bool { return false }
func (s Span) Nil() bool { return false }

func (v Bool) Zero() bool { return bool(!v) }
func (i Int) Zero() bool  { return i == 0 }
func (r Real) Zero() bool { return r == 0 }
func (s Str) Zero() bool  { return s == "" }
func (r Raw) Zero() bool  { return len(r) == 0 }
func (u UUID) Zero() bool { return u == UUID{} }
func (t Time) Zero() bool { return time.Time(t).IsZero() }
func (s Span) Zero() bool { return s == 0 }

func (v Bool) Value() Val { return v }
func (i Int) Value() Val  { return i }
func (r Real) Value() Val { return r }
func (s Str) Value() Val  { return s }
func (r Raw) Value() Val  { return r }
func (u UUID) Value() Val { return u }
func (t Time) Value() Val { return t }
func (s Span) Value() Val { return s }

func (v Bool) String() string { return strconv.FormatBool(bool(v)) }
func (i Int) String() string  { return fmt.Sprintf("%d", i) }
func (r Real) String() string { return fmt.Sprintf("%g", r) }
func (s Str) String() string  { return string(s) }
func (r Raw) String() string  { return cor.FormatRaw(r) }
func (u UUID) String() string { return cor.FormatUUID(u) }
func (t Time) String() string { return cor.FormatTime(time.Time(t)) }
func (s Span) String() string { return cor.FormatSpan(time.Duration(s)) }

func (v Bool) Print(p *bfr.P) error { return p.Fmt(v.String()) }
func (i Int) Print(p *bfr.P) error  { return p.Fmt(i.String()) }
func (r Real) Print(p *bfr.P) error { return p.Fmt(r.String()) }
func (s Str) Print(p *bfr.P) error  { return p.Quote(string(s)) }
func (r Raw) Print(p *bfr.P) error  { return p.Quote(r.String()) }
func (u UUID) Print(p *bfr.P) error { return p.Quote(u.String()) }
func (t Time) Print(p *bfr.P) error { return p.Quote(t.String()) }
func (s Span) Print(p *bfr.P) error { return p.Quote(s.String()) }

func (v Bool) MarshalJSON() ([]byte, error) { return []byte(v.String()), nil }
func (i Int) MarshalJSON() ([]byte, error)  { return []byte(i.String()), nil }
func (r Real) MarshalJSON() ([]byte, error) { return []byte(r.String()), nil }
func (s Str) MarshalJSON() ([]byte, error)  { return bfr.JSON(s) }
func (r Raw) MarshalJSON() ([]byte, error)  { return bfr.JSON(r) }
func (u UUID) MarshalJSON() ([]byte, error) { return bfr.JSON(u) }
func (t Time) MarshalJSON() ([]byte, error) { return bfr.JSON(t) }
func (s Span) MarshalJSON() ([]byte, error) { return bfr.JSON(s) }

func (*Bool) New() Mut { return new(Bool) }
func (*Int) New() Mut  { return new(Int) }
func (*Real) New() Mut { return new(Real) }
func (*Str) New() Mut  { return new(Str) }
func (*Raw) New() Mut  { return new(Raw) }
func (*UUID) New() Mut { return new(UUID) }
func (*Time) New() Mut { return new(Time) }
func (*Span) New() Mut { return new(Span) }

func (v *Bool) Ptr() interface{} { return v }
func (i *Int) Ptr() interface{}  { return i }
func (r *Real) Ptr() interface{} { return r }
func (s *Str) Ptr() interface{}  { return s }
func (r *Raw) Ptr() interface{}  { return r }
func (u *UUID) Ptr() interface{} { return u }
func (t *Time) Ptr() interface{} { return t }
func (s *Span) Ptr() interface{} { return s }

func (v *Bool) NewWith(ptr reflect.Value) (Mut, error) { return mustRef(ptrBool, ptr) }
func (i *Int) NewWith(ptr reflect.Value) (Mut, error)  { return mustRef(ptrInt, ptr) }
func (r *Real) NewWith(ptr reflect.Value) (Mut, error) { return mustRef(ptrReal, ptr) }
func (s *Str) NewWith(ptr reflect.Value) (Mut, error)  { return mustRef(ptrStr, ptr) }
func (r *Raw) NewWith(ptr reflect.Value) (Mut, error)  { return mustRef(ptrRaw, ptr) }
func (u *UUID) NewWith(ptr reflect.Value) (Mut, error) { return mustRef(ptrUUID, ptr) }
func (t *Time) NewWith(ptr reflect.Value) (Mut, error) { return mustRef(ptrTime, ptr) }
func (s *Span) NewWith(ptr reflect.Value) (Mut, error) { return mustRef(ptrSpan, ptr) }

func (v *Bool) Reflect() reflect.Value { return reflect.ValueOf(v).Elem() }
func (i *Int) Reflect() reflect.Value  { return reflect.ValueOf(i).Elem() }
func (r *Real) Reflect() reflect.Value { return reflect.ValueOf(r).Elem() }
func (s *Str) Reflect() reflect.Value  { return reflect.ValueOf(s).Elem() }
func (r *Raw) Reflect() reflect.Value  { return reflect.ValueOf(r).Elem() }
func (u *UUID) Reflect() reflect.Value { return reflect.ValueOf(u).Elem() }
func (t *Time) Reflect() reflect.Value { return reflect.ValueOf(t).Elem() }
func (s *Span) Reflect() reflect.Value { return reflect.ValueOf(s).Elem() }

func (v *Bool) Assign(p Val) error {
	if n, err := ToBool(p); err != nil {
		return err
	} else {
		*v = n
	}
	return nil
}

func (i *Int) Assign(p Val) error {
	if n, err := ToInt(p); err != nil {
		return err
	} else {
		*i = n
	}
	return nil
}

func (r *Real) Assign(p Val) error {
	if n, err := ToReal(p); err != nil {
		return err
	} else {
		*r = n
	}
	return nil
}

func (s *Str) Assign(p Val) error {
	if n, err := ToStr(p); err != nil {
		return err
	} else {
		*s = n
	}
	return nil
}
func (r *Raw) Assign(p Val) error {
	if n, err := ToRaw(p); err != nil {
		return err
	} else {
		*r = n
	}
	return nil
}
func (u *UUID) Assign(p Val) error {
	if n, err := ToUUID(p); err != nil {
		return err
	} else {
		*u = n
	}
	return nil
}

func (t *Time) Assign(p Val) error {
	if n, err := ToTime(p); err != nil {
		return err
	} else {
		*t = n
	}
	return nil
}

func (s *Span) Assign(p Val) error {
	if n, err := ToSpan(p); err != nil {
		return err
	} else {
		*s = n
	}
	return nil
}

func (t Time) Equal(o Time) bool { return time.Time(t).Equal(time.Time(o)) }
func (t Time) After(o Time) bool { return time.Time(t).After(time.Time(o)) }

func (s Span) Seconds() float64 { return time.Duration(s).Seconds() }

func mustRef(ref reflect.Type, v reflect.Value) (Mut, error) {
	t := v.Type()
	if t != ref {
		if !t.ConvertibleTo(ref) {
			return nil, fmt.Errorf("cannot %s convert to %s", v.Type(), ref)
		}
		v = v.Convert(ref)
	}
	return v.Interface().(Mut), nil
}

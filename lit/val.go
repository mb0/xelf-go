package lit

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type (
	Null = typ.Null
	Bool bool
	Num  float64
	Int  int64
	Real float64
	Char string
	Str  string
	Raw  []byte
	UUID [16]byte
	Time time.Time
	Span time.Duration

	BoolMut bool
	NumMut  float64
	IntMut  int64
	RealMut float64
	CharMut string
	StrMut  string
	RawMut  []byte
	UUIDMut [16]byte
	TimeMut time.Time
	SpanMut time.Duration
)

// NewUUID returns a new random lit uuid v4 value.
func NewUUID() UUID { return UUID(cor.NewUUID()) }

// Now returns a new lit time value truncated to millisecond precision.
func Now() Time                  { return Time(time.Now().Truncate(time.Millisecond)) }
func (t Time) Equal(o Time) bool { return time.Time(t).Equal(time.Time(o)) }
func (t Time) After(o Time) bool { return time.Time(t).After(time.Time(o)) }
func (s Span) Seconds() float64  { return time.Duration(s).Seconds() }

func (Bool) Type() typ.Type { return typ.Bool }
func (Num) Type() typ.Type  { return typ.Num }
func (Int) Type() typ.Type  { return typ.Int }
func (Real) Type() typ.Type { return typ.Real }
func (Char) Type() typ.Type { return typ.Char }
func (Str) Type() typ.Type  { return typ.Str }
func (Raw) Type() typ.Type  { return typ.Raw }
func (UUID) Type() typ.Type { return typ.UUID }
func (Time) Type() typ.Type { return typ.Time }
func (Span) Type() typ.Type { return typ.Span }

func (*BoolMut) Type() typ.Type { return typ.Bool }
func (*NumMut) Type() typ.Type  { return typ.Num }
func (*IntMut) Type() typ.Type  { return typ.Int }
func (*RealMut) Type() typ.Type { return typ.Real }
func (*CharMut) Type() typ.Type { return typ.Char }
func (*StrMut) Type() typ.Type  { return typ.Str }
func (*RawMut) Type() typ.Type  { return typ.Raw }
func (*UUIDMut) Type() typ.Type { return typ.UUID }
func (*TimeMut) Type() typ.Type { return typ.Time }
func (*SpanMut) Type() typ.Type { return typ.Span }

func (v Bool) Nil() bool { return false }
func (i Num) Nil() bool  { return false }
func (i Int) Nil() bool  { return false }
func (r Real) Nil() bool { return false }
func (s Char) Nil() bool { return false }
func (s Str) Nil() bool  { return false }
func (r Raw) Nil() bool  { return false }
func (u UUID) Nil() bool { return false }
func (t Time) Nil() bool { return false }
func (s Span) Nil() bool { return false }

func (v *BoolMut) Nil() bool { return v == nil }
func (i *NumMut) Nil() bool  { return i == nil }
func (i *IntMut) Nil() bool  { return i == nil }
func (r *RealMut) Nil() bool { return r == nil }
func (s *CharMut) Nil() bool { return s == nil }
func (s *StrMut) Nil() bool  { return s == nil }
func (r *RawMut) Nil() bool  { return r == nil }
func (u *UUIDMut) Nil() bool { return u == nil }
func (t *TimeMut) Nil() bool { return t == nil }
func (s *SpanMut) Nil() bool { return s == nil }

func (v Bool) Zero() bool { return bool(!v) }
func (i Num) Zero() bool  { return i == 0 }
func (i Int) Zero() bool  { return i == 0 }
func (r Real) Zero() bool { return r == 0 }
func (s Char) Zero() bool { return s == "" }
func (s Str) Zero() bool  { return s == "" }
func (r Raw) Zero() bool  { return len(r) == 0 }
func (u UUID) Zero() bool { return u == UUID{} }
func (t Time) Zero() bool { return time.Time(t).IsZero() }
func (s Span) Zero() bool { return s == 0 }

func (v *BoolMut) Zero() bool { return v == nil || bool(!*v) }
func (i *NumMut) Zero() bool  { return i == nil || *i == 0 }
func (i *IntMut) Zero() bool  { return i == nil || *i == 0 }
func (r *RealMut) Zero() bool { return r == nil || *r == 0 }
func (s *CharMut) Zero() bool { return s == nil || *s == "" }
func (s *StrMut) Zero() bool  { return s == nil || *s == "" }
func (r *RawMut) Zero() bool  { return r == nil || len(*r) == 0 }
func (u *UUIDMut) Zero() bool { return u == nil || *u == UUIDMut{} }
func (t *TimeMut) Zero() bool { return t == nil || time.Time(*t).IsZero() }
func (s *SpanMut) Zero() bool { return s == nil || *s == 0 }

func (v Bool) Value() Val { return v }
func (i Num) Value() Val  { return Real(i) }
func (i Int) Value() Val  { return i }
func (r Real) Value() Val { return r }
func (s Char) Value() Val { return Str(s) }
func (s Str) Value() Val  { return s }
func (r Raw) Value() Val  { return r }
func (u UUID) Value() Val { return u }
func (t Time) Value() Val { return t }
func (s Span) Value() Val { return s }

func (v *BoolMut) Value() Val { return Bool(*v) }
func (i *NumMut) Value() Val  { return Real(*i) }
func (i *IntMut) Value() Val  { return Int(*i) }
func (r *RealMut) Value() Val { return Real(*r) }
func (s *CharMut) Value() Val { return Str(*s) }
func (s *StrMut) Value() Val  { return Str(*s) }
func (r *RawMut) Value() Val  { return Raw(*r) }
func (u *UUIDMut) Value() Val { return UUID(*u) }
func (t *TimeMut) Value() Val { return Time(*t) }
func (s *SpanMut) Value() Val { return Span(*s) }

func (v Bool) As(t typ.Type) (Val, error) { return wrapPrim(v.Mut(), t, typ.Bool) }
func (i Num) As(t typ.Type) (Val, error)  { return wrapPrim(i.Mut(), t, typ.Num) }
func (i Int) As(t typ.Type) (Val, error)  { return wrapPrim(i.Mut(), t, typ.Int) }
func (r Real) As(t typ.Type) (Val, error) { return wrapPrim(r.Mut(), t, typ.Real) }
func (s Char) As(t typ.Type) (Val, error) { return wrapPrim(s.Mut(), t, typ.Char) }
func (s Str) As(t typ.Type) (Val, error)  { return wrapPrim(s.Mut(), t, typ.Str) }
func (r Raw) As(t typ.Type) (Val, error)  { return wrapPrim(r.Mut(), t, typ.Raw) }
func (u UUID) As(t typ.Type) (Val, error) { return wrapPrim(u.Mut(), t, typ.UUID) }
func (t Time) As(n typ.Type) (Val, error) { return wrapPrim(t.Mut(), n, typ.Time) }
func (s Span) As(t typ.Type) (Val, error) { return wrapPrim(s.Mut(), t, typ.Span) }

func (v *BoolMut) As(t typ.Type) (Val, error) { return wrapPrim(v, t, typ.Bool) }
func (i *NumMut) As(t typ.Type) (Val, error)  { return wrapPrim(i, t, typ.Num) }
func (i *IntMut) As(t typ.Type) (Val, error)  { return wrapPrim(i, t, typ.Int) }
func (r *RealMut) As(t typ.Type) (Val, error) { return wrapPrim(r, t, typ.Real) }
func (s *CharMut) As(t typ.Type) (Val, error) { return wrapPrim(s, t, typ.Char) }
func (s *StrMut) As(t typ.Type) (Val, error)  { return wrapPrim(s, t, typ.Str) }
func (r *RawMut) As(t typ.Type) (Val, error)  { return wrapPrim(r, t, typ.Raw) }
func (u *UUIDMut) As(t typ.Type) (Val, error) { return wrapPrim(u, t, typ.UUID) }
func (t *TimeMut) As(n typ.Type) (Val, error) { return wrapPrim(t, n, typ.Time) }
func (s *SpanMut) As(t typ.Type) (Val, error) { return wrapPrim(s, t, typ.Span) }

func (v Bool) Mut() Mut { return (*BoolMut)(&v) }
func (i Num) Mut() Mut  { return (*NumMut)(&i) }
func (i Int) Mut() Mut  { return (*IntMut)(&i) }
func (r Real) Mut() Mut { return (*RealMut)(&r) }
func (s Char) Mut() Mut { return (*CharMut)(&s) }
func (s Str) Mut() Mut  { return (*StrMut)(&s) }
func (r Raw) Mut() Mut  { return (*RawMut)(&r) }
func (u UUID) Mut() Mut { return (*UUIDMut)(&u) }
func (t Time) Mut() Mut { return (*TimeMut)(&t) }
func (s Span) Mut() Mut { return (*SpanMut)(&s) }

func (v *BoolMut) Mut() Mut { return v }
func (i *NumMut) Mut() Mut  { return i }
func (i *IntMut) Mut() Mut  { return i }
func (r *RealMut) Mut() Mut { return r }
func (s *CharMut) Mut() Mut { return s }
func (s *StrMut) Mut() Mut  { return s }
func (r *RawMut) Mut() Mut  { return r }
func (u *UUIDMut) Mut() Mut { return u }
func (t *TimeMut) Mut() Mut { return t }
func (s *SpanMut) Mut() Mut { return s }

func (s Char) Len() int { return len(s) }
func (s Str) Len() int  { return len(s) }
func (r Raw) Len() int  { return len(r) }

func (s *CharMut) Len() int { return len(*s) }
func (s *StrMut) Len() int  { return len(*s) }
func (r *RawMut) Len() int  { return len(*r) }

func (v Bool) String() string { return strconv.FormatBool(bool(v)) }
func (i Num) String() string  { return fmt.Sprintf("%g", i) }
func (i Int) String() string  { return fmt.Sprintf("%d", i) }
func (r Real) String() string { return fmt.Sprintf("%g", r) }
func (s Char) String() string { return string(s) }
func (s Str) String() string  { return string(s) }
func (r Raw) String() string  { return cor.FormatRaw(r) }
func (u UUID) String() string { return cor.FormatUUID(u) }
func (t Time) String() string { return cor.FormatTime(time.Time(t)) }
func (s Span) String() string { return cor.FormatSpan(time.Duration(s)) }

func (v *BoolMut) String() string { return strconv.FormatBool(bool(*v)) }
func (i *NumMut) String() string  { return fmt.Sprintf("%g", *i) }
func (i *IntMut) String() string  { return fmt.Sprintf("%d", *i) }
func (r *RealMut) String() string { return fmt.Sprintf("%g", *r) }
func (s *CharMut) String() string { return string(*s) }
func (s *StrMut) String() string  { return string(*s) }
func (r *RawMut) String() string  { return cor.FormatRaw(*r) }
func (u *UUIDMut) String() string { return cor.FormatUUID(*u) }
func (t *TimeMut) String() string { return cor.FormatTime(time.Time(*t)) }
func (s *SpanMut) String() string { return cor.FormatSpan(time.Duration(*s)) }

func (v Bool) Print(p *bfr.P) error { return p.Fmt(v.String()) }
func (i Num) Print(p *bfr.P) error  { return p.Fmt(i.String()) }
func (i Int) Print(p *bfr.P) error  { return p.Fmt(i.String()) }
func (r Real) Print(p *bfr.P) error { return p.Fmt(r.String()) }
func (s Char) Print(p *bfr.P) error { return p.Quote(string(s)) }
func (s Str) Print(p *bfr.P) error  { return p.Quote(string(s)) }
func (r Raw) Print(p *bfr.P) error  { return p.Quote(r.String()) }
func (u UUID) Print(p *bfr.P) error { return p.Quote(u.String()) }
func (t Time) Print(p *bfr.P) error { return p.Quote(t.String()) }
func (s Span) Print(p *bfr.P) error { return p.Quote(s.String()) }

func (v *BoolMut) Print(p *bfr.P) error { return p.Fmt(v.String()) }
func (i *NumMut) Print(p *bfr.P) error  { return p.Fmt(i.String()) }
func (i *IntMut) Print(p *bfr.P) error  { return p.Fmt(i.String()) }
func (r *RealMut) Print(p *bfr.P) error { return p.Fmt(r.String()) }
func (s *CharMut) Print(p *bfr.P) error { return p.Quote(string(*s)) }
func (s *StrMut) Print(p *bfr.P) error  { return p.Quote(string(*s)) }
func (r *RawMut) Print(p *bfr.P) error  { return p.Quote(r.String()) }
func (u *UUIDMut) Print(p *bfr.P) error { return p.Quote(u.String()) }
func (t *TimeMut) Print(p *bfr.P) error { return p.Quote(t.String()) }
func (s *SpanMut) Print(p *bfr.P) error { return p.Quote(s.String()) }

func (v Bool) MarshalJSON() ([]byte, error) { return []byte(v.String()), nil }
func (i Num) MarshalJSON() ([]byte, error)  { return []byte(i.String()), nil }
func (i Int) MarshalJSON() ([]byte, error)  { return []byte(i.String()), nil }
func (r Real) MarshalJSON() ([]byte, error) { return []byte(r.String()), nil }
func (s Char) MarshalJSON() ([]byte, error) { return bfr.JSON(s) }
func (s Str) MarshalJSON() ([]byte, error)  { return bfr.JSON(s) }
func (r Raw) MarshalJSON() ([]byte, error)  { return bfr.JSON(r) }
func (u UUID) MarshalJSON() ([]byte, error) { return bfr.JSON(u) }
func (t Time) MarshalJSON() ([]byte, error) { return bfr.JSON(t) }
func (s Span) MarshalJSON() ([]byte, error) { return bfr.JSON(s) }

func (v *BoolMut) MarshalJSON() ([]byte, error) { return []byte(v.String()), nil }
func (i *NumMut) MarshalJSON() ([]byte, error)  { return []byte(i.String()), nil }
func (i *IntMut) MarshalJSON() ([]byte, error)  { return []byte(i.String()), nil }
func (r *RealMut) MarshalJSON() ([]byte, error) { return []byte(r.String()), nil }
func (s *CharMut) MarshalJSON() ([]byte, error) { return bfr.JSON(s) }
func (s *StrMut) MarshalJSON() ([]byte, error)  { return bfr.JSON(s) }
func (r *RawMut) MarshalJSON() ([]byte, error)  { return bfr.JSON(r) }
func (u *UUIDMut) MarshalJSON() ([]byte, error) { return bfr.JSON(u) }
func (t *TimeMut) MarshalJSON() ([]byte, error) { return bfr.JSON(t) }
func (s *SpanMut) MarshalJSON() ([]byte, error) { return bfr.JSON(s) }

func (r *Raw) UnmarshalJSON(b []byte) error  { return unmarshal(b, (*RawMut)(r)) }
func (u *UUID) UnmarshalJSON(b []byte) error { return unmarshal(b, (*UUIDMut)(u)) }
func (t *Time) UnmarshalJSON(b []byte) error { return unmarshal(b, (*TimeMut)(t)) }
func (s *Span) UnmarshalJSON(b []byte) error { return unmarshal(b, (*SpanMut)(s)) }

func (r *RawMut) UnmarshalJSON(b []byte) error  { return unmarshal(b, r) }
func (u *UUIDMut) UnmarshalJSON(b []byte) error { return unmarshal(b, u) }
func (t *TimeMut) UnmarshalJSON(b []byte) error { return unmarshal(b, t) }
func (s *SpanMut) UnmarshalJSON(b []byte) error { return unmarshal(b, s) }

func (*BoolMut) New() Mut { return new(BoolMut) }
func (*NumMut) New() Mut  { return new(NumMut) }
func (*IntMut) New() Mut  { return new(IntMut) }
func (*RealMut) New() Mut { return new(RealMut) }
func (*CharMut) New() Mut { return new(CharMut) }
func (*StrMut) New() Mut  { return new(StrMut) }
func (*RawMut) New() Mut  { return new(RawMut) }
func (*UUIDMut) New() Mut { return new(UUIDMut) }
func (*TimeMut) New() Mut { return new(TimeMut) }
func (*SpanMut) New() Mut { return new(SpanMut) }

func (v *BoolMut) Ptr() interface{} { return v }
func (i *NumMut) Ptr() interface{}  { return i }
func (i *IntMut) Ptr() interface{}  { return i }
func (r *RealMut) Ptr() interface{} { return r }
func (s *CharMut) Ptr() interface{} { return s }
func (s *StrMut) Ptr() interface{}  { return s }
func (r *RawMut) Ptr() interface{}  { return r }
func (u *UUIDMut) Ptr() interface{} { return u }
func (t *TimeMut) Ptr() interface{} { return t }
func (s *SpanMut) Ptr() interface{} { return s }

func (v *BoolMut) Parse(a ast.Ast) error {
	if isNull(a) {
		*v = false
		return nil
	}
	if a.Kind != knd.Sym || (a.Raw != "false" && a.Raw != "true") {
		return ast.ErrInvalidBool(a)
	}
	*v = len(a.Raw) == 4
	return nil
}
func (i *NumMut) Parse(a ast.Ast) error {
	if isNull(a) {
		*i = 0
		return nil
	}
	if a.Kind != knd.Num {
		return ast.ErrExpect(a, knd.Num)
	}
	n, err := strconv.ParseFloat(a.Raw, 64)
	if err != nil {
		return ast.ErrInvalid(a, knd.Num, err)
	}
	*i = NumMut(n)
	return nil
}
func (i *IntMut) Parse(a ast.Ast) error {
	if isNull(a) {
		*i = 0
		return nil
	}
	if a.Kind != knd.Num {
		return ast.ErrExpect(a, knd.Int)
	}
	n, err := strconv.ParseInt(a.Raw, 10, 64)
	if err != nil {
		return ast.ErrInvalid(a, knd.Int, err)
	}
	*i = IntMut(n)
	return nil
}
func (r *RealMut) Parse(a ast.Ast) error {
	if isNull(a) {
		*r = 0
		return nil
	}
	if a.Kind != knd.Real && a.Kind != knd.Num {
		return ast.ErrExpect(a, knd.Num)
	}
	n, err := strconv.ParseFloat(a.Raw, 64)
	if err != nil {
		return ast.ErrInvalid(a, knd.Real, err)
	}
	*r = RealMut(n)
	return nil
}
func (s *CharMut) Parse(a ast.Ast) error {
	txt, err := unquoteStr(a)
	if err != nil {
		return err
	}
	*s = CharMut(txt)
	return nil
}
func (s *StrMut) Parse(a ast.Ast) error {
	txt, err := unquoteStr(a)
	if err != nil {
		return err
	}
	*s = StrMut(txt)
	return nil
}
func (r *RawMut) Parse(a ast.Ast) error {
	txt, err := unquoteStr(a)
	if err != nil {
		return err
	}
	n, err := cor.ParseRaw(txt)
	if err != nil {
		return err
	}
	*r = n
	return nil
}
func (u *UUIDMut) Parse(a ast.Ast) error {
	txt, err := unquoteStr(a)
	if err != nil {
		return err
	}
	if txt == "" {
		*u = UUIDMut{}
		return nil
	}
	n, err := cor.ParseUUID(txt)
	if err != nil {
		return ast.ErrInvalid(a, knd.UUID, err)
	}
	*u = n
	return nil
}
func (t *TimeMut) Parse(a ast.Ast) error {
	txt, err := unquoteStr(a)
	if err != nil {
		return err
	}
	if txt == "" {
		*t = TimeMut{}
		return nil
	}
	n, err := cor.ParseTime(txt)
	if err != nil {
		return ast.ErrInvalid(a, knd.Time, err)
	}
	*t = TimeMut(n)
	return nil
}
func (s *SpanMut) Parse(a ast.Ast) error {
	txt, err := unquoteStr(a)
	if err != nil {
		return err
	}
	if txt == "" {
		*s = 0
		return nil
	}
	n, err := cor.ParseSpan(txt)
	if err != nil {
		return ast.ErrInvalid(a, knd.Span, err)
	}
	*s = SpanMut(n)
	return nil
}

func (v *BoolMut) Assign(p Val) error {
	if n, err := ToBool(p); err != nil {
		return err
	} else {
		*v = BoolMut(n)
	}
	return nil
}

func (i *NumMut) Assign(p Val) error {
	if n, err := ToReal(p); err != nil {
		return err
	} else {
		*i = NumMut(n)
	}
	return nil
}
func (i *IntMut) Assign(p Val) error {
	if n, err := ToInt(p); err != nil {
		return err
	} else {
		*i = IntMut(n)
	}
	return nil
}

func (r *RealMut) Assign(p Val) error {
	if n, err := ToReal(p); err != nil {
		return err
	} else {
		*r = RealMut(n)
	}
	return nil
}

func (s *CharMut) Assign(p Val) error {
	if n, err := ToStr(p); err != nil {
		return err
	} else {
		*s = CharMut(n)
	}
	return nil
}
func (s *StrMut) Assign(p Val) error {
	if n, err := ToStr(p); err != nil {
		return err
	} else {
		*s = StrMut(n)
	}
	return nil
}
func (r *RawMut) Assign(p Val) error {
	if n, err := ToRaw(p); err != nil {
		return err
	} else {
		*r = RawMut(n)
	}
	return nil
}
func (u *UUIDMut) Assign(p Val) error {
	if n, err := ToUUID(p); err != nil {
		return err
	} else {
		*u = UUIDMut(n)
	}
	return nil
}

func (t *TimeMut) Assign(p Val) error {
	if n, err := ToTime(p); err != nil {
		return err
	} else {
		*t = TimeMut(n)
	}
	return nil
}

func (s *SpanMut) Assign(p Val) error {
	if n, err := ToSpan(p); err != nil {
		return err
	} else {
		*s = SpanMut(n)
	}
	return nil
}

func wrapPrim(v Mut, t, o typ.Type) (Val, error) {
	if o == t {
		return v, nil
	}
	if o.AssignableTo(t) {
		return Wrap(v, t), nil
	}
	if o.ConvertibleTo(t) {
		n := ZeroWrap(t)
		if v.Zero() {
			return n, nil
		}
		return n, n.Assign(v)
	}
	return nil, fmt.Errorf("cannot convert %T from %s to %s", v, v.Type(), t)
}

func isNull(a ast.Ast) bool { return a.Kind == knd.Sym && a.Raw == "null" }
func unquoteStr(a ast.Ast) (string, error) {
	if isNull(a) {
		return "", nil
	}
	if a.Kind != knd.Char {
		return "", ast.ErrExpect(a, knd.Char)
	}
	str, err := cor.Unquote(a.Raw)
	if err != nil {
		return "", ast.ErrInvalid(a, knd.Char, err)
	}
	return str, nil
}
func unmarshal(b []byte, m Mut) error {
	a, err := ast.Read(bytes.NewReader(b), "")
	if err != nil {
		return err
	}
	return m.Parse(a)
}

var (
	ptrBoolMut = reflect.TypeOf((*BoolMut)(nil))
	ptrNumMut  = reflect.TypeOf((*NumMut)(nil))
	ptrIntMut  = reflect.TypeOf((*IntMut)(nil))
	ptrRealMut = reflect.TypeOf((*RealMut)(nil))
	ptrCharMut = reflect.TypeOf((*CharMut)(nil))
	ptrStrMut  = reflect.TypeOf((*StrMut)(nil))
	ptrRawMut  = reflect.TypeOf((*RawMut)(nil))
	ptrUUIDMut = reflect.TypeOf((*UUIDMut)(nil))
	ptrTimeMut = reflect.TypeOf((*TimeMut)(nil))
	ptrSpanMut = reflect.TypeOf((*SpanMut)(nil))
)

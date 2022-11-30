package lit

import (
	"fmt"
	"time"

	"xelf.org/xelf/cor"
)

type Wrapper interface{ Unwrap() Val }

func Unwrap(v Val) Val {
	for w, ok := v.(Wrapper); ok; w, ok = v.(Wrapper) {
		v = w.Unwrap()
	}
	if v == nil {
		v = Null{}
	}
	return v
}

func ToBool(v Val) (b Bool, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *BoolMut:
		b = Bool(*v)
	case Bool:
		b = v
	default:
		switch v := v.Value().(type) {
		case Null:
		case Bool:
			b = v
		default:
			err = fmt.Errorf("not a bool value %[1]T %[1]s", v)
		}
	}
	return
}

func ToInt(v Val) (n Int, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *NumMut:
		n = Int(*v)
	case *IntMut:
		n = Int(*v)
	case *RealMut:
		n = Int(*v)
	case Num:
		n = Int(v)
	case Int:
		n = v
	case Real:
		n = Int(v)
	default:
		switch v := v.Value().(type) {
		case Null:
		case Int:
			n = v
		case Real:
			n = Int(v)
		default:
			err = fmt.Errorf("not a num value %[1]T %[1]s", v)
		}
	}
	return
}

func ToReal(v Val) (n Real, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *NumMut:
		n = Real(*v)
	case *IntMut:
		n = Real(*v)
	case *RealMut:
		n = Real(*v)
	case Num:
		n = Real(v)
	case Int:
		n = Real(v)
	case Real:
		n = v
	default:
		switch v := v.Value().(type) {
		case Null:
		case Int:
			n = Real(v)
		case Real:
			n = v
		default:
			err = fmt.Errorf("not a num value %[1]T %[1]s", v)
		}
	}
	return
}

func ToStr(v Val) (s Str, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *CharMut:
		s = Str(*v)
	case *StrMut:
		s = Str(*v)
	case Char:
		s = Str(v)
	case Str:
		s = v
	default:
		switch v := v.Value().(type) {
		case Null:
		case Str:
			s = v
		case Raw:
			s = Str(cor.FormatRaw(v))
		case UUID:
			s = Str(cor.FormatUUID(v))
		case Time:
			s = Str(cor.FormatTime(time.Time(v)))
		case Span:
			s = Str(cor.FormatSpan(time.Duration(v)))
		default:
			err = fmt.Errorf("not a char value %[1]T %[1]s", v)
		}
	}
	return
}

func ToRaw(v Val) (r Raw, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *RawMut:
		r = Raw(*v)
	case Raw:
		r = v
	default:
		switch v := v.Value().(type) {
		case Null:
		case Raw:
			r = v
		case Str:
			n, err := cor.ParseRaw(string(v))
			return Raw(n), err
		case UUID:
			r = Raw(cor.FormatUUID(v))
		case Time:
			r = Raw(cor.FormatTime(time.Time(v)))
		case Span:
			r = Raw(cor.FormatSpan(time.Duration(v)))
		default:
			err = fmt.Errorf("not a raw value %[1]T %[1]s", v)
		}
	}
	return
}

func ToUUID(v Val) (u UUID, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *UUID:
		u = *v
	case UUID:
		u = v
	default:
		switch v := v.Value().(type) {
		case Null:
		case UUID:
			u = v
		case Str:
			n, err := cor.ParseUUID(string(v))
			return UUID(n), err
		case Raw:
			n, err := cor.ParseUUID(string(v))
			return UUID(n), err
		default:
			err = fmt.Errorf("not a uuid value %[1]T %[1]s", v)
		}
	}
	return
}

func ToTime(v Val) (t Time, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *Time:
		t = *v
	case Time:
		t = v
	default:
		switch v := v.Value().(type) {
		case Null:
		case Time:
			t = v
		case Str:
			n, err := cor.ParseTime(string(v))
			return Time(n), err
		default:
			err = fmt.Errorf("not a time value %[1]T %[1]s", v)
		}
	}
	return
}

func ToSpan(v Val) (s Span, err error) {
	if v == nil || v.Nil() {
		return
	}
	switch v := v.(type) {
	case *Span:
		s = *v
	case Span:
		s = v
	default:
		switch v := v.Value().(type) {
		case Null:
		case Span:
			s = v
		case Str:
			n, err := cor.ParseSpan(string(v))
			return Span(n), err
		default:
			err = fmt.Errorf("not a span value %[1]T %[1]s", v)
		}
	}
	return
}

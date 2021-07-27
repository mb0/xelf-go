package lit

import (
	"fmt"
	"time"

	"xelf.org/xelf/cor"
)

func Unwrap(v Val) Val {
	switch p := v.(type) {
	case *AnyPrx:
		if !p.Nil() {
			return p.val
		}
	case *OptMut:
		if !p.Nil() {
			return p.Mut
		}
	}
	return v
}

func ToBool(v Val) (b Bool, err error) {
	switch v := v.(type) {
	case nil:
	case Null:
	case Bool:
		b = v
	case *Bool:
		if v != nil {
			b = *v
		}
	default:
		switch v := v.Value().(type) {
		case Bool:
			b = v
		default:
			err = fmt.Errorf("not a bool value %[1]T %[1]s", v)
		}
	}
	return
}

func ToInt(v Val) (n Int, err error) {
	switch v := v.(type) {
	case nil:
	case Null:
	case Int:
		n = v
	case Real:
		n = Int(v)
	case *Int:
		if v != nil {
			n = *v
		}
	case *Real:
		if v != nil {
			n = Int(*v)
		}
	default:
		switch v := v.Value().(type) {
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
	switch v := v.(type) {
	case nil:
	case Null:
	case Int:
		n = Real(v)
	case Real:
		n = v
	case *Int:
		if v != nil {
			n = Real(*v)
		}
	case *Real:
		if v != nil {
			n = *v
		}
	default:
		switch v := v.Value().(type) {
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
	switch v := v.(type) {
	case nil:
	case Null:
	case Str:
		s = v
	case *Str:
		if v != nil {
			s = *v
		}
	default:
		switch v := v.Value().(type) {
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
	switch v := v.(type) {
	case nil:
	case Null:
	case Raw:
		r = v
	case *Raw:
		if v != nil {
			r = *v
		}
	default:
		switch v := v.Value().(type) {
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
	switch v := v.(type) {
	case nil:
	case Null:
	case UUID:
		u = v
	case *UUID:
		if v != nil {
			u = *v
		}
	default:
		switch v := v.Value().(type) {
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
	switch v := v.(type) {
	case nil:
	case Null:
	case Time:
		t = v
	case *Time:
		if v != nil {
			t = *v
		}
	default:
		switch v := v.Value().(type) {
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
	switch v := v.(type) {
	case nil:
	case Null:
	case Span:
		s = v
	case *Span:
		if v != nil {
			s = *v
		}
	default:
		switch v := v.Value().(type) {
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

package lit

import (
	"fmt"

	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

func Zero(t typ.Type) (Mut, error) {
	if t.Kind&knd.Typ != 0 {
		t = typ.El(t)
		return &t, nil
	}
	if t.Kind&knd.Bool != 0 {
		return new(Bool), nil
	}
	if t.Kind&knd.Num == knd.Real {
		return new(Real), nil
	}
	if t.Kind&knd.Num != 0 {
		return new(Int), nil
	}
	if t.Kind&knd.Char == knd.Char {
		return new(Str), nil
	}
	if t.Kind&knd.Str != 0 {
		return new(Str), nil
	}
	if t.Kind&knd.Raw != 0 {
		return new(Raw), nil
	}
	if t.Kind&knd.UUID != 0 {
		return new(UUID), nil
	}
	if t.Kind&knd.Time != 0 {
		return new(Time), nil
	}
	if t.Kind&knd.Span != 0 {
		return new(Span), nil
	}
	if t.Kind&knd.List != 0 {
		return &List{El: typ.El(t)}, nil
	}
	if t.Kind&knd.Dict != 0 {
		return &Dict{El: typ.El(t)}, nil
	}
	if t.Kind&(knd.Rec|knd.Obj) != 0 {
		return NewStrc(t)
	}
	return nil, fmt.Errorf("no zero value for %s", t)
}

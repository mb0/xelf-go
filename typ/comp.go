package typ

import "xelf.org/xelf/knd"

const kndVal = knd.Any | knd.Exp

// AssignableTo returns whether *all* values represented by type t can be assigned to dst.
func (t Type) AssignableTo(dst Type) bool {
	if !t.Kind.IsAlt() || t.Kind&knd.Any == knd.Any {
		return t.assignableTo(dst)
	}
	if t.always(dst) || t.Kind&knd.None != 0 && dst.Kind&knd.None != 0 {
		return true
	}
	for _, tt := range altTypes(t) {
		if !tt.assignableTo(dst) {
			return false
		}
	}
	return true
}

// ConvertibleTo returns whether *any* value represented by type t can be assigned to dst.
// That means char is convertible to time, but str is not.
func (t Type) ConvertibleTo(dst Type) bool {
	if !t.Kind.IsAlt() || t.Kind&knd.Any == knd.Any {
		return t.convertibleTo(dst)
	}
	if t.always(dst) || t.Kind&knd.None != 0 && dst.Kind&knd.None != 0 {
		return true
	}
	for _, tt := range altTypes(t) {
		if tt.convertibleTo(dst) {
			return true
		}
	}
	return false
}

func (t Type) always(dst Type) bool {
	return t.ID > 0 && t.ID == dst.ID ||
		dst.Kind&knd.Var != 0 && dst.Kind&kndVal == knd.Void ||
		dst.Kind&knd.Any == knd.None && t.Kind&knd.None != 0
}

func (t Type) assignableTo(dst Type) bool {
	if t.always(dst) {
		return true
	}
	sk := t.Kind & kndVal
	if sk == knd.Void {
		return false
	}
	switch db := dst.Body.(type) {
	case nil:
		return dst.Kind&sk == sk
	case *Type:
		if dst.Kind&sk == sk {
			return elem(t).AssignableTo(*db)
		}
	case *AltBody:
		if dst.Kind&sk == sk {
			return true
		}
		for _, da := range db.Alts {
			if t.assignableTo(da) {
				return true
			}
		}
	case *ParamBody:
		if k := sk &^ knd.None; dst.Kind&k != k &&
			(k&knd.Spec == 0 || dst.Kind&knd.Spec == 0) &&
			(k&knd.Obj == 0 || dst.Kind&knd.Obj == 0) {
			return false
		}
		sb, ok := t.Body.(*ParamBody)
		if !ok {
			return false
		}
		for _, dp := range db.Params {
			si := sb.FindKeyIndex(dp.Key)
			if si >= 0 {
				if !sb.Params[si].Type.AssignableTo(dp.Type) {
					return false
				}
			} else if !dp.IsOpt() {
				return false
			}
		}
		return true
	case *ConstBody:
		if dst.Kind&sk == sk {
			_, ok := t.Body.(*ConstBody)
			return ok && dst.Ref == t.Ref
		}
		// we can assign constant names and values
		if sk == knd.Str || sk == knd.Int {
			return true
		}
	}
	return false
}

func (t Type) convertibleTo(dst Type) bool {
	if t.always(dst) || t.Kind&knd.Var != 0 && t.Kind&kndVal == knd.Void {
		return true
	}
	sk := t.Kind & kndVal
	if sk == knd.Void {
		return false
	}
	switch db := dst.Body.(type) {
	case nil:
		return dst.Kind&sk != 0
	case *Type:
		if k := sk &^ knd.None; dst.Kind&k != 0 {
			return elem(t).ConvertibleTo(*db)
		}
	case *AltBody:
		if dst.Kind&sk == sk {
			return true
		}
		for _, da := range db.Alts {
			if t.ConvertibleTo(da) {
				return true
			}
		}
	case *ParamBody:
		if k := sk &^ knd.None; dst.Kind&k != k &&
			(k&knd.Spec == 0 || dst.Kind&knd.Spec == 0) &&
			(k&knd.Obj == 0 || dst.Kind&knd.Obj == 0) {
			return false
		}
		sb, ok := t.Body.(*ParamBody)
		if !ok {
			return false
		}
		for _, dp := range db.Params {
			si := sb.FindKeyIndex(dp.Key)
			if si >= 0 {
				if !sb.Params[si].Type.ConvertibleTo(dp.Type) {
					return false
				}
			} else if !dp.IsOpt() {
				return false
			}
		}
		return true
	case *ConstBody:
		if dst.Kind&sk != sk {
			_, ok := t.Body.(*ConstBody)
			return ok && dst.Ref == t.Ref
		}
		// we can assign constant names and values
		if sk == knd.Str {
			return true
		}
		if sk == knd.Int {
			return true
		}
	}
	return false
}

// ResolvableTo returns whether the resolved value of t is convertible to the resolved dest.
// That call|char, call or exp are all possibly resolvable to time, but not call|str.
func (t Type) ResolvableTo(dst Type) bool {
	var tids, dstids []int32
	t, tids = unwrapExp(t)
	dst, dstids = unwrapExp(dst)
	if idMatch(tids, dstids) {
		return true
	}
	return t.ConvertibleTo(dst)
}

func elem(t Type) Type {
	if el := El(t); el != Void {
		return el
	}
	if k := t.Kind & knd.All; k == knd.Form || k == knd.Func {
		if pb, _ := t.Body.(*ParamBody); pb != nil && len(pb.Params) > 0 {
			return pb.Params[len(pb.Params)-1].Type
		}
	}
	return Any
}

func unwrapExp(t Type) (_ Type, ids []int32) {
	for t.Kind&knd.Exp != 0 {
		if t.ID > 0 {
			ids = append(ids, t.ID)
		}
		t = elem(t)
	}
	return t, ids
}

func idMatch(a, b []int32) bool {
	if len(a) < len(b) {
		a, b = b, a
	}
	if len(b) == 0 {
		return false
	}
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}

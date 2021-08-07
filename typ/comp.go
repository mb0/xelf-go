package typ

import "xelf.org/xelf/knd"

const kndVal = knd.All | knd.Exp

// AssignableTo returns whether *all* values represented by type t can be assigned to dst.
func (t Type) AssignableTo(dst Type) bool {
	if t.ID > 0 && t.ID == dst.ID {
		return true
	}
	if dst.Kind&knd.Var != 0 && dst.Kind&kndVal == knd.Void {
		return true
	}
	sk := t.Kind & kndVal
	if sk == knd.Void {
		return false
	}
	switch db := dst.Body.(type) {
	case nil:
		return dst.Kind&sk == sk
	case *ElBody:
		if dst.Kind&sk == sk {
			return elem(t).AssignableTo(db.El)
		}
	case *SelBody:
		return db.EqualHist(t.Body, nil)
	case *RefBody:
		sb, ok := t.Body.(*RefBody)
		return ok && db.Ref == sb.Ref
	case *AltBody:
		if dst.Kind&sk == sk {
			return true
		}
		for _, da := range db.Alts {
			if t.AssignableTo(da) {
				return true
			}
		}
	case *ParamBody:
		if dst.Kind&sk != sk &&
			(sk&knd.Spec == 0 || dst.Kind&knd.Spec == 0) &&
			(sk&knd.Strc == 0 || dst.Kind&knd.Strc == 0) {
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
		if dst.Kind&sk != sk {
			sb, ok := t.Body.(*ConstBody)
			return ok && sb.Name == db.Name
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

// ConvertibleTo returns whether *any* value represented by type t can be assigned to dst.
// That means str is convertible to time, even though only a subset of strings can successfully
// converted to a time value.
func (t Type) ConvertibleTo(dst Type) bool {
	if t.ID > 0 && t.ID == dst.ID {
		return true
	}
	if dst.Kind&knd.Var != 0 && dst.Kind&kndVal == knd.Void {
		return true
	}
	if t.Kind&knd.Var != 0 && t.Kind&kndVal == knd.Void {
		return true
	}
	sk := t.Kind & kndVal
	if sk == knd.Void {
		return false
	}
	switch db := dst.Body.(type) {
	case nil:
		return dst.Kind&sk != 0
	case *ElBody:
		if dst.Kind&sk != 0 {
			return elem(t).ConvertibleTo(db.El)
		}
	case *SelBody:
		return db.EqualHist(t.Body, nil)
	case *RefBody:
		sb, ok := t.Body.(*RefBody)
		return ok && db.Ref == sb.Ref
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
		if dst.Kind&sk != sk &&
			(sk&knd.Spec == 0 || dst.Kind&knd.Spec == 0) &&
			(sk&knd.Strc == 0 || dst.Kind&knd.Strc == 0) {
			return false
		}
		sb, ok := t.Body.(*ParamBody)
		if !ok {
			return false
		}
		for _, sp := range sb.Params {
			di := db.FindKeyIndex(sp.Key)
			if di >= 0 {
				if !sp.Type.ConvertibleTo(db.Params[di].Type) {
					return false
				}
			}
		}
		return true
	case *ConstBody:
		if dst.Kind&sk != sk {
			sb, ok := t.Body.(*ConstBody)
			return ok && sb.Name == db.Name
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

func expr(t Type) (knd.Kind, Type) {
	k := t.Kind & knd.Exp
	if k == knd.Void {
		return knd.Lit, t
	}
	return k, elem(t)
}

func elem(t Type) Type {
	if t.Body == nil {
		return Any
	}
	b, ok := t.Body.(*ElBody)
	if !ok {
		return Void
	}
	return b.El
}

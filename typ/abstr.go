package typ

import "xelf.org/xelf/knd"

// Abstract returns a more generic type for the given value type.
// When we select literal values from generic typed containers we get the literal value type. e.g.
// for a dict|any {now:'2021-08-19'} selecting '.now' produces a str and not a char value. We need
// the char type so we infer its real type as time.
func Abstract(v Type) Type {
	switch k := v.Kind & knd.Prim; k {
	case knd.Str:
		return Char
	case knd.Int:
		return Num
	}
	return v
}

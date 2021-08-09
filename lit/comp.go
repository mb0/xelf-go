package lit

import (
	"fmt"
	"log"

	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// log ignored equal errors for a while until were certain everthing is fine
const logEqual = true

// Equal returns whtether two literal values are structurally equal.
// Value types and implementations are not compared.
func Equal(x, y Val) bool {
	if x == nil || y == nil {
		return y == x
	}
	if x.Nil() {
		return y.Nil()
	}
	if y.Nil() {
		return false
	}
	if x.Zero() {
		return y.Zero()
	}
	if y.Zero() {
		return false
	}
	k := x.Type().Kind & knd.All
	switch {
	case k&knd.Data == knd.Data:
		xv, yv := x.Value(), y
		if y.Type().Kind&knd.Data == knd.Data {
			yv = y.Value()
		}
		return Equal(xv, yv)
	case k == knd.Typ:
		a, b, err := typePair(x, y)
		if err != nil {
			if logEqual {
				log.Printf("equal type err: %v", err)
			}
			return false
		}
		return a.Equal(b)
	case k&knd.Spec != 0:
		return x == y
	case k == knd.Bool:
		a, b, err := boolPair(x, y)
		if err != nil {
			if logEqual {
				log.Printf("equal bool err: %v", err)
			}
			return false
		}
		return a == b
	case k&knd.Num != 0:
		a, b, err := realPair(x, y)
		if err != nil {
			if logEqual {
				log.Printf("equal num err: %v", err)
			}
			return false
		}
		return a == b
	case k == knd.Span:
		a, b, err := spanPair(x, y)
		if err != nil {
			if logEqual {
				log.Printf("equal span err: %v", err)
			}
			return false
		}
		return a == b
	case k == knd.Time:
		a, b, err := timePair(x, y)
		if err != nil {
			if logEqual {
				log.Printf("equal time err: %v", err)
			}
			return false
		}
		return a.Equal(b)
	case k&knd.Char != 0:
		a, b, err := strPair(x, y)
		if err != nil {
			if logEqual {
				log.Printf("equal time err: %v", err)
			}
			return false
		}
		return a == b
	case k&knd.Keyr != 0:
		xk, ok := x.(Keyr)
		if !ok {
			if logEqual {
				log.Printf("expect keyr got %T", x)
			}
			return false
		}
		yk, ok := y.(Keyr)
		if !ok || xk.Len() != yk.Len() {
			return false
		}
		err := xk.IterKey(func(key string, xv Val) error {
			yv, err := yk.Key(key)
			if err != nil {
				return err
			}
			if ok = Equal(xv, yv); !ok {
				return BreakIter
			}
			return nil
		})
		if err != nil {
			if logEqual {
				log.Printf("equal keyr err: %v", err)
			}
			return false
		}
		return ok
	case k&knd.Idxr != 0:
		xi, ok := x.(Idxr)
		if !ok {
			if logEqual {
				log.Printf("expect idxr got %T", x)
			}
			return false
		}
		yi, ok := y.(Idxr)
		if !ok || xi.Len() != yi.Len() {
			return false
		}
		err := xi.IterIdx(func(idx int, xv Val) error {
			yv, err := yi.Idx(idx)
			if err != nil {
				return err
			}
			if ok = Equal(xv, yv); !ok {
				return BreakIter
			}
			return nil
		})
		if err != nil {
			if logEqual {
				log.Printf("equal idxr err: %v", err)
			}
			return false
		}
		return ok
	}
	panic("cannot equal %[1]T %[1]s, %[2]T %[2]s")
}

func Compare(x, y Val) (int8, error) {
	k := x.Type().Kind & knd.Data
	if k == knd.Void {
		if y.Type().Kind != k {
			return -1, nil
		}
		return 0, nil
	}
	if k == knd.Bool {
		a, b, err := boolPair(x, y)
		if err != nil {
			return 0, err
		}
		if !a && b {
			return -1, nil
		}
		if a && !b {
			return 1, nil
		}
		return 0, nil
	}
	if k&knd.Num != 0 {
		a, b, err := realPair(x, y)
		if err != nil {
			return 0, err
		}
		if a < b {
			return -1, nil
		}
		if a > b {
			return 1, nil
		}
		return 0, nil
	}
	if k == knd.Span {
		a, b, err := spanPair(x, y)
		if err != nil {
			return 0, err
		}
		if a < b {
			return -1, nil
		}
		if a > b {
			return 1, nil
		}
		return 0, nil
	}
	if k == knd.Time {
		a, b, err := timePair(x, y)
		if err != nil {
			return 0, err
		}
		if b.After(a) {
			return -1, nil
		}
		if a.After(b) {
			return 1, nil
		}
		return 0, nil
	}
	if k&knd.Str != 0 {
		a, b, err := strPair(x, y)
		if err != nil {
			return 0, err
		}
		if a < b {
			return -1, nil
		}
		if a > b {
			return 1, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("cannot compare %[1]T %[1]s, %[2]T %[2]s", x, y)
}
func boolPair(x, y Val) (a, b Bool, err error) {
	if a, err = ToBool(x); err == nil {
		b, err = ToBool(y)
	}
	return
}
func realPair(x, y Val) (a, b Real, err error) {
	if a, err = ToReal(x); err == nil {
		b, err = ToReal(y)
	}
	return
}
func strPair(x, y Val) (a, b Str, err error) {
	if a, err = ToStr(x); err == nil {
		b, err = ToStr(y)
	}
	return
}
func spanPair(x, y Val) (a, b Span, err error) {
	if a, err = ToSpan(x); err == nil {
		b, err = ToSpan(y)
	}
	return
}
func timePair(x, y Val) (a, b Time, err error) {
	if a, err = ToTime(x); err == nil {
		b, err = ToTime(y)
	}
	return
}
func typePair(x, y Val) (a, b typ.Type, err error) {
	if a, err = typ.ToType(x); err == nil {
		b, err = typ.ToType(y)
	}
	return
}

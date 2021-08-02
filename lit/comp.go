package lit

import (
	"fmt"

	"xelf.org/xelf/knd"
)

func Equal(x, y Val) (bool, error) {
	k := x.Type().Kind & knd.Data
	if k == knd.Bool {
		a, b, err := boolPair(x, y)
		if err != nil {
			return false, err
		}
		return a == b, nil
	}
	if k&knd.Num != 0 {
		a, b, err := realPair(x, y)
		if err != nil {
			return false, err
		}
		return a == b, nil
	}
	if k == knd.Span {
		a, b, err := spanPair(x, y)
		if err != nil {
			return false, err
		}
		return a == b, nil
	}
	if k == knd.Time {
		a, b, err := timePair(x, y)
		if err != nil {
			return false, err
		}
		return a.Equal(b), nil
	}
	if k&knd.Char != 0 {
		a, b, err := strPair(x, y)
		if err != nil {
			return false, err
		}
		return a == b, nil
	}
	return false, fmt.Errorf("cannot equal %[1]T %[1]s, %[2]T %[2]s", x, y)
}

func Compare(x, y Val) (int8, error) {
	k := x.Type().Kind & knd.Data
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

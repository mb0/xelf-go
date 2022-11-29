package lit

import (
	"errors"

	"xelf.org/xelf/knd"
)

// Clone deep copies v or returns an error. It uses the same value implementations,
// but may turn none-mutable primitive values into a mutable representation.
func Clone(v Val) (Val, error) {
	return Edit(v, func(v Val) (Val, error) {
		n := v.Mut().New()
		return n, n.Assign(v)
	})
}

// SkipCont can be returned from EditFunc to skip over the contained values.
var SkipCont = errors.New("skip contained values")

// EditFunc edits and returns v, a new value, SkipCont, BreakIter or any other error.
type EditFunc func(v Val) (Val, error)

// Edit edits v and returns the result or an error.
// Function f is never called with nil.
func Edit(v Val, f EditFunc) (r Val, err error) {
	if v == nil || v.Nil() {
		return v, nil
	}
	if r, err = f(v); err != nil {
		if err == SkipCont {
			err = nil
		}
	} else if hasCont(r) {
		e := editor{Func: f, seen: map[Val]Val{v: r}}
		return r, e.editCont(r)
	}
	return r, err
}

type editor struct {
	Func EditFunc
	seen map[Val]Val
}

func (e *editor) editVal(v Val) (r Val, err error) {
	if v == nil || v.Nil() {
		return v, nil
	}
	if r, ok := e.seen[v]; ok {
		return r, SkipCont
	}
	if r, err = e.Func(v); err != nil {
		if err == SkipCont {
			e.seen[v] = r
			err = nil
		}
	} else {
		e.seen[v] = r
		if hasCont(r) {
			err = e.editCont(r)
		}
	}
	return r, err
}
func (e *editor) editCont(v Val) error {
	switch a := Unwrap(v).(type) {
	case Idxr:
		return a.IterIdx(func(idx int, el Val) error {
			r, err := e.editVal(el)
			if err != nil || r == el {
				return err
			}
			return a.SetIdx(idx, r)
		})
	case Keyr:
		return a.IterKey(func(key string, el Val) error {
			r, err := e.editVal(el)
			if err != nil || r == el {
				return err
			}
			return a.SetKey(key, r)
		})
	}
	return nil
}
func hasCont(v Val) bool { return !v.Zero() && v.Type().Kind&(knd.Keyr|knd.Idxr) != 0 }

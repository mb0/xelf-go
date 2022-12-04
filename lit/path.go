package lit

import (
	"fmt"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// Select reads path and returns the selected value from val or an error.
func Select(val Val, path string) (Val, error) {
	p, err := cor.ParsePath(path)
	if err != nil {
		return nil, err
	}
	return SelectPath(val, p)
}

// SelectPath returns the selected value at path from val or an error.
func SelectPath(val Val, path cor.Path) (_ Val, err error) {
	for i, s := range path {
		if val.Type().Kind&knd.Typ != 0 {
			t, err := typ.ToType(val)
			if err != nil {
				return nil, err
			}
			return typ.SelectPath(t, path[i:])
		}
		if s.Sep() == '/' {
			val, err = SelectList(val, path[i:])
		} else if s.IsIdx() {
			val, err = SelectIdx(val, s.Idx)
		} else if s.Key != "" {
			val, err = SelectKey(val, s.Key)
		}
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

func SelectKey(val Val, key string) (Val, error) {
	v := Unwrap(val)
	if a, ok := v.(Keyr); ok {
		return a.Key(key)
	}
	if v == (Null{}) {
		return v, nil
	}
	return nil, fmt.Errorf("key segment %q expects keyr got %[2]T %[2]s", key, v)
}

func SelectIdx(val Val, idx int) (res Val, err error) {
	v := Unwrap(val)
	if a, ok := v.(Idxr); ok {
		return a.Idx(idx)
	}
	if v == (Null{}) {
		return v, nil
	}
	return nil, fmt.Errorf("idx segment %d expects idxr got %[2]T %[2]s", idx, v)
}

func SelectList(val Val, path cor.Path) (Val, error) {
	v := Unwrap(val)
	if a, ok := v.(Idxr); ok {
		res := &List{Typ: typ.List, Vals: make([]Val, 0, a.Len())}
		return res, a.IterIdx(func(_ int, v Val) (err error) {
			return collectIdxrVal(v, path, res)
		})
	}
	if v == (Null{}) {
		return v, nil
	}
	return nil, fmt.Errorf("list select %s expects idxr got %[2]T %[2]s", path, v)
}

func collectIdxrVal(v Val, path cor.Path, into *List) (err error) {

	if s := path.Fst(); s.IsIdx() {
		v, err = SelectIdx(v, s.Idx)
	} else if s.Key != "" {
		v, err = SelectKey(v, s.Key)
	}
	if err == nil && len(path) > 1 {
		v, err = SelectPath(v, path[1:])
	}
	if err != nil {
		return err
	}
	el := typ.ContEl(into.Typ)
	if n := typ.Alt(el, v.Type()); el != n {
		into.Typ = typ.ListOf(n)
	}
	into.Vals = append(into.Vals, v)
	return nil
}

// AssignPath sets an element of root at path to val or returns an error.
// It fails on missing intermediate container values.
func AssignPath(mut Mut, path cor.Path, val Val) (err error) {
	var root Val = mut
	for _, s := range path {
		var next Val
		if s.IsIdx() {
			next, err = SelectIdx(root, s.Idx)
		} else if s.Key != "" {
			next, err = SelectKey(root, s.Key)
		} else if len(path) == 1 {
			break
		} else {
			return fmt.Errorf("unexpected empty assign path")
		}
		if err != nil {
			return err
		}
		if next.Nil() {
			return fmt.Errorf("no value in %T at %s", s, root)
		}
		root = next
	}
	mut, ok := root.(Mut)
	if !ok {
		return fmt.Errorf("not a mutable value %T at %s", root, path)
	}
	return mut.Assign(val)
}

// CreatePath creates an element of root at path to val or returns an error.
// It resizes and creates missing intermediate container values using the registry.
func CreatePath(mut Mut, path cor.Path, val Val) (err error) {
	npath := path
	cur := mut
	for i, s := range path {
		var next Val
		if s.IsIdx() {
			next, err = SelectIdx(cur, s.Idx)
		} else if s.Key != "" {
			next, err = SelectKey(cur, s.Key)
		} else if len(path) == 1 {
			npath = path[i+1:]
			break
		} else {
			return fmt.Errorf("unexpected empty assign path")
		}
		if err != nil {
			break
		}
		nmut, ok := next.(Mut)
		if !ok {
			break
		}
		if nmut.Nil() {
			if x, ok := nmut.(*typ.Wrap); ok {
				x.OK = true
			} else {
				break
			}
		}
		cur, npath = nmut, path[i+1:]
	}
	if len(npath) == 0 {
		return cur.Assign(val)
	}
	s := npath[0]
	pt := cur.Type()
	et, err := typ.SelectPath(pt, cor.Path{s})
	if err != nil {
		return err
	}
	var ev Val
	isAny := et == typ.Void || et.Kind&knd.Data == knd.Data
	if isAny && len(npath) == 1 {
		ev = Unwrap(val)
	} else {
		if isAny {
			if npath[1].Key == "" {
				et = typ.Idxr
			} else {
				et = typ.Keyr
			}
		}
		z := Zero(typ.Deopt(et))
		if len(npath) > 1 {
			err = CreatePath(z, npath[1:], val)
		} else {
			err = z.Assign(val)
		}
		if err != nil {
			return err
		}
		ev = Unwrap(z)
	}
	if x, ok := cur.(*typ.Wrap); ok {
		x.OK = true
		cur = x.Val
	}
	if s.Key != "" {
		if a, ok := cur.(Keyr); ok {
			return a.SetKey(s.Key, ev)
		}
	} else {
		if a, ok := cur.(Idxr); ok {
			return a.SetIdx(s.Idx, ev)
		}
	}
	return fmt.Errorf("not an applicable value %T %s at %s", cur, cur, npath)
}

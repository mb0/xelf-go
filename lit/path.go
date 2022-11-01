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
		if s.Sel {
			val, err = SelectList(val, path[i:])
		} else if s.Key != "" {
			val, err = SelectKey(val, s.Key)
		} else {
			val, err = SelectIdx(val, s.Idx)
		}
		if err != nil {
			return nil, err
		}
	}
	return val, nil
}

func SelectKey(val Val, key string) (Val, error) {
	switch a := val.(type) {
	case Null:
		return a, nil
	case Keyr:
		return a.Key(key)
	default:
		if a, ok := val.Value().(Keyr); ok {
			return a.Key(key)
		}
	}
	return nil, fmt.Errorf("key segment %q expects keyr got %[2]T %[2]s", key, val)
}

func SelectIdx(val Val, idx int) (res Val, err error) {
	switch a := val.(type) {
	case Null:
		return a, nil
	case Idxr:
		return a.Idx(idx)
	default:
		if a, ok := val.Value().(Idxr); ok {
			return a.Idx(idx)
		}
	}
	return nil, fmt.Errorf("idx segment %d expects idxr got %[2]T %[2]s", idx, val)
}

func SelectList(val Val, path cor.Path) (_ Val, err error) {
	res := &List{}
	switch a := val.(type) {
	case *Vals:
		res.Vals = make([]Val, 0, len(*a))
		for _, v := range *a {
			if err = collectIdxrVal(v, path, res); err != nil {
				break
			}
		}
	case *List:
		res.Reg = a.Reg
		res.Vals = make([]Val, 0, len(a.Vals))
		for _, v := range a.Vals {
			if err = collectIdxrVal(v, path, res); err != nil {
				break
			}
		}
	case *Obj:
		res.Reg = a.Reg
		for _, v := range a.Vals {
			if err = collectIdxrVal(v, path, res); err != nil {
				break
			}
		}
	case *ListPrx:
		res.Reg = a.Reg
		err = collectIdxr(a, path, res)
	case *ObjPrx:
		res.Reg = a.Reg
		err = collectIdxr(a, path, res)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func collectIdxr(idxr Idxr, path cor.Path, into *List) error {
	into.Vals = make([]Val, 0, idxr.Len())
	return idxr.IterIdx(func(idx int, v Val) (err error) {
		return collectIdxrVal(v, path, into)
	})
}

func collectIdxrVal(v Val, path cor.Path, into *List) (err error) {
	if s := path[0]; s.Key != "" {
		v, err = SelectKey(v, s.Key)
	} else {
		v, err = SelectIdx(v, s.Idx)
	}
	if err == nil && len(path) > 1 {
		v, err = SelectPath(v, path[1:])
	}
	if err != nil {
		return err
	}
	into.El = typ.Alt(into.El, v.Type())
	into.Vals = append(into.Vals, v)
	return nil
}

// AssignPath sets an element of root at path to val or returns an error.
// It fails on missing intermediate container values.
func AssignPath(mut Mut, path cor.Path, val Val) (err error) {
	var root Val = mut
	for _, s := range path {
		var next Val
		if s.Key != "" {
			next, err = SelectKey(root, s.Key)
		} else {
			next, err = SelectIdx(root, s.Idx)
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
func CreatePath(reg *Reg, mut Mut, path cor.Path, val Val) (err error) {
	npath := path
	cur := mut
	for i, s := range path {
		var next Val
		if s.Key != "" {
			next, err = SelectKey(cur, s.Key)
		} else {
			next, err = SelectIdx(cur, s.Idx)
		}
		if err != nil {
			break
		}
		nmut, ok := next.(Mut)
		if !ok {
			break
		}
		if nmut.Nil() {
			x, ok := nmut.(*OptMut)
			if !ok {
				break
			}
			x.Null = false
		}
		cur, npath = nmut, path[i+1:]
	}
	if len(npath) == 0 {
		o, ok := cur.(*OptMut)
		if ok {
			o.Null = val.Nil()
			cur = o.Mut
		}
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
		ev = val.Value()
	} else {
		if isAny {
			if npath[1].Key == "" {
				et = typ.Idxr
			} else {
				et = typ.Keyr
			}
		}
		z, err := reg.Zero(et)
		if err != nil {
			return err
		}
		if len(npath) > 1 {
			err = CreatePath(reg, z, npath[1:], val)
		} else {
			err = z.Assign(val)
		}
		if err != nil {
			return err
		}
		ev = z.Value()
	}
	if o, ok := cur.(*OptMut); ok {
		o.Null = false
		cur = o.Mut
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

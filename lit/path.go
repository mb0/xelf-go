package lit

import (
	"fmt"

	"xelf.org/xelf/cor"
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
	for _, s := range path {
		if s.Key != "" {
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
	case Keyr:
		return a.Key(key)
	}
	return nil, fmt.Errorf("key segment %q expects keyr got %[2]T %[2]s", key, val)
}

func SelectIdx(val Val, idx int) (res Val, err error) {
	switch a := val.(type) {
	case Idxr:
		return a.Idx(idx)
	}
	return nil, fmt.Errorf("idx segment %d expects idxr got %[2]T %[2]s", idx, val)
}

// AssignPath sets an element of root at path to val or returns an error.
// It fails on missing intermediate container values.
func AssignPath(root Val, path cor.Path, val Val) (err error) {
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

// Create creates an element of root at path to val or returns an error.
// It resizes and creates missing intermediate container values using the registry.
func CreatePath(reg *Reg, root Val, path cor.Path, val Val) (err error) {
	npath := path
	for i, s := range path {
		var next Val
		if s.Key != "" {
			next, err = SelectKey(root, s.Key)
		} else {
			next, err = SelectIdx(root, s.Idx)
		}
		if err != nil {
			break
		}
		if next.Nil() {
			x, ok := next.(*OptMut)
			if !ok {
				break
			}
			x.null = false
		}
		root, npath = next, path[i+1:]
	}
	mut, ok := root.(Mut)
	if !ok {
		return fmt.Errorf("not a mutable value %T at %s", root, path)
	}
	if len(npath) == 0 {
		return mut.Assign(val)
	}
	s := npath[0]
	pt := mut.Type()
	et, err := typ.SelectPath(pt, cor.Path{s})
	if err != nil {
		return err
	}
	var ev Val
	if et == typ.Void && len(npath) == 1 {
		ev = val.Value()
	} else {
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
	o, ok := mut.(*OptMut)
	if ok {
		o.null = false
		mut = o.Mut
	}
	if s.Key != "" {
		if a, ok := mut.(Keyr); ok {
			return a.SetKey(s.Key, ev)
		}
	} else {
		if a, ok := mut.(Idxr); ok {
			return a.SetIdx(s.Idx, ev)
		}
	}
	return fmt.Errorf("not an applicable value %T at %s", mut, npath)
}
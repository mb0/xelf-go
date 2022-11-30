package typ

import (
	"fmt"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
)

// Select reads path and returns the selected type from t or an error.
func Select(t Type, path string) (Type, error) {
	p, err := cor.ParsePath(path)
	if err != nil {
		return Void, err
	}
	return SelectPath(t, p)
}

// SelectPath returns the selected type from t or an error.
func SelectPath(t Type, path cor.Path) (r Type, err error) {
	for i, s := range path {
		if s.Sel {
			r, err = SelectList(t, path[i:])
		} else if s.Key != "" {
			r, err = SelectKey(t, s.Key)
		} else {
			r, err = SelectIdx(t, s.Idx)
		}
		if err != nil {
			return Void, err
		}
		t = r
	}
	return t, nil
}

func SelectKey(t Type, key string) (Type, error) {
	if t.Kind&(knd.Keyr|knd.Spec) == 0 {
		return Void, fmt.Errorf("want keyr got %s", t)
	}
	switch b := t.Body.(type) {
	case *Type:
		return *b, nil
	case *ParamBody:
		if p := b.FindKeyIndex(cor.Keyed(key)); p >= 0 {
			return b.Params[p].Type, nil
		}
	}
	if t.Kind&knd.Dict == 0 {
		return Void, fmt.Errorf("key %s not found in %s", key, t)
	}
	return Any, nil
}

func SelectIdx(t Type, idx int) (Type, error) {
	if t.Kind&(knd.Idxr|knd.Spec) == 0 {
		return Void, fmt.Errorf("want idxr got %s", t)
	}
	switch b := t.Body.(type) {
	case *Type:
		return *b, nil
	case *ParamBody:
		i, l := idx, len(b.Params)
		if i < 0 {
			i = l + i
		}
		if i < 0 || i >= l {
			return Void, fmt.Errorf("idx %d %w [0:%d]", idx, ErrIdxBounds, l-1)
		}
		return b.Params[i].Type, nil
	}
	return Any, nil
}
func SelectList(t Type, p cor.Path) (r Type, err error) {
	if t.Kind&(knd.Idxr|knd.Spec) == 0 {
		return Void, fmt.Errorf("want idxr got %s", t)
	}
	switch b := t.Body.(type) {
	case nil:
		r = Any
	case *Type:
		r = *b
	case *ParamBody:
		l := len(b.Params)
		if t.Kind&knd.Spec != 0 {
			l--
		}
		for i := 0; i < l; i++ {
			r = Alt(r, b.Params[i].Type)
		}
	}
	if r == Any {
		return List, nil
	}
	if s := p[0]; s.Key != "" {
		r, err = SelectKey(r, s.Key)
	} else {
		r, err = SelectIdx(r, s.Idx)
	}
	if err == nil && len(p) > 1 {
		r, err = SelectPath(r, p[1:])
	}
	if err != nil {
		return Void, err
	}
	return ListOf(r), nil
}

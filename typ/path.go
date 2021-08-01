package typ

import (
	"fmt"

	"xelf.org/xelf/cor"
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
	for _, s := range path {
		if s.Key != "" {
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
	if !t.ConvertibleTo(Keyr) {
		return Void, fmt.Errorf("want keyr got %s", t)
	}
	switch b := t.Body.(type) {
	case *ElBody:
		return b.El, nil
	case *ParamBody:
		if p := b.FindKeyIndex(key); p >= 0 {
			return b.Params[p].Type, nil
		}
	}
	return Any, nil
}

func SelectIdx(t Type, idx int) (Type, error) {
	if !t.ConvertibleTo(Idxr) {
		return Void, fmt.Errorf("want idxr got %s", t)
	}
	switch b := t.Body.(type) {
	case *ElBody:
		return b.El, nil
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

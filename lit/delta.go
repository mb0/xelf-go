package lit

import (
	"strings"

	"xelf.org/xelf/cor"
)

// Delta is a list of path edits that describe a transformation from one value to another.
// Each element has an edit key and value. The key describes the edit path and the kind of edit.
// Special chars are path separators and dollar './$' in the whole path '+*-' in the suffix.
// If the key contains a dollar sign as segment, the edit value is a list of segment values before
// the any edit data. The minus suffix means we want to delete a dict key. One star or plus suffix
// means we have either a nested delta if the edit data is a keyr or list ops if the edit data is a
// list. If the first star or plus is followed by a star again, then it is a delta ops edit data.
type Delta Keyed

func (d Delta) String() string { return Keyed(d).String() }

// Diff returns the delta between values a and b or an error. The delta applied to a results in b.
// The simplest and correct answer is always to return b. We however do make some effort to find a
// simpler set of changes, but do not guarantee to return the shortest edit path.
func Diff(a, b Val) (Delta, error) { return diffVals(a, b, cor.Path{{Sel: 'n'}}, nil) }
func diffVals(a, b Val, pre cor.Path, d Delta) (Delta, error) {
	a, b = a.Value(), b.Value()
	switch aa := a.(type) {
	case Keyr:
		if bb, ok := b.(Keyr); ok {
			return diffKeyr(aa, bb, pre, d)
		}
	case Idxr:
		if bb, ok := b.(Idxr); ok {
			return diffIdxr(aa, bb, pre, d)
		}
	case Str:
		if bb, ok := b.(Str); ok {
			return diffStr(aa, bb, pre, d)
		}
	case Raw:
		if bb, ok := b.(Raw); ok {
			return diffRaw(aa, bb, pre, d)
		}
	default:
		if Equal(a, b) {
			return d, nil
		}
	}
	d = addEdit(d, pre, b, "")
	return d, nil
}

func addEdit(d Delta, p cor.Path, v Val, suf string) Delta {
	var vars Vals
	for i, s := range p {
		if isSafe(s) {
			continue
		}
		vars = append(vars, Str(s.Key))

		if s.Key = "$"; i == 0 && s.Sel == 0 {
			s.Sel = '.'
		}
		p[i] = s
	}
	kv := KeyVal{p.Suffix(suf), v}
	if len(vars) > 0 {
		vars = append(vars, v)
		kv.Val = &vars
	}
	return append(d, kv)
}
func isSafe(s cor.Seg) bool {
	return s.Key == "" || s.Key != "$" && !strings.ContainsAny(s.Key, "./ +-*\t\n")
}

func diffKeyr(a, b Keyr, pre cor.Path, d Delta) (Delta, error) {
	// we may want different behaviour for dicts and obj
	// dict keys can be deleted, obj keys only be set to zero
	// dict may be unordered while obj fields are ordered

	// lets first figure out dicts and then think about objs. start by getting all the keys
	ak, bk := a.Keys(), b.Keys()
	// the order does not matter so create a map of a's keys
	km := make(map[string]bool, len(ak))
	for _, k := range ak {
		km[k] = true
	}
	// now check b's keys against the map
	for _, k := range bk {
		if flag, ok := km[k]; !flag {
			if !ok {
				// does not exist in a
				v, err := b.Key(k)
				if err != nil {
					return nil, err
				}
				d = addEdit(d, addKeySeg(pre, k), v, "")
				// mark as handled
				km[k] = false
			} // duplicate key in b
			continue
		}
		// exists in a and b
		av, err := a.Key(k)
		if err != nil {
			return nil, err
		}
		bv, err := b.Key(k)
		if err != nil {
			return nil, err
		}
		// call delta on the values
		nvals, err := diffVals(av, bv, addKeySeg(pre, k), nil)
		if err != nil {
			return nil, err
		}
		// append edits and mark as handled
		d = append(d, nvals...)
		km[k] = false
	}
	for k, v := range km {
		if v { // deleted key
			d = addEdit(d, addKeySeg(pre, k), Null{}, "-")
		}
	}
	return d, nil
}

func emptyDot(p cor.Path) bool {
	return len(p) == 0 || len(p) == 1 && p[0].Empty()
}

func addKeySeg(p cor.Path, k string) cor.Path { return addSeg(p, cor.Seg{Key: k}) }
func addIdxSeg(p cor.Path, idx int) cor.Path  { return addSeg(p, cor.Seg{Sel: '.', Idx: idx}) }
func addSeg(p cor.Path, s cor.Seg) cor.Path {
	if emptyDot(p) {
		if s.Key == "$" && s.Sel == 0 {
			s.Sel = '.'
		}
		return cor.Path{s}
	}
	return append(p, s)
}

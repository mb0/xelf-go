package lit

import (
	"fmt"

	"github.com/mb0/diff"
)

// Delta returns a set of path edits that can be applied to a to arrive at b.
// The simple and correct answer is always to return b. We do make some effort to find a
// simpler set of changes, but make no guarantee to return the shortest edit path.
func Delta(a, b Val) ([]KeyVal, error) { return delta(a, b, ".", nil) }
func delta(a, b Val, pre string, vals []KeyVal) ([]KeyVal, error) {
	if aa, ok := a.(Keyr); ok {
		if bb, ok := b.(Keyr); ok {
			return deltaKeyr(aa, bb, pre, vals)
		}
	} else if aa, ok := toVals(a); ok {
		if bb, ok := toVals(b); ok {
			return deltaIdxr(a, b, aa, bb, pre, vals)
		}
	} else if Equal(a, b) {
		return vals, nil
	}
	return append(vals, KeyVal{Key: stripTailDot(pre), Val: b}), nil
}

func deltaIdxr(a, b Val, aa, bb []Val, pre string, vals []KeyVal) ([]KeyVal, error) {
	chgs := diff.Diff(len(aa), len(bb), &valsDiff{aa, bb})
	if len(chgs) == 0 {
		return vals, nil
	}
	// how much and how often we retain and delete from a and insert from b
	ops, t := diffToOps(chgs, aa, bb)
	if !t.changed() {
		return vals, nil
	} else if t.replaced() {
		return append(vals, KeyVal{Key: stripTailDot(pre), Val: b}), nil
	}
	// we have at least two ops and known at least one of them to be ret and one del or ins
	// ops of the same kind are merged and do not follow each other

	// we want to detect append and use special syntax. append does only occur when we have
	// two ops u,v where u is ret and v is ins
	if len(ops) == 2 && ops[0].N > 0 && ops[1].N == 0 {
		// lets return the special append op
		return append(vals, KeyVal{stripTailDot(pre) + "+", &List{Vals: ops[1].V}}), nil

	}
	// we also want to detect replacing a single element and use idx path notation. that does
	// only occur in two instances:

	// we have three ops u,v,w. retn is 1. v is del or ins and either u or w is the other
	// and we have four ops u,v,w,x. retn is 2. v and w are del and ins
	if len(ops) == 3 && t.retn == 1 || len(ops) == 4 && t.retn == 2 {
		u, v, w := ops[0], ops[1], ops[2]
		if v.N == -1 {
			if len(w.V) == 1 {
				return deltaSub(aa[u.N], w.V[0], pre, u.N, vals)
			}
			if t.retn == 1 && len(u.V) == 1 {
				return deltaSub(aa[0], u.V[0], pre, 0, vals)
			}
		} else if len(v.V) == 1 {
			if w.N == -1 {
				return deltaSub(aa[u.N], v.V[0], pre, u.N, vals)
			}
			if t.retn == 1 && u.N == -1 {
				return deltaSub(aa[0], v.V[0], pre, 0, vals)
			}
		}
	}
	// lets return the ops as list
	return append(vals, KeyVal{stripTailDot(pre) + "*", opsToList(ops)}), nil
}

func deltaSub(a, b Val, pre string, idx int, vals []KeyVal) ([]KeyVal, error) {
	path := fmt.Sprintf("%s%d.", pre, idx)
	return delta(a, b, path, vals)
}

func deltaKeyr(a, b Keyr, pre string, vals []KeyVal) ([]KeyVal, error) {
	// we may want different behaviour for dicts and strc
	// dict keys can be deleted, strc keys only be set to zero
	// dict may be unordered while strc fields are ordered

	// lets first figure out dicts and then think about strcs. start by getting all the keys
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
				path := k
				if pre != "." {
					path = pre + k
				}
				vals = append(vals, KeyVal{path, v})
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
		path := pre + k
		nvals, err := delta(av, bv, path+".", nil)
		if err != nil {
			return nil, err
		}
		if pre == "." {
			// check for simple path and turn them into plain keys
			for i, nv := range nvals {
				if nv.Key == path {
					nvals[i].Key = k
				}
			}
		}
		// append edits and mark as handled
		vals = append(vals, nvals...)
		km[k] = false
	}
	for k, v := range km {
		if v { // deleted key
			vals = append(vals, KeyVal{fmt.Sprintf("%s%s-", pre, k), Null{}})
		}
	}
	return vals, nil
}

func stripTailDot(s string) string {
	if len(s) > 1 && s[len(s)-1] == '.' {
		return s[:len(s)-1]
	}
	return s
}
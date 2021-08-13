package lit

import (
	"fmt"

	"github.com/mb0/diff"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
)

// Delta returns a set of path edits that can be applied to a to arrive at b.
// The simple and correct answer is always to return b. We do make some effort to find a
// simpler set of changes, but make no guarantee to return the shortest edit path.
func Delta(a, b Val) (map[string]Val, error) {
	d := make(map[string]Val)
	return d, delta(a, b, ".", d)
}
func delta(a, b Val, pre string, d map[string]Val) error {
	if aa, ok := a.(Keyr); ok {
		if bb, ok := b.(Keyr); ok {
			return deltaKeyr(aa, bb, pre, d)
		}
	} else if aa, ok := toVals(a); ok {
		if bb, ok := toVals(b); ok {
			return deltaIdxr(a, b, aa, bb, pre, d)
		}
	} else if Equal(a, b) {
		return nil
	}
	d[stripTailDot(pre)] = b
	return nil
}

// Apply applies edits d to mutable a or returns an error.
func Apply(reg *Reg, mut Mut, d map[string]Val) error {
	for k, v := range d {
		key := k
		if key != "" && key != "." && key[0] == '.' {
			lst := len(key) - 1
			if suf := key[lst]; suf == '+' {
				return applyListAppend(mut, key[:lst], v)
			} else if suf == '*' {
				return applyListOps(mut, key[:lst], v)
			} else if suf == '-' {
				return applyKeyrDel(mut, key[:lst])
			}
		}
		p, err := cor.ParsePath(key)
		if err != nil {
			return err
		}
		err = CreatePath(reg, mut, p, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func selMut(mut Mut, path string, full bool) (res Mut, p cor.Path, s cor.Seg, err error) {
	p, err = cor.ParsePath(path)
	if err != nil {
		return
	}
	if len(p) == 0 {
		return mut, p, s, nil
	}
	lst := len(p) - 1
	s = p[lst]
	if !full {
		p = p[:lst]
	}
	if len(p) > 0 {
		var found Val
		found, err = SelectPath(mut, p)
		if err != nil {
			return
		}
		m, ok := found.(Mut)
		if !ok {
			err = fmt.Errorf("expect mutable got %T", found)
			return
		}
		mut = m
	}
	return mut, p, s, nil
}

func applyKeyrDel(mut Mut, path string) error {
	mut, _, s, err := selMut(mut, path, false)
	if err != nil {
		return err
	}
	if s.Key == "" {
		return fmt.Errorf("expect key got %v in %s", s, path)
	}
	k, ok := Unwrap(mut).(Keyr)
	if !ok {
		return fmt.Errorf("expect keyr got %T", mut)
	}
	return k.SetKey(s.Key, nil)
}

func applyListAppend(mut Mut, key string, v Val) error {
	mut, _, _, err := selMut(mut, key, true)
	if err != nil {
		return err
	}
	args, ok := toVals(v)
	if !ok {
		return fmt.Errorf("expect list ops got %T", v)
	}
	vals, ok := toVals(Unwrap(mut))
	if !ok {
		return fmt.Errorf("expect list ops list target got %T", mut)
	}
	res := make([]Val, 0, len(vals)+len(args))
	res = append(res, vals...)
	res = append(res, args...)
	return mut.Assign(&List{Vals: res})
}

func applyListOps(mut Mut, key string, v Val) error {
	mut, _, _, err := selMut(mut, key, true)
	if err != nil {
		return err
	}
	ops, ok := toVals(v)
	if !ok {
		return fmt.Errorf("expect list ops got %T", v)
	}
	vals, ok := toVals(Unwrap(mut))
	if !ok {
		return fmt.Errorf("expect list ops list target got %T", mut)
	}
	res := make([]Val, 0, len(vals))
	var ret, del int
	for _, op := range ops {
		if op.Type().Kind&knd.Int != 0 {
			n, err := ToInt(op)
			if err != nil {
				return err
			}
			if n > 0 {
				idx := ret + del
				res = append(res, vals[idx:idx+int(n)]...)
				ret += int(n)
			} else if n < 0 {
				del += int(-n)
			} else {
				return fmt.Errorf("unexpected zero ops")
			}
		} else if op.Type().Kind&knd.List != 0 {
			vs, ok := toVals(op)
			if !ok {
				return fmt.Errorf("expect list op vals list got %T", v)
			}
			res = append(res, vs...)
		}
	}
	if idx := ret + del; idx < len(vals) {
		res = append(res, vals[idx:]...)
	}
	return mut.Assign(&List{Vals: res})
}

func deltaIdxr(a, b Val, aa, bb []Val, pre string, d map[string]Val) error {
	chgs := diff.Diff(len(aa), len(bb), &valsDiff{aa, bb})
	if len(chgs) == 0 {
		return nil
	}
	// how much and how often we retain and delete from a and insert from b
	ops, t := diffToOps(chgs, aa, bb)
	if !t.changed() {
		return nil
	} else if t.replaced() {
		d[stripTailDot(pre)] = b
		return nil
	}
	// we have at least two ops and known at least one of them to be ret and one del or ins
	// ops of the same kind are merged and do not follow each other

	// we want to detect append and use special syntax. append does only occur when we have
	// two ops u,v where u is ret and v is ins
	if len(ops) == 2 && ops[0].N > 0 && ops[1].N == 0 {
		// lets return the special append op
		d[stripTailDot(pre)+"+"] = &List{Vals: ops[1].V}
		return nil
	}
	// we also want to detect replacing a single element and use idx path notation. that does
	// only occur in two instances:

	// we have three ops u,v,w. retn is 1. v is del or ins and either u or w is the other
	// and we have four ops u,v,w,x. retn is 2. v and w are del and ins
	if len(ops) == 3 && t.retn == 1 || len(ops) == 4 && t.retn == 2 {
		u, v, w := ops[0], ops[1], ops[2]
		if v.N == -1 {
			if len(w.V) == 1 {
				return deltaSub(aa[u.N], w.V[0], pre, u.N, d)
			}
			if t.retn == 1 && len(u.V) == 1 {
				return deltaSub(aa[0], u.V[0], pre, 0, d)
			}
		} else if len(v.V) == 1 {
			if w.N == -1 {
				return deltaSub(aa[u.N], v.V[0], pre, u.N, d)
			}
			if t.retn == 1 && u.N == -1 {
				return deltaSub(aa[0], v.V[0], pre, 0, d)
			}
		}
	}
	// lets return the ops as list
	d[stripTailDot(pre)+"*"] = opsToList(ops)
	return nil
}

func deltaSub(a, b Val, pre string, idx int, d map[string]Val) error {
	path := fmt.Sprintf("%s%d.", pre, idx)
	return delta(a, b, path, d)
}

func deltaKeyr(a, b Keyr, pre string, d map[string]Val) error {
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
					return err
				}
				path := k
				if pre != "." {
					path = pre + k
				}
				d[path] = v
				// mark as handled
				km[k] = false
			} // duplicate key in b
			continue
		}
		// exists in a and b
		av, err := a.Key(k)
		if err != nil {
			return err
		}
		bv, err := b.Key(k)
		if err != nil {
			return err
		}
		// call delta on the values
		path := pre + k
		nvals := make(map[string]Val)
		err = delta(av, bv, path+".", nvals)
		if err != nil {
			return err
		}
		if pre == "." {
			// check for simple path and turn them into plain keys
			for nk, nv := range nvals {
				if nk == path {
					delete(nvals, nk)
					nvals[k] = nv
				}
			}
		}
		// append edits and mark as handled
		for nk, nv := range nvals {
			d[nk] = nv
		}
		km[k] = false
	}
	for k, v := range km {
		if v { // deleted key
			d[pre+k+"-"] = Null{}
		}
	}
	return nil
}

func stripTailDot(s string) string {
	if len(s) > 1 && s[len(s)-1] == '.' {
		return s[:len(s)-1]
	}
	return s
}

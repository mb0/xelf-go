package lit

import (
	"bytes"
	"fmt"

	"xelf.org/xelf/cor"
)

// Apply applies edits d to mutable a or returns an error.
func Apply(mut Mut, d Delta) (Mut, error) {
	for _, kv := range d {
		k := kv.Key
		if k == "" {
			return nil, fmt.Errorf("empty delta edit path")
		}
		suf := k[len(k)-1]
		switch suf {
		case '-', '*', '+':
			k = k[:len(k)-1]
		}
		p, err := cor.ParsePath(k)
		if err != nil {
			return nil, err
		}
		if vn := p.CountVars(); vn > 0 {
			vals, ok := kv.Val.(*Vals)
			if !ok {
				return nil, fmt.Errorf("expect path vars got %T", kv.Val)
			}
			vs := *vals
			n := len(vs)
			long := suf != '-' || vn == n-1
			if long {
				n--
			}
			vars := make([]string, n)
			for i := range vars {
				vars[i] = vs[i].String()
			}
			err = p.FillVars(vars)
			if err != nil {
				return nil, err
			}
			if long {
				kv.Val = vs[len(vs)-1]
			} else {
				kv.Val = Null{}
			}
		}
		switch suf {
		case '-':
			return mut, applyDelete(mut, p)
		case '*':
			return mut, applyOps(mut, p, kv.Val, false)
		case '+':
			return mut, applyOps(mut, p, kv.Val, true)
		}
		if (len(p) == 0 || len(p) == 1 && p.Fst().Empty()) && kv.Nil() {
			mut = AnyWrap(mut.Type())
		} else {
			mut, err = CreatePath(mut, p, kv.Val)
			if err != nil {
				return nil, err
			}
		}
	}
	return mut, nil
}

func selMut(mut Mut, p cor.Path, full bool) (res Mut, _ cor.Path, s cor.Seg, err error) {
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

func applyDelete(mut Mut, p cor.Path) error {
	mut, _, s, err := selMut(mut, p, false)
	if err != nil {
		return err
	}
	if s.Key == "" {
		return fmt.Errorf("expect key got %v in %s", s, p)
	}
	k, ok := Unwrap(mut).(Keyr)
	if !ok {
		return fmt.Errorf("expect keyr got %T", mut)
	}
	return k.SetKey(s.Key, nil)
}

func applyOps(mut Mut, p cor.Path, val Val, mirror bool) error {
	mut, _, _, err := selMut(mut, p, true)
	if err != nil {
		return err
	}
	ops, ok := toVals(val)
	if !ok {
		return fmt.Errorf("expect op slice got %T", val)
	}
	switch v := Unwrap(mut).Mut().(type) {
	case Idxr:
		vals, ok := toVals(v)
		if !ok {
			return fmt.Errorf("expect ops val slice got %T", v)
		}
		return applyListOps(mut, vals, ops, mirror)
	case *CharMut:
		// TODO decide whether it is a str or raw op
		return applyStrOps(mut, []rune(*v), ops, mirror)
	case *StrMut:
		return applyStrOps(mut, []rune(*v), ops, mirror)
	case *RawMut:
		return applyRawOps(mut, []byte(*v), ops, mirror)
	}
	return fmt.Errorf("expect ops val got %T for target %T", val, Unwrap(mut).Mut())
}

func applyStrOps(mut Mut, rs []rune, vals Vals, mirror bool) error {
	ops := make(StrOps, 0, len(vals)+1)
	err := readOps(len(rs), vals, func(n int, v Val) {
		s, _ := v.(Str)
		ops = append(ops, StrOp{N: n, V: s})
	})
	if err != nil {
		return err
	}
	if mirror {
		mirrorOps(ops)
	}
	var b bytes.Buffer
	var ret, del int
	for _, op := range ops {
		if op.N > 0 {
			idx := ret + del
			b.WriteString(string(rs[idx : idx+op.N]))
			ret += op.N
		} else if op.N < 0 {
			del += -op.N
		} else {
			b.WriteString(string(op.V))
		}
	}
	return mut.Assign(Str(b.String()))
}

func applyRawOps(mut Mut, bs []byte, vals Vals, mirror bool) error {
	ops := make(RawOps, 0, len(vals)+1)
	err := readOps(len(bs), vals, func(n int, v Val) {
		r, _ := v.(Raw)
		ops = append(ops, RawOp{N: n, V: r})
	})
	if err != nil {
		return err
	}
	if mirror {
		mirrorOps(ops)
	}
	var b bytes.Buffer
	var ret, del int
	for _, op := range ops {
		if op.N > 0 {
			idx := ret + del
			b.Write(bs[idx : idx+op.N])
			ret += op.N
		} else if op.N < 0 {
			del += -op.N
		} else {
			b.Write(op.V)
		}
	}
	return mut.Assign(Raw(b.Bytes()))
}
func applyListOps(mut Mut, vals Vals, ovals Vals, mirror bool) error {
	ops := make(ListOps, 0, len(ovals)+1)
	err := readOps(len(vals), ovals, func(n int, v Val) {
		vs, _ := toVals(v)
		ops = append(ops, ListOp{N: n, V: vs})
	})
	if err != nil {
		return err
	}
	if mirror {
		mirrorOps(ops)
	}
	var ret, del, nn int
	for _, op := range ops {
		if op.N > 0 {
			nn += op.N
		}
		nn += len(op.V)
	}
	res := make(Vals, 0, nn)
	for _, op := range ops {
		if op.N > 0 {
			idx := ret + del
			res = append(res, vals[idx:idx+op.N]...)
			ret += op.N
		} else if op.N < 0 {
			del += -op.N
		} else {
			res = append(res, op.V...)
		}
	}
	return mut.Assign(&res)
}

// toVals returns the idxr values of v or nil if v is not an idxr or has no values.
func toVals(val Val) (Vals, bool) {
	switch v := val.(type) {
	case *Vals:
		return *v, true
	case *List:
		return v.Vals, true
	default:
		if x, ok := v.(Idxr); ok {
			vs := make(Vals, 0, x.Len())
			x.IterIdx(func(idx int, el Val) error {
				vs = append(vs, el)
				return nil
			})
			return vs, true
		}
	}
	return nil, false
}

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
		if p.HasVars() {
			vals, ok := kv.Val.(*Vals)
			if !ok {
				return nil, fmt.Errorf("expect path vars got %T", kv.Val)
			}
			vs := *vals
			vars := make([]string, len(vs)-1)
			for i := range vars {
				vars[i] = vs[i].String()
			}
			err = p.FillVars(vars)
			if err != nil {
				return nil, err
			}
			kv.Val = vs[len(vs)-1]
		}
		switch suf {
		case '-':
			return mut, applyDelete(mut, p)
		case '*':
			// TODO special
			return mut, applyOps(mut, p, kv.Val)
		case '+':
			// TODO mirror op
			return mut, applyAppend(mut, p, kv.Val)
		}
		if (len(p) == 0 || len(p) == 1 && p.Fst().Empty()) && kv.Nil() {
			mut = AnyWrap(mut.Type())
		} else {
			err = CreatePath(mut, p, kv.Val)
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

func applyAppend(mut Mut, p cor.Path, val Val) error {
	mut, _, _, err := selMut(mut, p, true)
	if err != nil {
		return err
	}
	switch v := val.(type) {
	case *Vals:
		if apdr, ok := mut.Value().(Appender); ok {
			return apdr.Append(*v...)
		}
	case Str:
		if s, ok := mut.Value().(Str); ok {
			return mut.Assign(s + v)
		}
	case Raw:
		if r, ok := mut.Value().(Raw); ok {
			return mut.Assign(append(r, v...))
		}
	}
	return fmt.Errorf("expect ops val got %T for target %T", val, mut.Value())
}

func applyOps(mut Mut, p cor.Path, val Val) error {
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
		return applyListOps(mut, vals, ops)
	case *CharMut:
		// TODO decide whether it is a str or raw op
		return applyStrOps(mut, []rune(*v), ops)
	case *StrMut:
		return applyStrOps(mut, []rune(*v), ops)
	case *RawMut:
		return applyRawOps(mut, []byte(*v), ops)
	}
	return fmt.Errorf("expect ops val got %T for target %T", val, Unwrap(mut).Mut())
}

func applyStrOps(mut Mut, rs []rune, ops Vals) error {
	var b bytes.Buffer
	var ret, del int
	for _, op := range ops {
		switch n := op.(type) {
		case Int:
			if n > 0 {
				idx := ret + del
				b.WriteString(string(rs[idx : idx+int(n)]))
				ret += int(n)
			} else if n < 0 {
				del += int(-n)
			}
		case Str:
			b.WriteString(string(n))
		}
	}
	if idx := ret + del; idx < len(rs) {
		b.WriteString(string(rs[idx:]))
	}
	return mut.Assign(Str(b.String()))
}

func applyRawOps(mut Mut, bs []byte, ops Vals) error {
	var b bytes.Buffer
	var ret, del int
	for _, op := range ops {
		switch n := op.(type) {
		case Int:
			if n > 0 {
				idx := ret + del
				b.Write(bs[idx : idx+int(n)])
				ret += int(n)
			} else if n < 0 {
				del += int(-n)
			}
		case Raw:
			b.Write(n)
		}
	}
	if idx := ret + del; idx < len(bs) {
		b.Write(bs[idx:])
	}
	return mut.Assign(Raw(b.Bytes()))
}
func applyListOps(mut Mut, vals Vals, ops Vals) error {
	var ret, del, add, nn int
	for _, op := range ops {
		switch n := op.(type) {
		case Int:
			if n > 0 {
				ret += int(n)
			} else if n < 0 {
				del += int(-n)
			}
		case *Vals:
			add += len(*n)
		}
		add += len(vals) - ret - del
	}
	nn, ret, del = ret+add, 0, 0
	res := make(Vals, 0, nn)
	for _, op := range ops {
		switch n := op.(type) {
		case Int:
			if n > 0 {
				idx := ret + del
				res = append(res, vals[idx:idx+int(n)]...)
				ret += int(n)
			} else if n < 0 {
				del += int(-n)
			}
		case *Vals:
			res = append(res, *n...)
		}
	}
	if idx := ret + del; idx < len(vals) {
		res = append(res, vals[idx:]...)
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

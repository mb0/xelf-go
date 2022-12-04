package lit

import (
	"bytes"
	"fmt"

	"xelf.org/xelf/cor"
)

// Apply applies edits d to mutable a or returns an error.
func Apply(mut Mut, d Delta) error {
	for _, kv := range d {
		key := kv.Key
		if key != "" && key != "." && key[0] == '.' {
			lst := len(key) - 1
			if suf := key[lst]; suf == '+' {
				return applyAppend(mut, key[:lst], kv.Val)
			} else if suf == '*' {
				return applyOps(mut, key[:lst], kv.Val)
			} else if suf == '-' {
				return applyDelete(mut, key[:lst])
			}
		}
		p, err := cor.ParsePath(key)
		if err != nil {
			return err
		}
		err = CreatePath(mut, p, kv.Val)
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

func applyDelete(mut Mut, path string) error {
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

func applyAppend(mut Mut, key string, val Val) error {
	mut, _, _, err := selMut(mut, key, true)
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

func applyOps(mut Mut, key string, val Val) error {
	mut, _, _, err := selMut(mut, key, true)
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
		case Raw:
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

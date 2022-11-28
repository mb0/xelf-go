package lit

import "xelf.org/xelf/knd"

func Copy(v Val) (Val, error) {
	if v.Nil() {
		return v, nil
	}
	if v.Type().Kind&(knd.Keyr|knd.Idxr) == 0 {
		n := v.Mut().New()
		return n, n.Assign(v)
	}
	return deepCopy(v, make(map[Mut]Mut))
}

func deepCopy(v Val, cache map[Mut]Mut) (_ Val, err error) {
	m := v.Mut()
	if n := cache[m]; n != nil {
		return n, nil
	}
	n := m.New()
	cache[m] = n
	switch mc := m.(type) {
	case Keyr:
		nc := n.(Keyr)
		err = mc.IterKey(func(key string, el Val) error {
			nel, err := deepCopy(el, cache)
			if err != nil {
				return err
			}
			return nc.SetKey(key, nel)
		})
	case Idxr:
		nc := n.(Idxr)
		err = mc.IterIdx(func(idx int, el Val) error {
			nel, err := deepCopy(el, cache)
			if err != nil {
				return err
			}
			return nc.SetIdx(idx, nel)
		})
	default:
		err = n.Assign(m)
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

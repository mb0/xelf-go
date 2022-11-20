package lit

func Copy(v Val) (Val, error) {
	cache := make(map[Mut]Mut)
	return deepCopy(v, cache)
}
func deepCopy(v Val, cache map[Mut]Mut) (Val, error) {
	m, ok := v.(Mut)
	if !ok {
		return v.Value(), nil
	}
	if n := cache[m]; n != nil {
		return n, nil
	}
	n, err := m.New()
	if err != nil {
		return nil, err
	}
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
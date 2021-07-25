package typ

import (
	"sort"

	"xelf.org/xelf/knd"
)

func altTypes(a Type) []Type {
	if !a.Kind.IsAlt() {
		return []Type{a}
	}
	k := a.Kind &^ (knd.None | knd.Alt | knd.Var)
	n := k.Count()
	b, _ := a.Body.(*AltBody)
	if b != nil {
		n += len(b.Alts)
	}
	res := make([]Type, 0, n)
	for i := len(knd.Infos) - 1; k > 0 && i > 0; i-- {
		info := knd.Infos[i]
		if k&info.Kind == info.Kind {
			res = append(res, Type{Kind: info.Kind})
			k = k &^ info.Kind
		}
	}
	sort.Sort(byKind(res))
	if b != nil {
		res = append(res, b.Alts...)
	}
	return res
}

type byKind []Type

func (s byKind) Len() int           { return len(s) }
func (s byKind) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byKind) Less(i, j int) bool { return s[i].Kind < s[j].Kind }

package exp

import (
	"fmt"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

type Tags []Tag

func (tags Tags) FindKey(key string) *Tag {
	for i, t := range tags {
		if t.Tag == key {
			return &tags[i]
		}
	}
	return nil
}

func (tags Tags) Select(sym string) (*Lit, error) {
	path, err := cor.ParsePath(sym)
	if err != nil {
		return nil, err
	}
	if len(path) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	fst := path[0]
	var tag *Tag
	if fst.Key != "" {
		tag = tags.FindKey(fst.Key)
	} else if fst.Idx >= 0 && fst.Idx < len(tags) {
		tag = &tags[fst.Idx]
	}
	if tag == nil {
		return nil, fmt.Errorf("not found %v", fst)
	}
	l, ok := tag.Exp.(*Lit)
	if !ok {
		return nil, fmt.Errorf("unresolved")
	}
	return l.SelectPath(path[1:])
}

func (l *Lit) Select(path string) (*Lit, error) {
	p, err := cor.ParsePath(path)
	if err != nil {
		return nil, err
	}
	return l.SelectPath(p)
}
func (l *Lit) SelectPath(path cor.Path) (*Lit, error) {
	if len(path) == 0 {
		return l, nil
	}
	v, err := lit.SelectPath(l.Val, path)
	if err != nil {
		return nil, err
	}
	t, err := typ.SelectPath(l.Res, path)
	if err != nil {
		return nil, err
	}
	return &Lit{Res: t, Val: v}, nil
}

package lit

import (
	"bytes"
	"fmt"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

type List struct {
	El   typ.Type
	Vals []Val
}

func NewList(el typ.Type) *List { return &List{El: el} }

func (l *List) Type() typ.Type               { return typ.ListOf(l.El) }
func (l *List) Nil() bool                    { return l == nil }
func (l *List) Zero() bool                   { return len(l.Vals) == 0 }
func (l *List) Value() Val                   { return l }
func (l *List) MarshalJSON() ([]byte, error) { return bfr.JSON(l) }
func (l *List) String() string               { return bfr.String(l) }
func (l *List) Print(p *bfr.P) (err error) {
	p.Byte('[')
	if len(l.Vals) > 0 {
		p.Indent()
		for i, v := range l.Vals {
			if i > 0 {
				p.Sep()
				p.Break()
			}
			if err = v.Print(p); err != nil {
				return err
			}
		}
		p.Dedent()
	}
	return p.Byte(']')
}
func (l *List) New() Mut         { return &List{l.El, nil} }
func (l *List) Ptr() interface{} { return l }
func (l *List) Assign(p Val) error {
	// TODO compare types
	switch o := p.(type) {
	case nil:
		l.Vals = nil
	case Null:
		l.Vals = nil
	case Idxr:
		res := make([]Val, 0, o.Len())
		err := o.IterIdx(func(i int, v Val) error {
			res = append(res, v)
			return nil
		})
		if err != nil {
			return err
		}
		l.Vals = res
	default:
		return ErrAssign
	}
	return nil
}
func (l *List) Append(p Val) error {
	// TODO compare types
	if l.El == typ.Void {
		l.Vals = append(l.Vals, p)
		return nil
	}
	l.Vals = append(l.Vals, p)
	return nil
}
func (l *List) Len() int { return len(l.Vals) }
func (l *List) Idx(i int) (Val, error) {
	if i < 0 || i >= len(l.Vals) {
		return nil, ErrIdxBounds
	}
	return l.Vals[i], nil
}
func (l *List) SetIdx(i int, el Val) error {
	if i < 0 {
		return ErrIdxBounds
	}
	if el == nil {
		el = Null{}
	}
	if l.El != typ.Void && l.El != typ.Data {
		// TODO check and conevert el
	}
	if i >= len(l.Vals) {
		n := l.Vals
		if i < cap(n) {
			n = n[:i+1]
			for j := len(l.Vals) - 1; j < i; j++ {
				n[j] = nil
			}
		} else {
			n = make([]Val, i+1, (i+1)*3/2)
			copy(n, l.Vals)
		}
		l.Vals = n
	}
	l.Vals[i] = el
	return nil
}
func (l *List) IterIdx(it func(int, Val) error) error {
	for i, el := range l.Vals {
		if err := it(i, el); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (l *List) UnmarshalJSON(b []byte) error {
	lit, err := Read(bytes.NewReader(b), "")
	if err != nil {
		return err
	}
	o, ok := lit.Val.(*List)
	if !ok {
		return fmt.Errorf("expect list got %T", lit.Val)
	}
	*l = *o
	return nil
}

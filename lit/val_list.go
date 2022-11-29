package lit

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type Vals []Val

func (v Vals) Type() typ.Type                { return typ.Idxr }
func (v *Vals) Nil() bool                    { return v == nil }
func (v *Vals) Zero() bool                   { return v == nil || len(*v) == 0 }
func (v *Vals) Mut() Mut                     { return v }
func (v *Vals) Value() Val                   { return v }
func (v *Vals) UnmarshalJSON(b []byte) error { return unmarshal(b, v) }
func (v Vals) MarshalJSON() ([]byte, error)  { return bfr.JSON(v) }
func (v Vals) String() string                { return bfr.String(v) }
func (v Vals) Print(p *bfr.P) (err error) {
	p.Byte('[')
	if len(v) > 0 {
		p.Indent()
		for i, v := range v {
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

func (*Vals) New() Mut           { return &Vals{} }
func (v *Vals) Ptr() interface{} { return v }
func (v *Vals) Parse(a ast.Ast) (err error) {
	if isNull(a) {
		*v = nil
		return nil
	}
	if a.Kind != knd.Idxr {
		return ast.ErrExpect(a, knd.Idxr)
	}
	vs := make([]Val, 0, len(a.Seq))
	for _, e := range a.Seq {
		var el Val
		el, err = parseMutNull(e)
		if err != nil {
			return err
		}
		vs = append(vs, el)
	}
	*v = vs
	return nil
}
func (v *Vals) Assign(p Val) error {
	// TODO compare types
	switch o := p.(type) {
	case nil:
		*v = nil
	case Null:
		*v = nil
	case Idxr:
		res := make([]Val, 0, o.Len())
		err := o.IterIdx(func(i int, v Val) error {
			res = append(res, v)
			return nil
		})
		if err != nil {
			return err
		}
		*v = res
	default:
		return fmt.Errorf("%w %T to %T", ErrAssign, p, v)
	}
	return nil
}
func (v *Vals) Append(p Val) error {
	*v = append(*v, p)
	return nil
}
func (v Vals) Len() int { return len(v) }
func (v Vals) Idx(i int) (res Val, err error) {
	if i, err = checkIdx(i, len(v)); err != nil {
		return
	}
	return v[i], nil
}
func (v *Vals) SetIdx(i int, el Val) (err error) {
	if i, err = checkIdx(i, len(*v)); err != nil {
		return
	}
	if el == nil {
		el = Null{}
	}
	if i >= len(*v) {
		n := *v
		if i < cap(n) {
			n = n[:i+1]
			for j := len(*v) - 1; j < i; j++ {
				n[j] = nil
			}
		} else {
			n = make([]Val, i+1, (i+1)*3/2)
			copy(n, *v)
		}
		*v = n
	}
	(*v)[i] = el
	return nil
}
func (v Vals) IterIdx(it func(int, Val) error) error {
	for i, el := range v {
		if err := it(i, el); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}

type List struct {
	Typ typ.Type
	Vals
}

func NewList(el typ.Type, vs ...Val) *List {
	return &List{typ.ListOf(el), vs}
}

func (l *List) Type() typ.Type {
	if l.Typ == typ.Void {
		l.Typ = typ.List
	}
	return l.Typ
}
func (l *List) Nil() bool        { return l == nil }
func (l *List) Mut() Mut         { return l }
func (l *List) Value() Val       { return l }
func (l *List) New() Mut         { return &List{l.Typ, nil} }
func (l *List) Ptr() interface{} { return l }

func checkIdx(idx, l int) (int, error) {
	i := idx
	if i < 0 {
		i = l + i
	}
	if i < 0 || i >= l {
		return 0, fmt.Errorf("idx %d %w [0:%d]", idx, ErrIdxBounds, l-1)
	}
	return i, nil
}

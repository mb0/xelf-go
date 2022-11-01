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
func (v *Vals) Value() Val                   { return v }
func (v *Vals) UnmarshalJSON(b []byte) error { return unmarshal(b, v, nil) }
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

func (*Vals) New() (Mut, error)  { return &Vals{}, nil }
func (*Vals) WithReg(reg *Reg)   {}
func (v *Vals) Ptr() interface{} { return v }
func (v *Vals) Parse(reg typ.Reg, a ast.Ast) (err error) {
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
		el, err = parseMutNull(reg, e)
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
	Reg *Reg
	El  typ.Type
	Vals
}

func NewList(reg *Reg, el typ.Type) *List                { return &List{Reg: reg, El: el} }
func (l *List) Type() typ.Type                           { return typ.ListOf(l.El) }
func (l *List) Nil() bool                                { return l == nil }
func (l *List) Value() Val                               { return l }
func (l *List) UnmarshalJSON(b []byte) error             { return unmarshal(b, l, l.Reg) }
func (l *List) New() (Mut, error)                        { return &List{l.Reg, l.El, nil}, nil }
func (l *List) WithReg(reg *Reg)                         { l.Reg = reg }
func (l *List) Ptr() interface{}                         { return l }
func (l *List) Parse(reg typ.Reg, a ast.Ast) (err error) { return l.Vals.Parse(l.Reg, a) }

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

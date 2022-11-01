package lit

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type List struct {
	Reg  *Reg
	El   typ.Type
	Vals []Val
}

func NewList(reg *Reg, el typ.Type) *List { return &List{Reg: reg, El: el} }

func (l *List) Type() typ.Type               { return typ.ListOf(l.El) }
func (l *List) Nil() bool                    { return l == nil }
func (l *List) Zero() bool                   { return len(l.Vals) == 0 }
func (l *List) Value() Val                   { return l }
func (l *List) MarshalJSON() ([]byte, error) { return bfr.JSON(l) }
func (l *List) UnmarshalJSON(b []byte) error { return unmarshal(b, l, l.Reg) }
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
func (l *List) New() (Mut, error) { return &List{l.Reg, l.El, nil}, nil }
func (l *List) WithReg(reg *Reg)  { l.Reg = reg }
func (l *List) Ptr() interface{}  { return l }
func (l *List) Parse(reg typ.Reg, a ast.Ast) (err error) {
	if isNull(a) {
		l.Vals = nil
		return nil
	}
	if a.Kind != knd.List {
		return ast.ErrExpect(a, knd.List)
	}
	vs := make([]Val, 0, len(a.Seq))
	for _, e := range a.Seq {
		var el Val
		el, err := parseMutNull(l.Reg, e)
		if err != nil {
			return err
		}
		vs = append(vs, el)
	}
	l.Vals = vs
	return nil
}
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
		return fmt.Errorf("%w %T to %T", ErrAssign, p, l)
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
func (l *List) Idx(i int) (res Val, err error) {
	if i, err = checkIdx(i, len(l.Vals)); err != nil {
		return
	}
	return l.Vals[i], nil
}
func (l *List) SetIdx(i int, el Val) (err error) {
	if i, err = checkIdx(i, len(l.Vals)); err != nil {
		return
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

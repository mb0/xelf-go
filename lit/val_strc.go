package lit

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type Strc struct {
	Reg  *Reg
	Typ  typ.Type
	Vals []Val
}

func NewStrc(reg *Reg, t typ.Type) (*Strc, error) {
	s := &Strc{Reg: reg, Typ: t}
	err := s.init(false)
	return s, err
}
func (s *Strc) init(ext bool) (err error) {
	ps := s.ps()
	vs := make([]Val, len(ps))
	if ext {
		copy(vs, s.Vals)
	}
	for i, p := range ps {
		if ext && vs[i] != nil {
			continue
		}
		vs[i], err = s.Reg.Zero(p.Type)
		if err != nil {
			break
		}
	}
	s.Vals = vs
	return
}
func (s *Strc) Type() typ.Type { return s.Typ }
func (s *Strc) Nil() bool      { return len(s.Vals) == 0 }
func (s *Strc) Zero() bool {
	for _, v := range s.Vals {
		if v != nil && !v.Zero() {
			return false
		}
	}
	return true
}
func (s *Strc) Value() Val                   { return s }
func (s *Strc) MarshalJSON() ([]byte, error) { return bfr.JSON(s) }
func (s *Strc) UnmarshalJSON(b []byte) error { return unmarshal(b, s) }
func (s *Strc) String() string               { return bfr.String(s) }
func (s *Strc) Print(p *bfr.P) (err error) {
	p.Byte('{')
	for i, par := range s.ps() {
		if i > 0 {
			p.Sep()
		}
		p.RecordKey(par.Key)
		var v Val
		if i < len(s.Vals) {
			v = s.Vals[i]
		}
		if v == nil || v.Zero() {
			PrintZero(p, par.Type)
			continue
		}
		if err = v.Print(p); err != nil {
			return err
		}
	}
	return p.Byte('}')
}
func (s *Strc) New() (Mut, error) { return NewStrc(s.Reg, s.Typ) }
func (s *Strc) Ptr() interface{}  { return s }
func (s *Strc) Parse(a ast.Ast) error {
	if isNull(a) {
		return s.init(false)
	}
	if a.Kind != knd.Dict {
		return ast.ErrExpect(a, knd.Dict)
	}
	pb := s.Typ.Body.(*typ.ParamBody)
	vs := make([]Val, len(pb.Params))
	for _, e := range a.Seq {
		key, val, err := ast.UnquotePair(e)
		if err != nil {
			return err
		}
		el, err := s.Reg.parseMutNull(val)
		if err != nil {
			return err
		}
		i := pb.FindKeyIndex(key)
		if i >= 0 {
			vs[i] = el
		}
	}
	for i, v := range vs {
		if v != nil {
			continue
		}
		z, err := s.Reg.Zero(pb.Params[i].Type)
		if err != nil {
			return err
		}
		vs[i] = z
	}
	return nil
}
func (s *Strc) Assign(p Val) error {
	// TODO check types
	s.init(false)
	switch o := p.(type) {
	case nil:
	case Null:
	case Keyr:
		err := o.IterKey(func(k string, v Val) error {
			return s.SetKey(k, v)
		})
		if err != nil {
			return err
		}
	case Idxr:
		err := o.IterIdx(func(i int, v Val) error {
			return s.SetIdx(i, v)
		})
		if err != nil {
			return err
		}
	default:
		return ErrAssign
	}
	return nil
}
func (s *Strc) Len() int { return len(s.ps()) }
func (s *Strc) Idx(i int) (Val, error) {
	ps, ok := s.pidx(i)
	if !ok {
		return nil, ErrIdxBounds
	}
	if len(s.Vals) < len(ps) {
		if err := s.init(true); err != nil {
			return nil, err
		}
	}
	return s.Vals[i], nil
}
func (s *Strc) SetIdx(idx int, el Val) error {
	ps, ok := s.pidx(idx)
	if !ok {
		return ErrIdxNotFound
	}
	if len(s.Vals) < len(ps) {
		if err := s.init(true); err != nil {
			return err
		}
	}
	if el == nil {
		el = Null{}
	}
	s.Vals[idx] = el
	return nil
}
func (s *Strc) IterIdx(it func(int, Val) error) error {
	ps := s.ps()
	if len(s.Vals) < len(ps) {
		if err := s.init(true); err != nil {
			return err
		}
	}
	for i, v := range s.Vals {
		if err := it(i, v); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (s *Strc) Keys() []string {
	ps := s.ps()
	res := make([]string, 0, len(ps))
	for _, p := range ps {
		res = append(res, p.Key)
	}
	return res
}
func (s *Strc) Key(k string) (Val, error) {
	ps, i := s.pkey(k)
	if i < 0 {
		return nil, fmt.Errorf("strc %s %q: %w", s.Typ, k, ErrKeyNotFound)
	}
	if len(s.Vals) < len(ps) {
		if err := s.init(true); err != nil {
			return nil, err
		}
	}
	return s.Vals[i], nil
}
func (s *Strc) SetKey(k string, el Val) error {
	ps, i := s.pkey(k)
	if i < 0 {
		return fmt.Errorf("strc %s %q: %w", s.Typ, k, ErrKeyNotFound)
	}
	if len(s.Vals) < len(ps) {
		if err := s.init(true); err != nil {
			return err
		}
	}
	if el == nil {
		el = Null{}
	}
	s.Vals[i] = el
	return nil
}
func (s *Strc) IterKey(it func(string, Val) error) error {
	ps := s.ps()
	if len(s.Vals) < len(ps) {
		if err := s.init(true); err != nil {
			return err
		}
	}
	for i, p := range ps {
		if err := it(p.Key, s.Vals[i]); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (s *Strc) ps() []typ.Param {
	pb := s.Typ.Body.(*typ.ParamBody)
	return pb.Params
}
func (s *Strc) pidx(i int) (ps []typ.Param, ok bool) {
	if i < 0 {
		return nil, false
	}
	ps = s.ps()
	if i >= len(ps) {
		return nil, false
	}
	return ps, true
}

func (s *Strc) pkey(k string) ([]typ.Param, int) {
	pb := s.Typ.Body.(*typ.ParamBody)
	i := pb.FindKeyIndex(k)
	if i < 0 {
		return nil, i
	}
	return pb.Params, i
}

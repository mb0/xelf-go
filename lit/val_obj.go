package lit

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type Obj struct {
	Typ  typ.Type
	Vals []Val
}

func NewObj(t typ.Type) (*Obj, error) {
	s := &Obj{Typ: t}
	_, err := s.pinit(false)
	return s, err
}

func MakeObj(kvs []KeyVal) *Obj {
	vs := make([]Val, 0, len(kvs))
	ps := make([]typ.Param, 0, len(kvs))
	for _, kv := range kvs {
		ps = append(ps, typ.P(kv.Key, kv.Val.Type()))
		vs = append(vs, kv.Val)
	}
	return &Obj{Typ: typ.Obj("", ps...), Vals: vs}
}

func (s *Obj) Type() typ.Type { return s.Typ }
func (s *Obj) Nil() bool      { return len(s.Vals) == 0 }
func (s *Obj) Zero() bool {
	for _, v := range s.Vals {
		if v != nil && !v.Zero() {
			return false
		}
	}
	return true
}
func (s *Obj) Value() Val                   { return s }
func (s *Obj) MarshalJSON() ([]byte, error) { return bfr.JSON(s) }
func (s *Obj) UnmarshalJSON(b []byte) error { return unmarshal(b, s) }
func (s *Obj) String() string               { return bfr.String(s) }
func (s *Obj) Print(p *bfr.P) error {
	p.Byte('{')
	pb, err := s.pinit(true)
	if err != nil {
		return err
	}
	for i, par := range pb.Params {
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
func (s *Obj) New() (Mut, error) { return NewObj(s.Typ) }
func (s *Obj) Ptr() interface{}  { return s }
func (s *Obj) Parse(a ast.Ast) error {
	if isNull(a) {
		_, err := s.pinit(false)
		return err
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
		el, err := parseMutNull(val)
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
		vs[i] = Zero(pb.Params[i].Type)
	}
	return nil
}
func (s *Obj) Assign(p Val) error {
	// TODO check types
	if _, err := s.pinit(false); err != nil {
		return err
	}
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
func (s *Obj) Len() int {
	pb, _ := s.Typ.Body.(*typ.ParamBody)
	if pb == nil {
		return 0
	}
	return len(pb.Params)
}
func (s *Obj) Idx(i int) (Val, error) {
	if _, err := s.pidx(i); err != nil {
		return nil, err
	}
	return s.Vals[i], nil
}
func (s *Obj) SetIdx(idx int, el Val) error {
	if _, err := s.pidx(idx); err != nil {
		return err
	}
	if el == nil {
		el = Null{}
	}
	s.Vals[idx] = el
	return nil
}
func (s *Obj) IterIdx(it func(int, Val) error) error {
	_, err := s.pinit(true)
	if err != nil {
		return err
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
func (s *Obj) Keys() []string {
	pb, _ := s.Typ.Body.(*typ.ParamBody)
	if pb == nil {
		return nil
	}
	res := make([]string, 0, len(pb.Params))
	for _, p := range pb.Params {
		res = append(res, p.Key)
	}
	return res
}
func (s *Obj) Key(k string) (Val, error) {
	pb, err := s.pinit(true)
	if err != nil {
		return nil, err
	}
	i := pb.FindKeyIndex(k)
	if i < 0 {
		return nil, fmt.Errorf("obj %s %q: %w", s.Typ, k, ErrKeyNotFound)
	}
	return s.Vals[i], nil
}
func (s *Obj) SetKey(k string, el Val) error {
	pb, err := s.pinit(true)
	if err != nil {
		return err
	}
	i := pb.FindKeyIndex(k)
	if i < 0 {
		return fmt.Errorf("obj %s %q: %w", s.Typ, k, ErrKeyNotFound)
	}
	if el == nil {
		el = Null{}
	}
	s.Vals[i] = el
	return nil
}
func (s *Obj) IterKey(it func(string, Val) error) error {
	pb, err := s.pinit(true)
	if err != nil {
		return err
	}
	for i, p := range pb.Params {
		if err := it(p.Key, s.Vals[i]); err != nil {
			if err == BreakIter {
				return nil
			}
			return err
		}
	}
	return nil
}
func (s *Obj) pinit(ext bool) (pb *typ.ParamBody, err error) {
	pb, _ = s.Typ.Body.(*typ.ParamBody)
	if pb == nil || s.Typ.Kind&knd.Obj == 0 {
		err = fmt.Errorf("not an obj type %s", s.Typ)
	} else if !ext || len(s.Vals) < len(pb.Params) {
		vs := make([]Val, len(pb.Params))
		if ext {
			copy(vs, s.Vals)
		}
		for i, p := range pb.Params {
			if ext && vs[i] != nil {
				continue
			}
			vs[i] = Zero(p.Type)
		}
		s.Vals = vs
	}
	return
}
func (s *Obj) pidx(i int) (*typ.ParamBody, error) {
	if i < 0 {
		return nil, ErrIdxBounds
	}
	pb, err := s.pinit(true)
	if err != nil || i >= len(pb.Params) {
		return nil, err
	}
	return pb, nil
}

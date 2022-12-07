package typ

import (
	"xelf.org/xelf/cor"
)

// Param describes a obj field or spec parameter.
type Param struct {
	Name string
	Key  string
	Type
}

func P(name string, t Type) Param {
	return Param{Name: name, Key: cor.Keyed(name), Type: t}
}
func (p Param) IsOpt() bool {
	return p.Name != "" && p.Name[len(p.Name)-1] == '?'
}

// Const describes a named constant value for an enum or bits type.
type Const struct {
	Name string
	Key  string
	Val  int64
}

func C(name string, v int64) Const {
	return Const{Name: name, Key: cor.Keyed(name), Val: v}
}

type (
	// AltBody contains compound type alternatives.
	AltBody struct {
		Alts []Type
	}

	// ParamBody contains a name and a list of parameters for obj and spec types.
	ParamBody struct {
		Params []Param
	}

	// ConstBody contains a name and a list of constants for the enum and bits types.
	ConstBody struct {
		Consts []Const
	}
)

// A *Type itself can be used as element body for expression and container types.
func (t *Type) EqualBody(b Body, h Hist) bool {
	o, ok := b.(*Type)
	if !ok {
		return false
	}
	for _, p := range h {
		if p.A == t && p.B == o {
			return true
		}
	}
	h = append(h, BodyPair{t, o})
	if t.Kind != o.Kind || t.ID != o.ID || t.Ref != o.Ref {
		return false
	}
	if t.Body == nil {
		return o.Body == nil
	}
	return t.Body.EqualBody(o.Body, h)
}

func (b *AltBody) EqualBody(o Body, h Hist) bool {
	ob, ok := o.(*AltBody)
	if b == nil {
		return ok && ob == nil
	}
	if !ok || ob == nil || len(b.Alts) != len(ob.Alts) {
		return false
	}
	for i, p := range b.Alts {
		op := ob.Alts[i]
		if !p.EqualHist(op, h) {
			return false
		}
	}
	return true
}

func (b *ParamBody) EqualBody(o Body, h Hist) bool {
	ob, ok := o.(*ParamBody)
	if !ok || len(b.Params) != len(ob.Params) {
		return false
	}
	for _, p := range h {
		if p.A == b && p.B == o {
			return true
		}
	}
	h = append(h, BodyPair{b, o})
	for i, p := range b.Params {
		op := ob.Params[i]
		if p.Name != op.Name || p.Key != op.Key || !p.EqualHist(op.Type, h) {
			return false
		}
	}
	return true
}
func (b *ParamBody) FindKeyIndex(key string) int {
	for i, p := range b.Params {
		if p.Key == key {
			return i
		}
	}
	return -1
}

func (b *ConstBody) EqualBody(o Body, h Hist) bool {
	ob, ok := o.(*ConstBody)
	if !ok || len(b.Consts) != len(ob.Consts) {
		return false
	}
	for i, p := range b.Consts {
		op := ob.Consts[i]
		if p.Name != op.Name || p.Key != op.Key || p.Val != op.Val {
			return false
		}
	}
	return true
}
func (b *ConstBody) FindKeyIndex(key string) int {
	for i, p := range b.Consts {
		if p.Key == key {
			return i
		}
	}
	return -1
}
func (b *ConstBody) FindValIndex(val int64) int {
	for i, p := range b.Consts {
		if p.Val == val {
			return i
		}
	}
	return -1
}

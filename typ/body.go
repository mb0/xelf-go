package typ

import "xelf.org/xelf/cor"

// Param describes a strc field or spec parameter.
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
	// ElBody contains an element type for expression and container types.
	ElBody struct {
		El Type
	}

	// SelBody contains a selection type and selection path into that type.
	// Selection types are mainly used internally for partially resolved selections.
	SelBody struct {
		Sel  Type
		Path string
	}

	// RefBody contains the type reference to named type.
	RefBody struct {
		Ref string
	}

	// AltBody contains compound type alternatives.
	AltBody struct {
		Alts []Type
	}

	// ParamBody contains a name and a list of parameters for spec and rec types.
	ParamBody struct {
		Name   string
		Params []Param
	}

	// ConstBody contains a name and a list of constants for the enum and bits types.
	ConstBody struct {
		Name   string
		Consts []Const
	}
)

func (b *ElBody) EqualHist(o Body, h Hist) bool {
	ob, ok := o.(*ElBody)
	if !ok {
		return false
	}
	for _, p := range h {
		if p.A == b && p.B == o {
			return true
		}
	}
	h = append(h, BodyPair{b, o})
	return b.El.EqualHist(ob.El, h)
}

func (b *SelBody) EqualHist(o Body, h Hist) bool {
	ob, ok := o.(*SelBody)
	return ok && b.Path == ob.Path && b.Sel.EqualHist(ob.Sel, h)
}

func (b *RefBody) EqualHist(o Body, h Hist) bool {
	ob, ok := o.(*RefBody)
	return ok && b.Ref == ob.Ref
}

func (b *AltBody) EqualHist(o Body, h Hist) bool {
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

func (b *ParamBody) EqualHist(o Body, h Hist) bool {
	ob, ok := o.(*ParamBody)
	if !ok || b.Name != ob.Name || len(b.Params) != len(ob.Params) {
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

func (b *ConstBody) EqualHist(o Body, h Hist) bool {
	ob, ok := o.(*ConstBody)
	if !ok || b.Name != ob.Name || len(b.Consts) != len(ob.Consts) {
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

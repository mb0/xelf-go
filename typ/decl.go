package typ

import (
	"xelf.org/xelf/knd"
)

var (
	Void = Type{Kind: knd.Void}
	None = Type{Kind: knd.None}
	Bool = Type{Kind: knd.Bool}
	Num  = Type{Kind: knd.Num}
	Int  = Type{Kind: knd.Int}
	Real = Type{Kind: knd.Real}
	Char = Type{Kind: knd.Char}
	Str  = Type{Kind: knd.Str}
	Raw  = Type{Kind: knd.Raw}
	UUID = Type{Kind: knd.UUID}
	Time = Type{Kind: knd.Time}
	Span = Type{Kind: knd.Span}

	Lit  = Type{Kind: knd.Lit}
	Typ  = Type{Kind: knd.Typ}
	Sym  = Type{Kind: knd.Sym}
	Tag  = Type{Kind: knd.Tag}
	Tupl = Type{Kind: knd.Tupl}
	Call = Type{Kind: knd.Call}
	Exp  = Type{Kind: knd.Exp}

	Idxr = Type{Kind: knd.Idxr}
	Keyr = Type{Kind: knd.Keyr}
	List = Type{Kind: knd.List}
	Dict = Type{Kind: knd.Dict}

	Data = Type{Kind: knd.Data}
	Spec = Type{Kind: knd.Spec}
	Any  = Type{Kind: knd.Any}
)

var (
	VarTyp  = Type{Kind: knd.Typ | knd.Var, ID: -1}
	VarSpec = Type{Kind: knd.Spec | knd.Var, ID: -1}
)

func Opt(t Type) Type {
	t.Kind |= knd.None
	return t
}
func Deopt(t Type) Type {
	t.Kind &^= knd.None
	return t
}

func WithID(id int32, t Type) Type {
	t.ID = id
	return t
}
func WithRef(r string, t Type) Type {
	t.Ref = r
	return t
}
func Var(id int32, t Type) Type {
	t.Kind |= knd.Var
	t.ID = id
	return t
}

func Ref(name string) Type { return Type{Kind: knd.Ref, Ref: name} }
func Sel(sel string) Type  { return Type{Kind: knd.Sel, Body: &SelBody{Path: sel}} }

func Bits(n string, cs ...Const) Type {
	return Type{Kind: knd.Bits, Ref: n, Body: &ConstBody{Consts: cs}}
}
func Enum(n string, cs ...Const) Type {
	return Type{Kind: knd.Enum, Ref: n, Body: &ConstBody{Consts: cs}}
}

func Obj(n string, ps ...Param) Type {
	return Type{Kind: knd.Obj, Ref: n, Body: &ParamBody{Params: ps}}
}

func elType(k knd.Kind, el Type) Type {
	if el == Void {
		return Type{Kind: k}
	}
	return Type{Kind: k, Body: &ElBody{El: el}}
}

func TypOf(t Type) Type  { return elType(knd.Typ, t) }
func LitOf(t Type) Type  { return elType(knd.Lit, t) }
func SymOf(t Type) Type  { return elType(knd.Sym, t) }
func TagOf(t Type) Type  { return elType(knd.Tag, t) }
func CallOf(t Type) Type { return elType(knd.Call, t) }
func ListOf(t Type) Type { return elType(knd.List, t) }
func DictOf(t Type) Type { return elType(knd.Dict, t) }
func IdxrOf(t Type) Type { return elType(knd.Idxr, t) }
func KeyrOf(t Type) Type { return elType(knd.Keyr, t) }

func ElemTupl(t Type) Type       { return Type{Kind: knd.Tupl, Body: &ElBody{El: t}} }
func ParamTupl(ps ...Param) Type { return Type{Kind: knd.Tupl, Body: &ParamBody{Params: ps}} }

func Func(name string, ps ...Param) Type {
	return Type{Kind: knd.Func, Ref: name, Body: &ParamBody{Params: ps}}
}
func Form(name string, ps ...Param) Type {
	return Type{Kind: knd.Form, Ref: name, Body: &ParamBody{Params: ps}}
}

func El(t Type) Type {
	if b, ok := t.Body.(*ElBody); ok && b.El.Kind != knd.Void {
		return b.El
	}
	return Void
}
func ResEl(t Type) Type {
	if t.Kind&(knd.Lit|knd.Exp) != 0 {
		if r := El(t); r != Void {
			return r
		}
		return Any
	}
	return t
}
func ContEl(t Type) Type {
	if t.Kind&knd.Cont != 0 {
		if r := El(t); r != Void {
			return r
		}
		return Any
	}
	return t
}

func TuplEl(t Type) (Type, int) {
	switch b := t.Body.(type) {
	case *ElBody:
		return b.El, 1
	case *ParamBody:
		if n := len(b.Params); n > 0 {
			return t, n
		}
	}
	return Any, 0
}

// Last returns the last element type if t is a list or dict type otherwise t is returned as is.
func Last(t Type) Type {
	for {
		b, ok := t.Body.(*ElBody)
		if !ok {
			break
		}
		t = b.El
		if t == Void {
			return Any
		}
	}
	return t
}

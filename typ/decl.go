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
func Var(id int32, t Type) Type {
	t.Kind |= knd.Var
	t.ID = id
	return t
}

func Ref(name string) Type { return Type{knd.Ref, 0, &RefBody{Ref: name}} }
func Sel(sel string) Type  { return Type{knd.Sel, 0, &SelBody{Path: sel}} }

func Rec(ps ...Param) Type          { return Type{knd.Rec, 0, &ParamBody{Params: ps}} }
func Obj(n string, ps []Param) Type { return Type{knd.Obj, 0, &ParamBody{Name: n, Params: ps}} }

func elType(k knd.Kind, el Type) Type {
	if el == Void {
		return Type{Kind: k}
	}
	return Type{k, 0, &ElBody{El: el}}
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

func TuplList(t Type) Type     { return TuplRec(P("", t)) }
func TuplRec(ps ...Param) Type { return Type{knd.Tupl, 0, &ParamBody{Params: ps}} }

func Func(name string, ps ...Param) Type { return Type{knd.Func, 0, &ParamBody{name, ps}} }
func Form(name string, ps ...Param) Type { return Type{knd.Form, 0, &ParamBody{name, ps}} }

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

func Name(t Type) string {
	if t.Kind&(knd.Schm|knd.Spec|knd.Ref) != 0 {
		switch b := t.Body.(type) {
		case *ParamBody:
			return b.Name
		case *ConstBody:
			return b.Name
		case *RefBody:
			return b.Ref
		}
	}
	return ""
}

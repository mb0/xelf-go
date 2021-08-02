package knd

import (
	"fmt"
	"math/bits"
)

// Kind is a bitset describing a language element.
type Kind uint32

const (
	// modifier bits
	None Kind = 1 << iota
	Some

	// exp bits
	Lit
	Typ
	Sym
	Tag
	Tupl
	Call

	// prim bits
	Bool

	Int
	Real
	Bits

	Str
	Raw
	UUID
	Span
	Time
	Enum

	// cont
	List
	Dict
	Rec
	Obj

	// spec
	Func
	Form

	// meta
	Alt
	Var
	Ref
	Sel
)

const (
	Void = Kind(0)
	Exp  = Lit | Sym | Tag | Tupl | Call
	Schm = Bits | Enum | Obj
	Meta = Alt | Var | Ref | Sel

	Num  = Int | Real | Bits
	Char = Str | Raw | UUID | Span | Time | Enum
	Prim = Bool | Num | Char
	Cont = List | Dict
	Strc = Rec | Obj
	Idxr = List | Strc
	Keyr = Dict | Strc
	Data = Prim | Cont | Strc
	Spec = Func | Form
	All  = Data | Typ | Spec
	Any  = All | None
)

var ErrInvalid = fmt.Errorf("invalid")

// Parse returns the kind for str or an error.
func ParseName(str string) (Kind, error) {
	k, ok := strToKind[str]
	if !ok {
		return Void, ErrInvalid
	}
	return k, nil
}

// Name returns the simple name of kind k or an empty string.
func Name(k Kind) string { return kindToStr[k] }

func (k Kind) IsAlt() bool { return k&Alt != 0 || (k&Data).Count() > 1 }
func (k Kind) Count() int  { return bits.OnesCount32(uint32(k)) }

type Info struct {
	Name string
	Kind Kind
}

// Infos is a list of all named kinds.
var Infos = []Info{
	{"void", Void},
	{"none", None},
	{"some", Some},
	{"lit", Lit},
	{"typ", Typ},
	{"sym", Sym},
	{"tag", Tag},
	{"tupl", Tupl},
	{"call", Call},
	{"bool", Bool},
	{"int", Int},
	{"real", Real},
	{"bits", Bits},
	{"str", Str},
	{"raw", Raw},
	{"uuid", UUID},
	{"span", Span},
	{"time", Time},
	{"enum", Enum},
	{"list", List},
	{"dict", Dict},
	{"rec", Rec},
	{"obj", Obj},
	{"func", Func},
	{"form", Form},
	{"alt", Alt},
	{"var", Var},
	{"ref", Ref},
	{"sel", Sel},

	{"exp", Exp},
	{"schm", Schm},
	{"meta", Meta},

	{"num", Num},
	{"char", Char},
	{"prim", Prim},
	{"cont", Cont},
	{"strc", Strc},
	{"idxr", Idxr},
	{"keyr", Keyr},
	{"data", Data},
	{"spec", Spec},
	{"all", All},
	{"any", Any},
}

var kindToStr map[Kind]string
var strToKind map[string]Kind

func init() {
	kindToStr = make(map[Kind]string, len(Infos))
	strToKind = make(map[string]Kind, len(Infos))
	for _, n := range Infos {
		kindToStr[n.Kind] = n.Name
		strToKind[n.Name] = n.Kind
	}
}

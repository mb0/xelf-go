package lib

import (
	"fmt"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/typ"
)

// Core is a builtin environment with foundational specs.
var Core = exp.Builtins(make(Specs).Add(
	Or, And, Ok, Not, Err,
	Add, Sub, Mul, Div, Rem, Abs, Neg, Min, Max,
	Eq, Ne, Lt, Ge, Gt, Le, In, Ni, Equal,
	If, Swt, Df,
))

// Specs is spec map helper that can be converted to a builtin environment.
type Specs map[string]exp.Spec

// Add add all specs to this spec map.
func (sm Specs) Add(ss ...exp.Spec) Specs {
	for _, s := range ss {
		k := exp.SigName(s.Type())
		sm[k] = s
	}
	return sm
}

// AddMap add all specs for map m to this spec map.
func (sm Specs) AddMap(m map[string]exp.Spec) Specs {
	for k, s := range m {
		sm[k] = s
	}
	return sm
}

func impl(sig string) exp.SpecBase {
	t, err := typ.Parse(sig)
	if err != nil {
		panic(fmt.Errorf("impl sig %s: %v", sig, err))
	}
	return exp.SpecBase{Decl: t}
}

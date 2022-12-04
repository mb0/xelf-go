package lib

import (
	"xelf.org/xelf/exp"
)

// Core is a builtin environment with foundational specs.
var Core = exp.Builtins(make(Specs).Add(
	Or, And, Ok, Not, Err,
	Add, Sub, Mul, Div, Rem, Abs, Neg, Min, Max,
	Eq, Ne, Lt, Ge, Gt, Le, In, Ni, Equal,
	If, Swt, Df,
	Cat, Sep, Xelf, Json,
	Make, Sel, Len,
))

// Std extends the core environment with commonly used specs.
var Std = exp.Builtins(make(Specs).AddMap(Core).Add(
	Do, Dyn,
	Let, With,
	Mut, Append,
	Fn,
	Fold, Foldr, Range,
))

// Specs is spec map helper that can be converted to a builtin environment.
type Specs map[string]exp.Spec

// Add add all specs to this spec map.
func (sm Specs) Add(ss ...exp.Spec) Specs {
	for _, s := range ss {
		k := s.Type().Ref
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

var impl = exp.MustSpecBase

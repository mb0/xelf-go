package exp

import (
	"fmt"

	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Env is a scoped context to resolve symbols. Envs configure most of the program resolution.
type Env interface {
	// Parent returns the parent environment or nil.
	Parent() Env

	// Resl resolves a part of a symbol and returns the result or an error.
	Resl(p *Prog, s *Sym, k string, hint typ.Type) (Exp, error)

	// Eval evaluates a part of a symbol and returns a literal or an error.
	Eval(p *Prog, s *Sym, k string) (*Lit, error)
}

// Spec is a literal value for func or form specification used to resolve calls.
type Spec interface {
	lit.Val

	// Resl resolves a call using a type hint and returns the result or an error.
	Resl(p *Prog, env Env, c *Call, hint typ.Type) (Exp, error)

	// Eval evaluates a resolved call and returns a literal or an error.
	Eval(p *Prog, env Env, c *Call) (*Lit, error)
}

// Prog is the entry context to resolve an expression in an environment.
// Programs are bound to their expression and cannot be reused.
type Prog struct {
	Reg  *lit.Reg
	Exp  Exp
	Root Env
}

// NewProg returns a new program using the given registry, environment and expression.
// The registry argument can be nil, a new registry will be used by default.
func NewProg(reg *lit.Reg, env Env, exp Exp) *Prog {
	if reg == nil {
		reg = &lit.Reg{}
	}
	return &Prog{Reg: reg, Root: env, Exp: exp}
}

// Resl resolves an expression using a type hint and returns the result or an error.
func (p *Prog) Resl(env Env, e Exp, h typ.Type) (Exp, error) {
	return e, nil
}

// Eval evaluates a resolved expression and returns a literal or an error.
func (p *Prog) Eval(env Env, e Exp) (Exp, error) {
	return e, fmt.Errorf("not yet implemented")
}

package exp

import (
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Spec is a literal value for func or form specification used to resolve calls.
type Spec interface {
	lit.Val

	// Resl resolves a call using a type hint and returns the result or an error.
	Resl(p *Prog, env Env, c *Call, hint typ.Type) (Exp, error)

	// Eval evaluates a resolved call and returns a literal or an error.
	Eval(p *Prog, env Env, c *Call) (*Lit, error)
}

// SpecBase is partial base definition for spec implementations.
// Implementations only need embed this type and declare the value and eval methods.
type SpecBase struct{ Decl typ.Type }

func (s *SpecBase) Type() typ.Type               { return s.Decl }
func (s *SpecBase) Nil() bool                    { return s == nil }
func (s *SpecBase) Zero() bool                   { return false }
func (s *SpecBase) String() string               { return s.Decl.String() }
func (s *SpecBase) Print(p *bfr.P) error         { return s.Decl.Print(p) }
func (s *SpecBase) MarshalJSON() ([]byte, error) { return s.Decl.MarshalJSON() }

func (s *SpecBase) Resl(p *Prog, env Env, c *Call, h typ.Type) (Exp, error) {
	if c.Env == nil {
		c.Env = env
	}
	ps := SigArgs(c.Sig)
	n := len(ps)
	vari := s.Decl.Kind&knd.Spec == knd.Func && n > 0 && ps[n-1].Kind&knd.List != 0
	for i, pa := range ps {
		a := c.Args[i]
		if a != nil {
			ah := pa.Type
			if ah.Kind&knd.Tupl != 0 {
				ah, _ = tuplEl(ah)
			}
			if vari && i == n-1 {
				if _, ok := a.(*Tupl); ok {
					ah, _ = tuplEl(ah)
				}
			}
			e, err := p.Resl(c.Env, a, ah)
			if err != nil {
				return c, err
			}
			c.Args[i] = e
		}
	}
	rp := SigRes(c.Sig)
	if h != typ.Void {
		_, err := p.Sys.Unify(rp.Type, h)
		if err != nil {
			return c, err
		}
	}
	c.Sig = p.Sys.Update(c.Sig)
	return c, nil
}

func tuplEl(t typ.Type) (typ.Type, int) {
	b, ok := t.Body.(*typ.ParamBody)
	if !ok || len(b.Params) == 0 {
		return typ.Exp, 0
	}
	n := len(b.Params)
	if n == 1 {
		return b.Params[0].Type, 1
	}
	return typ.Type{Kind: knd.Rec, Body: b}, n
}

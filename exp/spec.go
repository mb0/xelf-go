package exp

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Spec is a func or form specification used to resolve calls.
type Spec interface {
	Type() typ.Type

	// Resl resolves a call using a type hint and returns the result or an error.
	// The first call to Resl should setup the environment for call c.
	Resl(p *Prog, env Env, c *Call, hint typ.Type) (Exp, error)

	// Eval evaluates a resolved call and returns a value or an error.
	Eval(p *Prog, c *Call) (lit.Val, error)
}

// SpecBase is partial base definition for spec implementations.
// Implementations only need embed this type and declare the eval method.
type SpecBase struct{ Decl typ.Type }

func MustSpecBase(sig string) SpecBase {
	decl, err := typ.Parse(sig)
	if err != nil {
		panic(fmt.Errorf("must parse signature %s: %v", sig, err))
	}
	return SpecBase{decl}
}

func (s *SpecBase) Type() typ.Type { return s.Decl }
func (s *SpecBase) Resl(p *Prog, env Env, c *Call, h typ.Type) (Exp, error) {
	if c.Env == nil {
		c.Env = env
	}
	ps := SigArgs(c.Sig)
	n := len(ps)
	vari := s.Decl.Kind&knd.Spec == knd.Func && n > 0 && ps[n-1].Kind&knd.List != 0
	for i, pa := range ps {
		a := c.Args[i]
		if a == nil {
			continue
		}
		ah := pa.Type
		if vari && i == n-1 {
			if _, ok := a.(*Tupl); ok {
				ah, _ = typ.TuplEl(ah)
			}
		}
		e, err := p.Resl(c.Env, a, ah)
		if err != nil {
			return c, err
		}
		c.Args[i] = e
	}
	rp := SigRes(c.Sig)
	ut, err := p.Sys.Unify(rp.Type, h)
	if err != nil {
		return c, err
	}
	rp.Type = ut
	c.Sig, err = p.Sys.Update(c.Sig)
	return c, err
}

func UnwrapSpec(v interface{}) (s *SpecRef) {
	for s, _ = v.(*SpecRef); s != nil; {
		if r, _ := s.Spec.(*SpecRef); r == nil {
			return s
		} else {
			s = r
		}
	}
	return s
}

// SpecRef can wrap any spec with a new type and can represent null and unresolved specs.
type SpecRef struct {
	Spec Spec
	Decl typ.Type
}

func NewSpecRef(s Spec) *SpecRef { return &SpecRef{Spec: s, Decl: s.Type()} }

func (s *SpecRef) Nil() bool  { return s == nil }
func (s *SpecRef) Zero() bool { return s == nil || s.Spec == nil }

func (s *SpecRef) Type() typ.Type { return s.Decl }
func (s *SpecRef) Mut() lit.Mut   { return s }
func (s *SpecRef) Value() lit.Val { return s }
func (s *SpecRef) As(t typ.Type) (lit.Val, error) {
	if s.Spec.Type().AssignableTo(t) {
		s.Decl = t
		return s, nil
	}
	return nil, fmt.Errorf("cannot convert %T from %s to %s", s, s.Type(), t)
}

func (s *SpecRef) String() string               { return s.Decl.String() }
func (s *SpecRef) Print(p *bfr.P) error         { return s.Decl.Print(p) }
func (s *SpecRef) MarshalJSON() ([]byte, error) { return s.Decl.MarshalJSON() }

func (s *SpecRef) New() lit.Mut     { return &SpecRef{Decl: s.Decl, Spec: nil} }
func (s *SpecRef) Ptr() interface{} { return s }
func (s *SpecRef) Parse(a ast.Ast) (err error) {
	return ast.ErrInvalid(a, knd.Spec, fmt.Errorf("cannot parse into spec values"))
}
func (s *SpecRef) Assign(val lit.Val) error {
	switch v := lit.Unwrap(val).(type) {
	case lit.Null:
		s.Spec = nil
		return nil
	case *SpecRef:
		if v.Decl.AssignableTo(s.Decl) {
			s.Spec = v.Spec
			return nil
		}
	}
	return fmt.Errorf("cannot assign %s to spec value", val.Type())
}
func (s *SpecRef) Resl(p *Prog, env Env, c *Call, hint typ.Type) (Exp, error) {
	if !s.Nil() {
		return s.Spec.Resl(p, env, c, hint)
	}
	return c, ast.ErrReslSpec(c.Src, c.Sig.Ref, fmt.Errorf("spec undefined"))
}
func (s *SpecRef) Eval(p *Prog, c *Call) (lit.Val, error) {
	if !s.Nil() {
		return s.Spec.Eval(p, c)
	}
	return nil, ast.ErrEval(c.Src, c.Sig.Ref, fmt.Errorf("spec undefined"))
}

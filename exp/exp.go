package exp

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Exp is the common interface of all expressions with kind, type and source info.
type Exp interface {
	Type() typ.Type
	Source() ast.Src
	String() string
	Print(*bfr.P) error
	Clone() Exp
}

// Lit is a literal expression with a literal value, which may include a type or spec.
type Lit struct {
	Val lit.Val
	Src ast.Src
}

func LitVal(v lit.Val) *Lit { return LitSrc(v, ast.Src{}) }
func LitSrc(v lit.Val, s ast.Src) *Lit {
	return &Lit{Val: v, Src: s}
}

func (a *Lit) Type() typ.Type {
	t := a.Val.Type()
	return typ.Type{Kind: knd.Lit, Body: &t}
}
func (a *Lit) Source() ast.Src { return a.Src }
func (a *Lit) String() string {
	if a.Val == nil {
		return "null"
	}
	return a.Val.String()
}
func (a *Lit) Print(p *bfr.P) error {
	if a.Val == nil {
		return p.Fmt("null")
	}
	return a.Val.Print(p)
}
func (a *Lit) Clone() Exp {
	v, err := lit.Clone(a.Val)
	if err != nil {
		panic(fmt.Errorf("unhandled err in lit clone:\n%w", err))
	}
	return &Lit{v, a.Src}
}
func (a *Lit) Value() lit.Val {
	if a == nil || a.Val == nil {
		return lit.Null{}
	}
	return a.Val.Value()
}

// Sym is a symbol expression which caches the resolving environment and a relative name.
type Sym struct {
	Res  typ.Type
	Sym  string
	Src  ast.Src
	Env  Env
	Path cor.Path
}

func (s *Sym) Type() typ.Type       { return typ.Type{Kind: knd.Sym, Body: &s.Res} }
func (s *Sym) Source() ast.Src      { return s.Src }
func (s *Sym) String() string       { return s.Sym }
func (s *Sym) Print(p *bfr.P) error { return p.Fmt(s.Sym) }
func (s *Sym) Clone() Exp           { return &Sym{s.Res, s.Sym, s.Src, nil, nil} }
func (s *Sym) Update(t typ.Type, env Env, p cor.Path) {
	s.Res, s.Env, s.Path = t, env, p
}

// Tag is a named quasi expression that is resolved by its parent call.
type Tag struct {
	Tag string
	Exp Exp
	Src ast.Src
}

func (t *Tag) Type() typ.Type {
	if t.Exp == nil {
		return typ.Tag
	}
	r := typ.Res(t.Exp.Type())
	return typ.Type{Kind: knd.Tag, Body: &r}
}
func (t *Tag) Source() ast.Src { return t.Src }
func (t *Tag) String() string  { return bfr.String(t) }
func (t *Tag) Print(p *bfr.P) error {
	p.Fmt(t.Tag)
	if t.Exp == nil {
		return p.Byte(';')
	}
	p.Byte(':')
	return t.Exp.Print(p)
}
func (t *Tag) Clone() Exp { return &Tag{t.Tag, t.Exp.Clone(), t.Src} }

// Tupl is a quasi multi-expression that is resolved by its parent call or a program.
type Tupl struct {
	Res typ.Type
	Els []Exp
	Src ast.Src
}

func (t *Tupl) Type() typ.Type  { return t.Res }
func (t *Tupl) Source() ast.Src { return t.Src }
func (t *Tupl) String() string  { return bfr.String(t) }
func (t *Tupl) Print(p *bfr.P) error {
	for i, e := range t.Els {
		if i != 0 {
			p.Byte(' ')
		}
		err := e.Print(p)
		if err != nil {
			return err
		}
	}
	return nil
}
func (t *Tupl) Clone() Exp {
	els := append(([]Exp)(nil), t.Els...)
	for i, e := range els {
		els[i] = e.Clone()
	}
	return &Tupl{t.Res, els, t.Src}
}

// Call is an executable expression that uses a spec to evaluate to a literal.
// It caches the resolved spec and environment.
type Call struct {
	Sig  typ.Type
	Spec Spec
	Args []Exp
	Env  Env
	Src  ast.Src
}

func (c *Call) Type() typ.Type {
	res := SigRes(c.Sig)
	if res == nil {
		return typ.Call
	}
	return typ.Type{Kind: knd.Call, Body: &res.Type}
}
func (c *Call) Source() ast.Src { return c.Src }
func (c *Call) String() string  { return bfr.String(c) }
func (c *Call) Print(p *bfr.P) error {
	p.Byte('(')
	name := c.Sig.Ref
	if name != "" {
		p.Fmt(name)
		p.Byte(' ')
	}
	for i, a := range c.Args {
		if i != 0 {
			p.Byte(' ')
		}
		err := a.Print(p)
		if err != nil {
			return err
		}
	}
	return p.Byte(')')
}
func (c *Call) Clone() Exp {
	args := append(([]Exp)(nil), c.Args...)
	for i, a := range args {
		args[i] = a.Clone()
	}
	return &Call{c.Sig, c.Spec, args, nil, c.Src}
}

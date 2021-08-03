package exp

import (
	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Exp is the common interface of all expressions with kind, type and source info.
type Exp interface {
	Kind() knd.Kind
	Resl() typ.Type
	Source() ast.Src
	String() string
	Print(*bfr.P) error
}

// Lit is a literal expression with a literal value, which may include a type or spec.
type Lit struct {
	Res typ.Type
	lit.Val
	Src ast.Src
}

func (a *Lit) Kind() knd.Kind  { return knd.Lit }
func (a *Lit) Resl() typ.Type  { return a.Res }
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

// Sym is a symbol expression which caches the resolving environment and a relative name.
type Sym struct {
	Type typ.Type
	Sym  string
	Src  ast.Src
	Env  Env
	Rel  string
}

func (s *Sym) Kind() knd.Kind       { return knd.Sym }
func (s *Sym) Resl() typ.Type       { return s.Type }
func (s *Sym) Source() ast.Src      { return s.Src }
func (s *Sym) String() string       { return s.Sym }
func (s *Sym) Print(p *bfr.P) error { return p.Fmt(s.Sym) }

// Tag is a named quasi expression that is resolved by its parent call.
type Tag struct {
	Tag string
	Exp Exp
	Src ast.Src
}

func (t *Tag) Kind() knd.Kind { return knd.Tag }
func (t *Tag) Resl() typ.Type {
	if t.Exp == nil {
		return typ.Void
	}
	return t.Exp.Resl()
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

// Tupl is a quasi multi-expression that is resolved by its parent call or a program.
type Tupl struct {
	Type typ.Type
	Els  []Exp
	Src  ast.Src
}

func (t *Tupl) Kind() knd.Kind  { return knd.Tupl }
func (t *Tupl) Resl() typ.Type  { return t.Type }
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

// Call is an executable expression that uses a spec to evaluate to a literal.
// It caches the resolved spec and environment.
type Call struct {
	Sig  typ.Type
	Spec Spec
	Args []Exp
	Env  Env
	Src  ast.Src
}

func (c *Call) Kind() knd.Kind { return knd.Call }
func (c *Call) Resl() (t typ.Type) {
	res := SigRes(c.Sig)
	if res == nil {
		return typ.Void
	}
	return res.Type
}
func (c *Call) Source() ast.Src { return c.Src }
func (c *Call) String() string  { return bfr.String(c) }
func (c *Call) Print(p *bfr.P) error {
	p.Byte('(')
	name := SigName(c.Sig)
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

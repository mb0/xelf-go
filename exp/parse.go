package exp

import (
	"io"
	"strconv"
	"strings"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

// Parse parses str and returns an expression or an error.
func Parse(reg *lit.Reg, str string) (Exp, error) { return Read(reg, strings.NewReader(str), "") }

// Read parses named reader r and returns an expression or an error.
func Read(reg *lit.Reg, r io.Reader, name string) (Exp, error) {
	a, err := ast.Read(r, name)
	if err != nil {
		return nil, err
	}
	if reg == nil {
		reg = &lit.Reg{}
	}
	return ParseAst(reg, a)
}

// ParseAst parses a as expression and returns it or an error.
func ParseAst(reg *lit.Reg, a ast.Ast) (Exp, error) {
	switch a.Kind {
	case knd.Int:
		n, err := strconv.ParseInt(a.Raw, 10, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Int, err)
		}
		return &Lit{Res: typ.Num, Val: lit.Int(n), Src: a.Src}, nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Real, err)
		}
		return &Lit{Res: typ.Real, Val: lit.Real(n), Src: a.Src}, nil
	case knd.Str:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Str, err)
		}
		return &Lit{Res: typ.Char, Val: lit.Str(txt), Src: a.Src}, nil
	case knd.Sym:
		switch a.Raw {
		case "null":
			return &Lit{Res: typ.None, Val: lit.Null{}, Src: a.Src}, nil
		case "false", "true":
			return &Lit{Res: typ.Bool, Val: lit.Bool(len(a.Raw) == 4), Src: a.Src}, nil
		}
		return &Sym{Sym: a.Raw, Src: a.Src}, nil
	case knd.List:
		list := &lit.List{Reg: reg}
		if err := list.Parse(a); err != nil {
			return nil, err
		}
		return &Lit{Res: typ.List, Val: list, Src: a.Src}, nil
	case knd.Dict:
		dict := &lit.Dict{Reg: reg}
		if err := dict.Parse(a); err != nil {
			return nil, err
		}
		return &Lit{Res: typ.Keyr, Val: dict, Src: a.Src}, nil
	case knd.Tag:
		if len(a.Seq) == 0 {
			return nil, ast.ErrInvalidTag(a.Tok)
		}
		t := a.Seq[0]
		tag := t.Raw
		var err error
		if t.Kind == knd.Str {
			tag, err = cor.Unquote(a.Raw)
			if err != nil {
				return nil, ast.ErrInvalid(a, knd.Str, err)
			}
		}
		var e Exp
		if len(a.Seq) > 1 {
			e, err = ParseAst(reg, a.Seq[1])
			if err != nil {
				return nil, err
			}
		}
		return &Tag{Tag: tag, Exp: e, Src: a.Src}, nil
	case knd.Call:
		if len(a.Seq) == 0 {
			return &Call{Src: a.Src}, nil
		} else if fst := a.Seq[0]; fst.Kind&(knd.Typ|knd.Call) != 0 && len(fst.Seq) == 0 {
			return &Call{Src: a.Src}, nil
		}
		res := &Call{Src: a.Src, Args: make([]Exp, 0, len(a.Seq))}
		for _, e := range a.Seq {
			el, err := ParseAst(reg, e)
			if err != nil {
				return nil, err
			}
			if el.Kind() != knd.Call || len(el.(*Call).Args) != 0 {
				res.Args = append(res.Args, el)
			}
		}
		return res, nil
	case knd.Typ:
		t, err := typ.ParseAst(a)
		if err != nil {
			return nil, err
		}
		return &Lit{Res: typ.Typ, Val: t, Src: a.Src}, nil
	}
	return nil, ast.ErrUnexpected(a)
}

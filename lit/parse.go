package lit

import (
	"io"
	"strconv"
	"strings"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// Read parses from r and returns a literal or an error.
func Read(r io.Reader, name string) (*Lit, error) {
	a, err := ast.Read(r, name)
	if err != nil {
		return nil, err
	}
	return ParseAst(a)
}

// Parse parses s and returns a literal or an error.
func Parse(s, name string) (*Lit, error) {
	return Read(strings.NewReader(s), name)
}

// ParseAst parses a as literal and returns it or an error.
func ParseAst(a ast.Ast) (*Lit, error) {
	switch a.Kind {
	case knd.Num:
		n, err := strconv.ParseInt(a.Raw, 10, 64)
		if err != nil {
			return nil, err
		}
		return &Lit{Res: typ.Num, Val: Int(n), Src: a.Src}, nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, err
		}
		return &Lit{Res: typ.Real, Val: Real(n), Src: a.Src}, nil
	case knd.Char:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, err
		}
		return &Lit{Res: typ.Char, Val: Str(txt), Src: a.Src}, nil
	case knd.Sym:
		if s := ParseSym(a.Raw, a.Src); s != nil {
			return s, nil
		}
	case knd.List:
		li := &List{Vals: make([]Val, 0, len(a.Seq))}
		res := &Lit{Res: typ.ListOf(typ.Any), Src: a.Src, Val: li}
		for _, e := range a.Seq {
			el, err := ParseAst(e)
			if err != nil {
				return nil, err
			}
			li.Vals = append(li.Vals, el.Val)
		}
		return res, nil
	case knd.Keyr:
		di := &Dict{Keyed: make([]KeyVal, 0, len(a.Seq))}
		res := &Lit{Res: typ.KeyrOf(typ.Any), Src: a.Src, Val: di}
		for _, e := range a.Seq {
			if e.Kind != knd.Tag || len(e.Seq) < 2 {
				return nil, ast.ErrExpectTag(e)
			}
			a, b := e.Seq[0], e.Seq[1]
			key := a.Raw
			if a.Kind == knd.Char {
				var err error
				key, err = cor.Unquote(key)
				if err != nil {
					return nil, ast.ErrUnquote(a.Tok, err)
				}
			}
			el, err := ParseAst(b)
			if err != nil {
				return nil, err
			}
			di.Keyed = append(di.Keyed, KeyVal{key, el.Val})
		}
		return res, nil
	}
	return nil, ast.ErrUnexpected(a)
}

func ParseSym(raw string, src ast.Src) *Lit {
	switch raw {
	case "null":
		return &Lit{Res: typ.None, Val: Null{}, Src: src}
	case "false", "true":
		return &Lit{Res: typ.Bool, Val: Bool(len(raw) == 4), Src: src}
	}
	return nil
}

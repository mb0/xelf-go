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
func Parse(str string) (Exp, error) {
	return Read(strings.NewReader(str), "")
}

// Read parses named reader r and returns an expression or an error.
func Read(r io.Reader, name string) (Exp, error) {
	as, err := ast.ReadAll(r, name)
	if err != nil {
		return nil, err
	}
	return ParseAll(as)
}

func ParseAll(as []ast.Ast) (Exp, error) {
	switch len(as) {
	case 0:
		return LitVal(lit.Null{}), nil
	case 1:
		return ParseAst(as[0])
	default:
		seq := make([]ast.Ast, 0, len(as)+1)
		seq = append(seq, ast.Ast{Tok: ast.Tok{Kind: knd.Sym, Raw: "do"}})
		seq = append(seq, as...)
		return ParseAst(ast.Ast{Tok: ast.Tok{Kind: knd.Call, Rune: '('}, Seq: seq})
	}
}

// ParseAst parses a as expression and returns it or an error.
func ParseAst(a ast.Ast) (Exp, error) {
	switch a.Kind {
	case knd.Num:
		n, err := strconv.ParseInt(a.Raw, 10, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Num, err)
		}
		return LitSrc(lit.Num(n), a.Src), nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Real, err)
		}
		return LitSrc(lit.Real(n), a.Src), nil
	case knd.Char:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Char, err)
		}
		return LitSrc(lit.Char(txt), a.Src), nil
	case knd.Sym:
		switch a.Raw {
		case "null":
			return LitSrc(lit.Null{}, a.Src), nil
		case "false", "true":
			return LitSrc(lit.Bool(len(a.Raw) == 4), a.Src), nil
		}
		return &Sym{Sym: a.Raw, Src: a.Src}, nil
	case knd.Idxr:
		vals := &lit.Vals{}
		if err := vals.Parse(a); err != nil {
			return nil, err
		}
		return LitSrc(vals, a.Src), nil
	case knd.Keyr:
		keyed := &lit.Keyed{}
		if err := keyed.Parse(a); err != nil {
			return nil, err
		}
		return LitSrc(keyed, a.Src), nil
	case knd.Tag:
		if len(a.Seq) == 0 {
			return nil, ast.ErrInvalidTag(a.Tok)
		}
		t := a.Seq[0]
		tag := t.Raw
		var err error
		if t.Kind == knd.Char {
			tag, err = cor.Unquote(a.Raw)
			if err != nil {
				return nil, ast.ErrInvalid(a, knd.Char, err)
			}
		}
		var e Exp
		if len(a.Seq) > 1 {
			e, err = ParseAst(a.Seq[1])
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
			el, err := ParseAst(e)
			if err != nil {
				return nil, err
			}
			if e.Kind != knd.Call || len(el.(*Call).Args) != 0 {
				res.Args = append(res.Args, el)
			}
		}
		return res, nil
	case knd.Typ:
		t, err := typ.ParseAst(a)
		if err != nil {
			return nil, err
		}
		return LitSrc(lit.Wrap(&t, typ.VarTyp), a.Src), nil
	}
	return nil, ast.ErrUnexpected(a)
}

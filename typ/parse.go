package typ

import (
	"io"
	"strconv"
	"strings"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/knd"
)

// Parse parses str and returns a type or an error.
func Parse(str string) (Type, error) { return Read(strings.NewReader(str), "") }

// Read parses named reader r and returns a type or an error.
func Read(r io.Reader, name string) (Type, error) {
	a, err := ast.Read(r, name)
	if err != nil {
		return Void, err
	}
	return ParseAst(a)
}

// ParseAst parses a as type and returns it or an error.
func ParseAst(a ast.Ast) (Type, error) { return parse(a, nil) }
func parse(a ast.Ast, hist []Type) (Type, error) {
	switch a.Kind {
	case knd.Sym:
		return ParseSym(a.Raw, a.Src, hist)
	case knd.Typ:
		if len(a.Seq) == 0 { // empty expression is void
			return Void, nil
		}
		fst := a.Seq[0]
		if fst.Kind != knd.Sym {
			return Void, ast.ErrUnexpected(a)
		}
		t, err := ParseSym(fst.Raw, fst.Src, hist)
		if err != nil {
			return t, err
		}
		return parseBody(a, a.Seq[1:], t, hist)
	}
	return Void, ast.ErrUnexpected(a)
}

func ParseSym(raw string, src ast.Src, hist []Type) (Type, error) {
	var res Type
	sp := strings.SplitN(raw, "|", -1)
	var s, v string
	for i := len(sp) - 1; i >= 0; i-- {
		s, v = sp[i], ""
		if s == "" {
			return Void, ast.ErrInvalidType(src, raw)
		}
		var tk knd.Kind
		lst := s[len(s)-1]
		none := lst == '?'
		some := !none && lst == '!'
		if none || some {
			s = s[:len(s)-1]
			if some {
				tk |= knd.Some
			} else {
				tk |= knd.None
			}
		}
		var tid int32
		var tr string
		var tb Body
		vi := strings.IndexByte(s, '@')
		if vi >= 0 {
			s, v = s[:vi], s[vi+1:]
			if len(v) == 0 {
				tk |= knd.Var
				tid = -1
			} else if r := v[0]; r >= '0' && r <= '9' {
				tk |= knd.Var
				pi := strings.IndexAny(v, "/.")
				if pi >= 0 {
					var vp string
					v, vp = v[:pi], v[pi:]
					tb = &SelBody{Path: vp}
				}
				id, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return Void, err
				}
				tid = int32(id)
			} else {
				tr = v
			}
		}
		if s != "" {
			if s[0] == '.' {
				// local ref
				tk |= knd.Sel
				tb = &SelBody{Path: s}
			} else if s[0] == '_' && (len(s) == 1 || s[1] == '.') {
				tk |= knd.Sel
				tb = &SelBody{Path: ".0" + s[1:]}
			} else {
				k, err := knd.ParseName(s)
				if err != nil {
					return Void, ast.ErrInvalidType(src, s)
				}
				tk |= k
			}
		}
		if tk&(knd.Exp|knd.Typ|knd.List|knd.Dict) != 0 {
			if res.Kind != 0 {
				tb = &ElBody{El: res}
			}
		}
		if tr != "" && tk&knd.All == 0 {
			tk |= knd.Ref
		}
		res = Type{tk, tid, tr, tb}
	}
	return res, nil
}

func parseBody(a ast.Ast, args []ast.Ast, t Type, hist []Type) (_ Type, err error) {
	if len(args) == 0 {
		return t, nil
	}
	el := &t
	eb, ok := el.Body.(*ElBody)
	for ok {
		el = &eb.El
		eb, ok = el.Body.(*ElBody)
	}
	switch el.Kind &^ (knd.Var | knd.Ref | knd.None) {
	case knd.Bits, knd.Enum:
		if len(args) > 0 {
			// TODO parse consts
		}
		el.Body = &ConstBody{}
		return t, nil
	case knd.Alt:
		alts := make([]Type, 0, len(args))
		for _, arg := range args {
			alt, err := parse(arg, hist)
			if err != nil {
				return Void, err
			}
			alts = append(alts, alt)
		}
		alt := Alt(alts...)
		el.Kind = alt.Kind | (el.Kind & knd.Var)
		el.Body = alt.Body
		return t, nil
	case knd.Obj, knd.Func, knd.Tupl, knd.Form:
	case knd.List, knd.Dict, knd.Typ:
		b, err := parse(args[0], hist)
		if err != nil {
			return Void, err
		}
		el.Body = &ElBody{El: b}
		return t, nil
	default:
		return Void, ast.ErrInvalidParams(a)
	}
	ps, err := parseParams(args, hist)
	if err != nil {
		return Void, err
	}
	if el.Kind&knd.Tupl != 0 && len(ps) == 1 && ps[0].Name == "" {
		el.Body = &ElBody{El: ps[0].Type}
	} else {
		el.Body = &ParamBody{Params: ps}
	}
	return t, nil
}

func parseName(a ast.Ast) (string, error) {
	if a.Kind != knd.Sym {
		return "", ast.ErrExpectSym(a)
	}
	return a.Raw, nil
}

func parseParams(args []ast.Ast, hist []Type) ([]Param, error) {
	res := make([]Param, 0, len(args))
	for len(args) > 0 {
		a := args[0]
		args = args[1:]
		var p Param
		if a.Kind == knd.Tag {
			name, err := parseName(a.Seq[0])
			if err != nil {
				return nil, err
			}
			p = P(name, Void)
			if len(a.Seq) > 1 {
				a = a.Seq[1]
			} else {
				res = append(res, p)
				continue
			}
		}
		b, err := parse(a, hist)
		if err != nil {
			return nil, err
		}
		p.Type = b
		res = append(res, p)
	}
	return res, nil
}

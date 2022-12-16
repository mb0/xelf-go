package typ

import (
	"io"
	"strconv"
	"strings"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/cor"
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
		return ParseSym(a.Raw, a.Src)
	case knd.Typ:
		if len(a.Seq) == 0 { // empty expression is void
			return Void, nil
		}
		fst := a.Seq[0]
		if fst.Kind != knd.Sym {
			return Void, ast.ErrUnexpected(a)
		}
		t, err := ParseSym(fst.Raw, fst.Src)
		if err != nil {
			return t, err
		}
		return parseBody(a, a.Seq[1:], t, hist)
	}
	return Void, ast.ErrUnexpected(a)
}

func ParseSym(raw string, src ast.Src) (Type, error) {
	var res Type
	sp := strings.Split(raw, "|")
	var s, v string
	for i := len(sp) - 1; i >= 0; i-- {
		s, v = sp[i], ""
		if s == "" {
			return Void, ast.ErrInvalidType(src, raw)
		}
		var r Type
		lst := s[len(s)-1]
		none := lst == '?'
		some := lst == '!'
		if none || some {
			s = s[:len(s)-1]
			if some {
				r.Kind = knd.Some
			} else {
				r.Kind = knd.None
			}
		}
		vi := strings.IndexByte(s, '@')
		if vi >= 0 {
			s, v = s[:vi], s[vi+1:]
			if len(v) == 0 {
				r.Kind |= knd.Var
				r.ID = -1
			} else if cor.Digit(rune(v[0])) {
				r.Kind |= knd.Var
				pi := strings.IndexAny(v, "/.")
				if pi >= 0 {
					v, r.Ref = v[:pi], v[pi:]
				}
				id, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return Void, err
				}
				r.ID = int32(id)
			} else {
				r.Ref = v
			}
		}
		if s != "" {
			if fst := s[0]; fst == '.' || fst == '_' && (len(s) == 1 || s[1] == '.') {
				// local ref cannot have a reference
				if r.Ref != "" {
					return Void, ast.ErrInvalidType(src, sp[i])
				}
				r.Kind |= knd.Sel
				r.Ref = s
				if fst == '_' {
					r.Ref = ".0" + s[1:]
				}
			} else {
				k, err := knd.ParseName(s)
				if err != nil {
					return Void, ast.ErrInvalidType(src, s)
				}
				r.Kind |= k
			}
		}
		if r.Ref != "" && r.Kind&(knd.All|knd.Sel) == 0 {
			r.Kind |= knd.Ref
		}
		if res.Kind != 0 {
			if isElKind(r.Kind) {
				tmp := res
				r.Body = &tmp
			} else {
				return Void, ast.ErrInvalidType(src, s)
			}
		}
		res = r
	}
	return res, nil
}

func isElKind(k knd.Kind) bool {
	if k&knd.Exp == 0 {
		switch k & knd.All {
		case knd.Cont, knd.List, knd.Dict:
		case knd.Typ, knd.Spec:
		default:
			return false
		}
	}
	return true
}

func parseBody(a ast.Ast, args []ast.Ast, t Type, hist []Type) (_ Type, err error) {
	if len(args) == 0 {
		return t, nil
	}
	el := &t
	eb, ok := el.Body.(*Type)
	for ok {
		el = eb
		eb, ok = el.Body.(*Type)
	}
	switch el.Kind &^ (knd.Var | knd.Ref | knd.None) {
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
	case knd.Obj, knd.Func, knd.Tupl, knd.Form:
		ps, err := parseParams(args, hist)
		if err != nil {
			return Void, err
		}
		if el.Kind&knd.Tupl != 0 && len(ps) == 1 && ps[0].Name == "" {
			el.Body = &ps[0].Type
		} else {
			el.Body = &ParamBody{Params: ps}
		}
	case knd.Bits, knd.Enum:
		cs, err := parseConsts(args)
		if err != nil {
			return Void, err
		}
		el.Body = &ConstBody{Consts: cs}
	default:
		if len(args) > 1 {
			return Void, ast.ErrInvalidParams(a)
		}
		b, err := parse(args[0], hist)
		if err != nil {
			return Void, err
		}
		el.Body = &b
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
func parseConsts(args []ast.Ast) ([]Const, error) {
	res := make([]Const, 0, len(args))
	for len(args) > 0 {
		a := args[0]
		args = args[1:]
		var p Const
		if a.Kind == knd.Tag {
			name, err := parseName(a.Seq[0])
			if err != nil {
				return nil, err
			}
			p = C(name, -1)
			if len(a.Seq) > 1 {
				a = a.Seq[1]
			} else {
				res = append(res, p)
				continue
			}
		}
		if a.Kind == knd.Num {
			num, _ := strconv.ParseInt(a.Raw, 10, 64)
			p.Val = num
		}
		res = append(res, p)
	}
	return res, nil
}

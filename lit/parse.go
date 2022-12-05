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

// Parse parses str and returns a generic value or an error.
func Parse(str string) (Val, error) {
	return Read(strings.NewReader(str), "string")
}

// ParseInto parses str into mut or returns an error.
func ParseInto(str string, mut Mut) error {
	return ReadInto(strings.NewReader(str), "string", mut)
}

// Read parses named reader r and returns a generic value or an error.
func Read(r io.Reader, name string) (Val, error) {
	a, err := ast.Read(r, name)
	if err != nil {
		return nil, err
	}
	return ParseVal(a)
}

// ReadInto parses named reader r into mut or returns an error.
func ReadInto(r io.Reader, name string, mut Mut) error {
	a, err := ast.Read(r, name)
	if err != nil {
		return err
	}
	return mut.Parse(a)
}

// ParseVal parses a as generic value and returns it or an error.
func ParseVal(a ast.Ast) (v Val, err error) {
	switch a.Kind {
	case knd.Num:
		n, err := strconv.ParseInt(a.Raw, 10, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Num, err)
		}
		return Num(n), nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Real, err)
		}
		return Real(n), nil
	case knd.Char:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Char, err)
		}
		return Char(txt), nil
	case knd.Sym:
		switch a.Raw {
		case "null":
			return Null{}, nil
		case "false", "true":
			return Bool(len(a.Raw) == 4), nil
		}
	case knd.Idxr:
		vs := make(Vals, 0, len(a.Seq))
		for _, e := range a.Seq {
			el, err := ParseVal(e)
			if err != nil {
				return nil, err
			}
			vs = append(vs, el)
		}
		return &vs, nil
	case knd.Keyr:
		kvs := make(Keyed, 0, len(a.Seq))
		for _, e := range a.Seq {
			key, val, err := ast.UnquotePair(e)
			if err != nil {
				return nil, err
			}
			el, err := ParseVal(val)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, KeyVal{key, el})
		}
		return &kvs, nil
	case knd.Typ:
		t, err := typ.ParseAst(a)
		if err != nil {
			return nil, err
		}
		return t, nil
	}
	return nil, ast.ErrUnexpected(a)
}

// ParseMut parses a as mutable value and returns it or an error.
// If the null symbol is parsed nil mutable is returned.
func ParseMut(a ast.Ast) (Mut, error) {
	switch a.Kind {
	case knd.Num:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Num, err)
		}
		return (*NumMut)(&n), nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Real, err)
		}
		return (*RealMut)(&n), nil
	case knd.Char:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Char, err)
		}
		return (*CharMut)(&txt), nil
	case knd.Sym:
		switch a.Raw {
		case "null":
			return nil, nil
		case "false", "true":
			ok := BoolMut(len(a.Raw) == 4)
			return &ok, nil
		}
	case knd.Idxr:
		li := &Vals{}
		return li, li.Parse(a)
	case knd.Keyr:
		di := &Keyed{}
		return di, di.Parse(a)
	case knd.Typ:
		t, err := typ.ParseAst(a)
		if err != nil {
			return nil, err
		}
		return &t, nil
	}
	return nil, ast.ErrUnexpected(a)
}

func parseMutNull(a ast.Ast) (Val, error) {
	if a.Kind == knd.Void {
		return Null{}, nil
	}
	m, err := ParseMut(a)
	if m == nil {
		return Null{}, err
	}
	return m, err
}

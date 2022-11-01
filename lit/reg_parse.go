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
func Parse(reg *Reg, str string) (Val, error) {
	return Read(reg, strings.NewReader(str), "string")
}

// ParseInto parses str into ptr or returns an error.
func ParseInto(reg *Reg, str string, ptr interface{}) error {
	return ReadInto(reg, strings.NewReader(str), "string", ptr)
}

// ParseIntoMut parses str into mut or returns an error.
func ParseIntoMut(reg *Reg, str string, mut Mut) error {
	return ReadIntoMut(reg, strings.NewReader(str), "string", mut)
}

// Read parses named reader r and returns a generic value or an error.
func Read(reg *Reg, r io.Reader, name string) (Val, error) {
	a, err := ast.Read(r, name)
	if err != nil {
		return nil, err
	}
	return reg.ParseVal(a)
}

// ReadInto parses named reader r into ptr or returns an error.
func ReadInto(reg *Reg, r io.Reader, name string, ptr interface{}) error {
	mut, err := reg.Proxy(ptr)
	if err != nil {
		return err
	}
	return ReadIntoMut(reg, r, name, mut)
}

// ReadIntoMut parses named reader r into mut or returns an error.
func ReadIntoMut(reg *Reg, r io.Reader, name string, mut Mut) error {
	a, err := ast.Read(r, name)
	if err != nil {
		return err
	}
	return mut.Parse(a)
}

// ParseVal parses a as generic value and returns it or an error.
func (reg *Reg) ParseVal(a ast.Ast) (v Val, err error) {
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
	case knd.List:
		li := &List{Reg: reg}
		vs := make([]Val, 0, len(a.Seq))
		for _, e := range a.Seq {
			el, err := reg.ParseVal(e)
			if err != nil {
				return nil, err
			}
			vs = append(vs, el)
		}
		li.Vals = vs
		return li, nil
	case knd.Dict:
		di := &Dict{Reg: reg}
		kvs := make([]KeyVal, 0, len(a.Seq))
		for _, e := range a.Seq {
			key, val, err := ast.UnquotePair(e)
			if err != nil {
				return nil, err
			}
			el, err := reg.ParseVal(val)
			if err != nil {
				return nil, err
			}
			kvs = append(kvs, KeyVal{key, el})
		}
		di.Keyed = kvs
		return di, nil
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
func (reg *Reg) ParseMut(a ast.Ast) (Mut, error) {
	switch a.Kind {
	case knd.Num:
		n, err := strconv.ParseInt(a.Raw, 10, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Num, err)
		}
		return (*Num)(&n), nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Real, err)
		}
		return (*Real)(&n), nil
	case knd.Char:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Char, err)
		}
		return (*Char)(&txt), nil
	case knd.Sym:
		switch a.Raw {
		case "null":
			return nil, nil
		case "false", "true":
			ok := Bool(len(a.Raw) == 4)
			return &ok, nil
		}
	case knd.List:
		li := &List{Reg: reg}
		return li, li.Parse(a)
	case knd.Dict:
		di := &Dict{Reg: reg}
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

func (reg *Reg) parseMutNull(a ast.Ast) (Val, error) {
	m, err := reg.ParseMut(a)
	if m == nil {
		return Null{}, err
	}
	return m, err
}

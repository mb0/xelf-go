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
func (reg *Reg) Parse(str string) (Val, error) {
	return reg.Read(strings.NewReader(str), "string")
}

// ParseInto parses str into ptr or returns an error.
func (reg *Reg) ParseInto(str string, ptr interface{}) error {
	return reg.ReadInto(strings.NewReader(str), "string", ptr)
}

// ParseIntoMut parses str into mut or returns an error.
func (reg *Reg) ParseIntoMut(str string, mut Mut) error {
	return reg.ReadIntoMut(strings.NewReader(str), "string", mut)
}

// Read parses named reader r and returns a generic value or an error.
func (reg *Reg) Read(r io.Reader, name string) (Val, error) {
	a, err := ast.Read(r, name)
	if err != nil {
		return nil, err
	}
	return reg.ParseVal(a)
}

// ReadInto parses named reader r into ptr or returns an error.
func (reg *Reg) ReadInto(r io.Reader, name string, ptr interface{}) error {
	mut, err := reg.Proxy(ptr)
	if err != nil {
		return err
	}
	return reg.ReadIntoMut(r, name, mut)
}

// ReadIntoMut parses named reader r into mut or returns an error.
func (reg *Reg) ReadIntoMut(r io.Reader, name string, mut Mut) error {
	a, err := ast.Read(r, name)
	if err != nil {
		return err
	}
	return mut.Parse(a)
}

// ParseVal parses a as generic value and returns it or an error.
func (reg *Reg) ParseVal(a ast.Ast) (v Val, err error) {
	switch a.Kind {
	case knd.Int:
		n, err := strconv.ParseInt(a.Raw, 10, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Int, err)
		}
		return Int(n), nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Real, err)
		}
		return Real(n), nil
	case knd.Str:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Str, err)
		}
		return Str(txt), nil
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
	}
	return nil, ast.ErrUnexpected(a)
}

// ParseMut parses a as mutable value and returns it or an error.
// If the null symbol is parsed nil mutable is returned.
func (reg *Reg) ParseMut(a ast.Ast) (Mut, error) {
	switch a.Kind {
	case knd.Int:
		n, err := strconv.ParseInt(a.Raw, 10, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Int, err)
		}
		return (*Int)(&n), nil
	case knd.Real:
		n, err := strconv.ParseFloat(a.Raw, 64)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Real, err)
		}
		return (*Real)(&n), nil
	case knd.Str:
		txt, err := cor.Unquote(a.Raw)
		if err != nil {
			return nil, ast.ErrInvalid(a, knd.Str, err)
		}
		return (*Str)(&txt), nil
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
	}
	return nil, ast.ErrUnexpected(a)
}

// ParseLit parses a as generic literal and returns it or an error.
func (reg *Reg) ParseLit(a ast.Ast) (*Lit, error) {
	switch a.Kind {
	case knd.Int, knd.Real, knd.Str, knd.Sym, knd.List, knd.Dict:
		v, err := reg.ParseVal(a)
		if err != nil {
			return nil, err
		}
		return &Lit{Res: typ.Type{Kind: a.Kind}, Val: v, Src: a.Src}, nil
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

func ParseSym(raw string, src ast.Src) *Lit {
	switch raw {
	case "null":
		return &Lit{Res: typ.None, Val: Null{}, Src: src}
	case "false", "true":
		return &Lit{Res: typ.Bool, Val: Bool(len(raw) == 4), Src: src}
	}
	return nil
}

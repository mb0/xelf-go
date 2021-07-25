// Package ast provides a token lexer and ast scanner for xelf.
package ast

import (
	"io"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
)

type Ast struct {
	Tok
	Seq []Ast
}

func (n *Ast) Print(b *bfr.P) error {
	if len(n.Seq) == 0 {
		b.Fmt(n.Tok.String())
	} else if n.Kind == knd.Tag && len(n.Seq) > 1 {
		b.Fmt(n.Seq[1].String())
		b.Fmt(n.Seq[0].String())
		if len(n.Seq) > 2 {
			n.Seq[2].Print(b)
		}
	} else {
		for i, s := range n.Seq {
			if i != 0 {
				b.Byte(' ')
			}
			s.Print(b)
		}
	}
	return b.Err

}
func (n Ast) String() string {
	if len(n.Seq) == 0 {
		return n.Tok.String()
	}
	return bfr.String(&n)
}

// Read returns the next ast read from r or an error.
func Read(r io.Reader, name string) (Ast, error) { return Scan(NewLexer(r, name)) }

// Scan returns the next Ast from l or an error.
func Scan(l *Lexer) (Ast, error) {
	t, err := l.Tok()
	if err != nil {
		return Ast{}, err
	}
	return ScanRest(l, t)
}

func ScanRest(l *Lexer, t Tok) (Ast, error) {
	res := Ast{Tok: t}
	if t.Kind&(knd.List|knd.Keyr|knd.Typ|knd.Call) == 0 {
		return res, nil
	}
	_, end := parens(t.Kind)
	if end == 0 {
		return res, nil
	}
	t, err := l.Tok()
	if err != nil {
		return res, err
	}
	for t.Rune != end {
		switch t.Rune {
		case ':', ';':
			return res, ErrInvalidTag(t)
		case ',':
			return res, ErrInvalidSep(t)
		}
		a, err := ScanRest(l, t)
		if err != nil {
			return a, err
		}
		t, err = l.Tok()
		if err != nil {
			return res, err
		}
		if t.Kind == knd.Tag {
			switch a.Kind {
			case knd.Sym, knd.Char:
			default:
				return res, ErrInvalidTag(t)
			}
			tt := Ast{Tok: t}
			tt.Src.Pos = a.Src.Pos
			t, err = l.Tok()
			if err != nil {
				res.Seq = append(res.Seq, tt)
				return res, err
			}
			if tt.Rune == ';' || !valStart(t) {
				tt.Seq = []Ast{a}
			} else {
				b, err := ScanRest(l, t)
				if err != nil {
					return res, err
				}
				tt.Seq = []Ast{a, b}
				tt.Src.End = b.Src.End
				t, err = l.Tok()
				if err != nil {
					return res, err
				}
			}
			res.Seq = append(res.Seq, tt)
		} else {
			res.Seq = append(res.Seq, a)
		}
		switch t.Rune {
		case ',':
			t, err = l.Tok()
			if err != nil {
				return res, err
			}
		}
	}
	res.Src.End = t.Src.End
	if t.Rune != end {
		return res, ErrTreeTerm(res.Tok)
	}
	return res, nil
}

func valStart(t Tok) bool {
	if t.Kind&(knd.Num|knd.Char|knd.Sym) != 0 {
		return true
	}
	switch t.Rune {
	case '[', '{', '<', '(':
		return true
	}
	return false
}

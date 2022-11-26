package ast

import (
	"log"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
)

type Formatter interface {
	Format(bfr.Writer, Ast) error
}

const (
	FmtMin uint16 = 1 << iota
	FmtMax
	FmtLst
)

type SimpleFormat struct {
	output *countWriter
	*bfr.P
	Flags, Max uint16
}

func (f *SimpleFormat) Format(w bfr.Writer, a Ast) error {
	f.output = &countWriter{W: w}
	f.P = &bfr.P{Writer: f.output, Plain: true}
	if f.Flags&^FmtMin != 0 {
		f.Tab = "\t"
	}
	return f.format(a, 1)
}
func (f *SimpleFormat) format(a Ast, isend byte) error {
	if a.Kind == knd.Tag && len(a.Seq) > 1 {
		f.Fmt(a.Seq[0].Tok.String())
		f.Fmt(a.Tok.String())
		if len(a.Seq) < 2 {
			return f.Err
		}
		a = a.Seq[1]
	}
	if len(a.Seq) == 0 {
		return f.Fmt(a.Tok.String())
	}
	if a.Kind == knd.Typ {
		f.Byte('<')
		for i, s := range a.Seq {
			if i != 0 {
				f.Byte(' ')
			}
			s.Print(f.P)
		}
		return f.Byte('>')
	}
	start, end := parens(a.Kind)
	if a.Kind != knd.Call {
		f.WriteRune(start)
		for i, s := range a.Seq {
			if i != 0 {
				f.Byte(' ')
			}
			f.format(s, 0)
		}
		f.WriteRune(end)
		return f.Err
	}
	if f.Flags&FmtLst != 0 {
		return f.formatLst(a, isend)
	}
	f.WriteRune(start)
	f.Depth++
	var lst Ast
	for i, s := range a.Seq {
		e := i == len(a.Seq)-1
		if i > 0 && f.Flags&FmtMax == 0 {
			if f.Flags&FmtMin == 0 || f.needsSpace(lst, s) {
				f.Byte(' ')
			}
		}
		if a.Kind != knd.Call {
			f.Break()
		}
		if e && isend > 0 {
			isend++
		}
		f.format(s, isend)
		if e {
			f.Dedent()
		} else if a.Kind == knd.Call {
			f.Break()
		}
		lst = s
	}
	f.WriteRune(end)
	return f.Err
}
func (f *SimpleFormat) formatLst(a Ast, isend byte) error {
	log.Printf("format last %s %v", a.Seq[0], isend)
	br := isend > 1
	if br {
		isend = 1
	}
	f.Byte('(')
	if br {
		f.Depth++
	}
	for i, s := range a.Seq {
		e := i == len(a.Seq)-1
		if !br && i > 0 {
			f.Byte(' ')
		}
		var eb byte
		if e && isend > 0 {
			eb = 1
			if !br {
				eb++
			}
		}
		f.format(s, eb)
		if br {
			if e {
				f.Depth--
			}
			f.Break()
		}
	}
	return f.Byte(')')
}

func (f *SimpleFormat) needsSpace(p, n Ast) bool {
	return n.Kind&(knd.Sym|knd.Num|knd.Tag) != 0 && (p.Kind&(knd.Sym|knd.Num) != 0 ||
		p.Kind&knd.Tag != 0 && (len(p.Seq) < 2 || p.Seq[1].Kind&(knd.Sym|knd.Num) != 0))
}

type countWriter struct {
	W       bfr.Writer
	O, L, C uint64 // offset, line, column
}

func (w *countWriter) count(r rune) {
	if w.O++; r == '\n' {
		w.L++
		w.C = 0
	} else {
		w.C++
	}
}
func (w *countWriter) WriteByte(b byte) error {
	w.count(rune(b))
	return w.W.WriteByte(b)
}
func (w *countWriter) WriteRune(r rune) (int, error) {
	w.count(r)
	return w.W.WriteRune(r)
}
func (w *countWriter) Write(b []byte) (int, error) {
	return w.WriteString(string(b))
}
func (w *countWriter) WriteString(s string) (int, error) {
	for _, r := range s {
		w.count(r)
	}
	return w.W.WriteString(s)
}

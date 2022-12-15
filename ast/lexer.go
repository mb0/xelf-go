package ast

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
)

const eof = -1

// Pos holds the line number and byte offset.
type Pos struct {
	Line int32 // line number 1-indexed
	Byte int32 // byte offset within the line 0-indexed
}

// Src spans from input position to an end position.
type Src struct {
	*Doc
	Pos
	End Pos
}

func (s Src) String() string {
	if s.Doc == nil || s.Name == "" {
		return fmt.Sprintf(":%d:%d", s.Line, s.Byte)
	}
	return fmt.Sprintf("%s:%d:%d", s.Name, s.Line, s.Byte)
}

// Tok is a lexer token with either a single rune or the raw string.
type Tok struct {
	Kind knd.Kind
	Rune rune
	Raw  string
	Src  Src
}

func (t Tok) String() string {
	if t.Rune != 0 {
		return string(t.Rune)
	}
	return t.Raw
}

type Doc struct {
	Name  string
	Lines []int32
}

// Lexer is simple token lexer.
type Lexer struct {
	r        io.RuneScanner
	cur, nxt rune
	idx      int32
	cun, nxn int
	doc      Doc
	err      error
}

// New returns a new Lexer for Reader r.
func NewLexer(r io.Reader, name string) *Lexer {
	l := &Lexer{idx: -1, doc: Doc{Name: name}}
	if rr, ok := r.(io.RuneScanner); ok {
		l.r = rr
	} else {
		l.r = bufio.NewReader(r)
	}
	l.next()
	return l
}

// next proceeds to and returns the next rune, updating the look-ahead.
func (l *Lexer) next() rune {
	if l.err != nil {
		return eof
	}
	l.cur, l.cun = l.nxt, l.nxn
	l.idx += int32(l.nxn)
	l.nxt, l.nxn, l.err = l.r.ReadRune()
	if l.cur == '\n' {
		l.doc.Lines = append(l.doc.Lines, int32(l.idx))
	}
	return l.cur
}

// pos returns a new pos at the current offset.
func (l *Lexer) pos() Pos {
	n, c := len(l.doc.Lines), l.idx
	if n > 0 {
		c -= l.doc.Lines[n-1]
	}
	return Pos{int32(n + 1), int32(c)}
}

// rtok returns a new token at the current offset.
func (l *Lexer) rtok(k knd.Kind) Tok {
	p := l.pos()
	return Tok{Kind: k, Rune: l.cur, Src: Src{Doc: &l.doc,
		Pos: p, End: Pos{p.Line, p.Byte + int32(l.cun)},
	}}
}

// tok returns a new token from start to the current offset.
func (l *Lexer) tok(k knd.Kind, start Pos, raw string) Tok {
	p := l.pos()
	return Tok{Kind: k, Raw: raw, Src: Src{Doc: &l.doc,
		Pos: start, End: Pos{p.Line, p.Byte + int32(l.cun)},
	}}
}

// Token reads and returns the next token or an error.
func (l *Lexer) Tok() (Tok, error) {
	r := l.next()
	for cor.Space(r) {
		r = l.next()
	}
	if r == eof {
		return l.rtok(knd.Void), l.err
	}
	if k := cor.Ctrl(r); k >= 0 {
		if k == int(knd.Str) {
			return l.lexString()
		}
		return l.rtok(knd.Kind(k)), nil
	}
	if cor.Digit(r) || r == '-' && cor.Digit(l.nxt) {
		return l.lexNumber()
	}
	if r >= '!' && r <= '~' && r != '\\' {
		return l.lexSymbol()
	}
	t := l.rtok(knd.Void)
	return t, ErrTokStart(t)
}

// lexString reads and returns a string token starting at the current offset.
func (l *Lexer) lexString() (Tok, error) {
	p := l.pos()
	q := l.cur
	var b strings.Builder
	b.WriteRune(q)
	c := l.next()
	var esc bool
	for c != eof && c != q || esc {
		if c == '\n' && q != '`' {
			t := l.tok(knd.Char, p, b.String())
			return t, ErrStrTerm(t)
		}
		esc = !esc && c == '\\' && q != '`'
		b.WriteRune(c)
		c = l.next()
	}
	if c == eof {
		t := l.tok(knd.Char, p, b.String())
		return t, ErrStrTerm(t)
	}
	b.WriteRune(q)
	return l.tok(knd.Char, p, b.String()), nil
}

// lexSymbol reads and returns a symbol token starting at the current offset.
func (l *Lexer) lexSymbol() (Tok, error) {
	p := l.pos()
	var b strings.Builder
	b.WriteRune(l.cur)
	for cor.SymPart(l.nxt) {
		b.WriteRune(l.next())
	}
	return l.tok(knd.Sym, p, b.String()), nil
}

// lexNumber reads and returns a number token starting at the current offset.
func (l *Lexer) lexNumber() (Tok, error) {
	p := l.pos()
	k := knd.Num
	var b strings.Builder
	if l.cur == '-' {
		b.WriteRune(l.cur)
		l.next()
	}
	b.WriteRune(l.cur)
	if l.cur != '0' {
		l.lexDigits(&b)
	} else if cor.Digit(l.nxt) {
		t := l.rtok(k)
		return t, ErrAdjZero(t)
	}
	if l.nxt == '.' {
		k = knd.Real
		b.WriteRune(l.nxt)
		l.next()
		if ok := l.lexDigits(&b); !ok {
			l.next()
			t := l.tok(k, p, b.String())
			return t, ErrNumFrac(t)
		}
	}
	if l.nxt == 'e' || l.nxt == 'E' {
		k = knd.Real
		b.WriteRune('e')
		l.next()
		if l.nxt == '+' || l.nxt == '-' {
			b.WriteRune(l.nxt)
			l.next()
		}
		if ok := l.lexDigits(&b); !ok {
			l.next()
			t := l.tok(k, p, b.String())
			return t, ErrNumExpo(t)
		}
	}
	return l.tok(k, p, b.String()), nil
}

// lexDigits reads the next digits and writes the to b.
// It returns false if no digit was read.
func (l *Lexer) lexDigits(b *strings.Builder) bool {
	if !cor.Digit(l.nxt) {
		return false
	}
	for ok := true; ok; ok = cor.Digit(l.nxt) {
		b.WriteRune(l.nxt)
		l.next()
	}
	return true
}

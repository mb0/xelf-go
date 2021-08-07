// Package typ provides a xelf type implementation, parser and inference system.
package typ

import (
	"fmt"
	"strings"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/knd"
)

// Type describes the shape of a xelf expression, literal or value.
type Type struct {
	Kind knd.Kind
	ID   int32
	Body
}

type BodyPair struct{ A, B Body }
type Hist []BodyPair

// Body contains additional type information.
type Body interface {
	EqualHist(Body, Hist) bool
}

// Equal returns whether type t and o are identical.
func (t Type) Equal(o Type) bool { return t.EqualHist(o, nil) }
func (t Type) EqualHist(o Type, h Hist) bool {
	if t.Kind != o.Kind || t.ID != o.ID {
		return false
	}
	if t.Body == nil {
		return o.Body == nil
	}
	return t.Body.EqualHist(o.Body, h)
}

func (t *Type) UnmarshalJSON(b []byte) error {
	s, err := cor.Unquote(string(b))
	if err != nil {
		return err
	}
	r, err := Parse(s)
	if err != nil {
		return err
	}
	*t = r
	return nil
}

func (t Type) String() string               { return bfr.String(t) }
func (t Type) MarshalJSON() ([]byte, error) { return bfr.JSON(t) }
func (t Type) Print(b *bfr.P) error {
	if b.JSON {
		b.Byte('"')
		b.JSON = false
		t.print(b, nil, true)
		b.JSON = true
		return b.Byte('"')
	}
	return t.print(b, nil, true)
}

func (t Type) print(b *bfr.P, sb *strings.Builder, enclose bool) error {
	if t.Kind == knd.Void {
		return b.Fmt("<>")
	}
	var isRef, isVar, isNone, isSel, isAlt bool
	k := t.Kind
	if isRef = k&knd.Ref != 0; isRef {
		k &^= knd.Ref
	}
	if isVar = k&knd.Var != 0; isVar {
		k &^= knd.Var
	}
	if isNone = k&knd.None != 0 && t.Kind != knd.None && k != knd.Any; isNone {
		k &^= knd.None
	}
	if isSel = k&knd.Sel != 0; isSel {
		k &^= knd.Sel
	}
	if sb == nil {
		sb = &strings.Builder{}
	}
	if isSel {
		sel := t.Body.(*SelBody)
		sb.WriteString(sel.Path)
	} else if k != knd.Void {
		s := knd.Name(k)
		if isAlt = k&knd.Alt != 0 || s == ""; isAlt {
			sb.WriteString("alt")
		} else if !(isNone && s == "any") {
			sb.WriteString(s)
		}
	}
	if isRef {
		sb.WriteByte('@')
		if rb, ok := t.Body.(*RefBody); ok {
			sb.WriteString(rb.Ref)
		} else {
			b.Err = fmt.Errorf("expect type ref body")
			return b.Err
		}
	} else if isVar {
		sb.WriteByte('@')
		if t.ID > 0 {
			fmt.Fprint(sb, t.ID)
		}
	}
	if isNone {
		sb.WriteByte('?')
	}
	if isAlt {
		b.Byte('<')
		b.Fmt(sb.String())
		for _, e := range altTypes(t) {
			b.Byte(' ')
			e.print(b, nil, false)
		}
		return b.Byte('>')
	}
	switch tb := t.Body.(type) {
	case *RefBody:
		if t.Kind&knd.Schm != 0 {
			b.Byte('<')
			b.Fmt(sb.String())
			b.Byte(' ')
			b.Fmt(tb.Ref)
			return b.Byte('>')
		}
	case *ElBody:
		if tb.El != Void {
			sb.WriteByte('|')
			return tb.El.print(b, sb, enclose)
		}
	case *ConstBody:
		b.Byte('<')
		b.Fmt(sb.String())
		b.Byte(' ')
		b.Fmt(tb.Name)
		return b.Byte('>')
	case *ParamBody:
		b.Byte('<')
		b.Fmt(sb.String())
		if tb.Name != "" && t.Kind&knd.Rec == 0 {
			b.Byte(' ')
			b.Fmt(tb.Name)
		}
		if t.Kind&knd.Obj == 0 {
			for _, p := range tb.Params {
				b.Byte(' ')
				if p.Name != "" {
					b.Fmt(p.Name)
					if p.Kind == knd.Void {
						b.Byte(';')
						continue
					}
					b.Byte(':')
				}
				p.Type.print(b, nil, false)
			}
		}
		return b.Byte('>')
	}
	if !enclose {
		return b.Fmt(sb.String())
	}
	b.Byte('<')
	b.Fmt(sb.String())
	return b.Byte('>')
}

package bfr

import (
	"fmt"
	"strings"

	"xelf.org/xelf/cor"
)

// Printer is an interface for types that can print
type Printer interface {
	Print(*P) error
}

// String writes p in xelf format and returns the result as string ignoring any error.
func String(p Printer) string {
	var b strings.Builder
	_ = p.Print(&P{Writer: &b, Plain: true})
	return b.String()
}

// JSON writes w in json format and returns the result as bytes or an error.
func JSON(p Printer) (res []byte, err error) {
	b := Get()
	defer Put(b)
	err = p.Print(&P{Writer: b, JSON: true})
	if err != nil {
		return nil, err
	}
	return append(res, b.Bytes()...), nil
}

// P is serialization context with output configuration flags for printers
type P struct {
	Writer
	JSON  bool
	Plain bool
	Depth int
	Tab   string
	Err   error
}

func (p *P) err(err error) error {
	if err != nil {
		p.Err = err
	}
	return p.Err
}
func (p *P) Byte(b byte) (err error) {
	return p.err(p.WriteByte(b))
}

// Fmt writes the formatted string to the buffer or returns an error
func (p *P) Fmt(f string, args ...interface{}) (err error) {
	if len(args) > 0 {
		_, err = fmt.Fprintf(p.Writer, f, args...)
	} else {
		_, err = p.WriteString(f)
	}
	return p.err(err)
}

func (p *P) Indent() bool {
	p.Depth++
	return p.Break()
}

func (p *P) Dedent() bool {
	p.Depth--
	return p.Break()
}

func (p *P) Break() bool {
	if p.Tab == "" {
		return false
	}
	err := p.WriteByte('\n')
	for i := p.Depth; err == nil && i > 0; i-- {
		_, err = p.WriteString(p.Tab)
	}
	p.err(err)
	return true
}

// Quote writes v as quoted string to the buffer or returns an error.
// The quote used depends on the json context flag.
func (p *P) Quote(v string) (err error) {
	if p.JSON {
		err = cor.WriteQuote(p, v, '"')
	} else {
		err = cor.WriteQuote(p, v, '\'')
	}
	return p.err(err)
}

// RecordKey writes key as quoted string followed by a colon to the buffer or returns an error.
// The quote used depends on the json context flag.
func (p *P) RecordKey(key string) (err error) {
	if p.JSON || !cor.IsSym(key) {
		err = p.Quote(key)
	} else {
		_, err = p.WriteString(key)
	}
	if err != nil {
		return p.err(err)
	}
	return p.Byte(':')
}

// Sep writes a record field or list element separator to the buffer or returns an error.
// The separator used depends on the json context flag.
func (p *P) Sep() error {
	if p.JSON {
		return p.Byte(',')
	}
	return p.Byte(' ')
}

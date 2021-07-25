// Package bfr provides a writer and printer interface and pool of bytes.Buffer.
package bfr

import (
	"bytes"
	"sync"
)

var pool = sync.Pool{New: func() interface{} {
	return &bytes.Buffer{}
}}

// Get returns a *bytes.Buffer from the pool.
func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

// Put a *bytes.Buffer back into the pool.
func Put(b *bytes.Buffer) {
	b.Reset()
	pool.Put(b)
}

// Writer is the common interface of bytes.Buffer, strings.Builder and bufio.Writer.
type Writer interface {
	Write([]byte) (int, error)
	WriteByte(byte) error
	WriteRune(rune) (int, error)
	WriteString(string) (int, error)
}

// Grow grows the buffer by n if it implements a Grow(int) method.
// Both bytes.Buffer and strings.Builder implement that method.
func Grow(b Writer, n int) {
	if v, ok := b.(interface{ Grow(int) }); ok {
		v.Grow(n)
	}
}

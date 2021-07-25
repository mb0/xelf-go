package lit

import (
	"fmt"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// BreakIter is a special error value that can be returned from iterators.
// It indicates that the iteration should be stopped even though no actual failure occurred.
var BreakIter = fmt.Errorf("break iter")

var (
	ErrIdxBounds   = fmt.Errorf("index out of bounds")
	ErrIdxNotFound = fmt.Errorf("idx not found")
	ErrKeyNotFound = fmt.Errorf("key not found")
	ErrAssign      = typ.ErrAssign
)

type Val = typ.LitVal
type Mut = typ.LitMut

type Lit struct {
	Res typ.Type
	Val
	Src ast.Src
}

func (a *Lit) Kind() knd.Kind {
	if a.Res.Kind&knd.Typ != 0 {
		return knd.Typ
	}
	return knd.Lit
}
func (a *Lit) Resl() typ.Type  { return a.Res }
func (a *Lit) Type() typ.Type  { return a.Res }
func (a *Lit) Source() ast.Src { return a.Src }

type Null struct{}

func (Null) Type() typ.Type               { return typ.None }
func (Null) Nil() bool                    { return true }
func (Null) Zero() bool                   { return true }
func (Null) Value() Val                   { return Null{} }
func (Null) String() string               { return "null" }
func (Null) Print(p *bfr.P) error         { return p.Fmt("null") }
func (Null) MarshalJSON() ([]byte, error) { return []byte("null"), nil }

// Idxr is the interface for indexer values.
type Idxr interface {
	Mut
	Len() int
	// Idx returns the literal of the element at idx or an error.
	Idx(idx int) (Val, error)
	// SetIdx sets the element value at idx to v and returns the indexer or an error.
	SetIdx(idx int, v Val) error
	// IterIdx iterates over elements, calling iter with the elements index and value.
	// If iter returns an error the iteration is aborted.
	IterIdx(iter func(int, Val) error) error
}

// Apdr is the interface for indexer values supporting append.
type Apdr interface {
	Idxr
	Append(v Val) error
}

// Keyr is the interface for keyer values.
type Keyr interface {
	Mut
	Len() int
	// Keys returns a string slice of all keys.
	Keys() []string
	// Key returns the value of the element with key or an error.
	Key(key string) (Val, error)
	// SetKey sets the elements value with key to v and returns the keyer or an error.
	SetKey(key string, v Val) error
	// IterKey iterates over elements, calling iter with the elements key and value.
	// If iter returns an error the iteration is aborted.
	IterKey(iter func(string, Val) error) error
}

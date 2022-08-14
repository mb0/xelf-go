package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

// BreakIter is a special error value that can be returned from iterators.
// It indicates that the iteration should be stopped even though no actual failure occurred.
var BreakIter = fmt.Errorf("break iter")

var (
	ErrIdxNotFound = fmt.Errorf("idx not found")
	ErrKeyNotFound = fmt.Errorf("key not found")
	ErrAssign      = typ.ErrAssign
	ErrIdxBounds   = typ.ErrIdxBounds
)

// Val is the common interface of all literal values.
type Val = typ.LitVal

// Mut is the common interface of all mutable literal values.
// Mutable values should have an UnmarshalJSON method unless the base type is natively supported.
type Mut = typ.LitMut

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

// Lenr is the common interface of value that have a length.
type Lenr interface {
	Val
	Len() int
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

// Prx is the interface for all reflection based mutable values.
type Prx interface {
	Mut
	// Reflect returns the reflect value pointed to by this proxy.
	Reflect() reflect.Value
	// NewWith returns a new proxy instance with ptr as value.
	// This method is used internally for proxy caching and should only be called with pointer
	// values known to be compatible with this proxy implementation.
	NewWith(ptr reflect.Value) (Mut, error)
}

func PrintZero(p *bfr.P, t typ.Type) error {
	k := t.Kind & knd.Any
	if k&knd.None != 0 || k.Count() != 1 {
		return p.Fmt("null")
	}
	switch k {
	case knd.Typ:
		return typ.Void.Print(p)
	case knd.Bool:
		return p.Fmt("false")
	case knd.Int, knd.Real:
		return p.Fmt("0")
	case knd.Str, knd.Raw:
		return Str("").Print(p)
	case knd.UUID:
		return UUID{}.Print(p)
	case knd.Time:
		return Time{}.Print(p)
	case knd.Span:
		return Span(0).Print(p)
	case knd.List:
		return p.Fmt(`[]`)
	case knd.Dict, knd.Obj:
		return p.Fmt(`{}`)
	}
	return p.Fmt("null")
}

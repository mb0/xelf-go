package cor

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// ErrUUID indicates an invalid input format when parsing an uuid.
var ErrUUID = fmt.Errorf("invalid uuid format")

// NewUUID returns 16 random bytes from crypto/rand.
func NewUUID() (id [16]byte) {
	if _, err := rand.Read(id[:]); err != nil {
		panic(err)
	}
	id[6] = (id[6] & 0x0f) | (4 << 4)
	id[8] = (id[8] & 0xbf) | 0x80
	return id
}

// UUID parses s and returns a pointer to the uuid bytes or nil on error.
func UUID(s string) *[16]byte {
	v, err := ParseUUID(s)
	if err != nil {
		return nil
	}
	return &v
}

// FormatUUID returns v as string in the canonical uuid format.
func FormatUUID(v [16]byte) string {
	var b strings.Builder
	w := hex.NewEncoder(&b)
	var nn int
	for i, n := range [5]int{4, 2, 2, 2, 6} {
		if i > 0 {
			b.WriteByte('-')
		}
		w.Write(v[nn : nn+n])
		nn += n
	}
	return b.String()
}

func MustParseUUID(s string) [16]byte {
	id, err := ParseUUID(s)
	if err != nil {
		panic(err)
	}
	return id
}

// ParseUUID parses s and return the uuid bytes or an error.
// It accepts 16 hex encoded bytes with up to four dashes in between.
func ParseUUID(s string) ([16]byte, error) {
	var res [16]byte
	if len(s) < 32 || len(s) > 36 {
		return res, ErrUUID
	}
	if len(s) > 36 {
		return res, ErrUUID
	}
	for i, o := 0, 0; i+1 < len(s) && o < 16; {
		a := s[i]
		if a == '-' {
			i++
			continue
		}
		b := s[i+1]
		a, aok := fromHex(a)
		b, bok := fromHex(b)
		if !aok || !bok {
			return res, ErrUUID
		}
		res[o] = a<<4 | b
		i += 2
		o += 1
	}
	return res, nil
}

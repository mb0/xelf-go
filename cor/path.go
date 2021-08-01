package cor

import (
	"fmt"
	"strconv"
	"strings"
)

// Seg is one segment of a path. It consists of a dot or slash, followed by a key or index.
type Seg struct {
	Key string
	Idx int
	Sel bool
}

func (s Seg) String() string {
	if s.Key != "" {
		return s.Key
	}
	return strconv.Itoa(s.Idx)
}

// Path consists of non-empty segments separated by dots '.' or slashes '/'.
// Segments starting with a digit or minus sign are idx segments that try to select into an idxr
// literal, otherwise the segment represents a key used to select into a keyr literal. Segments
// starting with a slash signify a selection from a idxr literal.
type Path []Seg

func (p Path) String() string {
	var b strings.Builder
	for _, s := range p {
		if s.Sel {
			b.WriteByte('/')
		} else {
			b.WriteByte('.')
		}
		b.WriteString(s.String())
	}
	return b.String()
}

// ParsePath reads and returns the dot separated segments for the path or an error.
func ParsePath(path string) (res Path, err error) {
	if path == "" {
		return nil, nil
	}
	var sel bool
	switch r := path[0]; r {
	case '.', '/':
		path = path[1:]
		sel = r == '/'
	}
	res = make(Path, 0, len(path)>>2)
	var last int
	for i, r := range path {
		switch r {
		case '.', '/':
			res, err = res.Add(path[last:i], sel)
			if err != nil {
				return nil, err
			}
			sel = r == '/'
			last = i + 1
		}
	}
	if len(path) > last {
		res, err = res.Add(path[last:], sel)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// Add appends a new segment and returns the new path or an error for empty segments.
func (p Path) Add(s string, sel bool) (Path, error) {
	if s == "" {
		return nil, fmt.Errorf("empty segment")
	}
	if c := s[0]; c == '-' || c >= '0' && c <= '9' {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		p = append(p, Seg{Idx: i, Sel: sel})
	} else {
		p = append(p, Seg{Key: s, Sel: sel})
	}
	return p, nil
}

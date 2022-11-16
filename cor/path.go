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
	for len(path) > 0 {
		res, path, err = addSeg(res, path)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

func addSeg(p Path, s string) (_ Path, rest string, err error) {
	var res Seg
	var idx, upper, other bool
	if r := s[0]; r == '.' || r == '/' {
		s = s[1:]
		res.Sel = r == '/'
	} else if len(p) > 0 {
		return p, s, fmt.Errorf("missing path sep")
	}
	for i, r := range s {
		if r == '.' || r == '/' {
			s, rest = s[:i], s[i:]
			break
		} else if r >= 'A' && r <= 'Z' {
			upper = true
		} else if r >= '0' && r <= '9' || r == '-' && i == 0 {
			idx = true
		} else {
			other = true
		}
	}
	if upper || other {
		if upper {
			s = strings.ToLower(s)
		}
		res.Key = s
	} else if idx && len(s) < 10 {
		res.Idx, _ = strconv.Atoi(s)
	} else {
		res.Key = s
	}
	if s != "" || res.Sel {
		p = append(p, res)
	}
	return p, rest, nil
}

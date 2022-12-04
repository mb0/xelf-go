package cor

import (
	"fmt"
	"strconv"
	"strings"
)

const magic = '?' // '.'&'?'=='.', '/'&'?'=='/', 'n'&'?'=='.', 'o'&'?'=='/', '.'|'@'=='n', :) !

// Seg is one segment of a path. It consists of a dot or slash, followed by a key or index.
type Seg struct {
	Key string
	Idx int
	Sel byte
}

func (s Seg) Sep() byte   { return (s.Sel & '?') /* magic */ }
func (s Seg) Empty() bool { return (s.Sel & '@') != 0 /* magic */ }
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
		if sep := s.Sep(); sep != 0 {
			b.WriteByte(sep)
		}
		b.WriteString(s.String())
	}
	return b.String()
}

// ParsePath reads and returns the segments for path or an error.
func ParsePath(path string) (res Path, err error) {
	for len(path) > 0 {
		res, path, err = addSeg(res, path)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

// FillPath reads and returns the segments for the path filled with vars or an error.
func FillPath(path string, vars ...string) (res Path, err error) {
	for len(path) > 0 {
		res, path, err = addSeg(res, path)
		if err != nil {
			return nil, err
		}
	}
	return res, res.FillVars(vars)
}

func addSeg(p Path, s string) (_ Path, rest string, err error) {
	var res Seg
	if r := s[0]; r == '.' || r == '/' {
		s = s[1:]
		res.Sel = r
	} else if len(p) > 0 {
		return p, s, fmt.Errorf("missing path sep")
	}
	rest = res.parse(s, false)
	p = append(p, res)
	return p, rest, nil
}

func (p Path) HasVars() bool {
	for _, s := range p {
		if s.Key == "$" {
			return true
		}
	}
	return false
}

func (p Path) FillVars(vars []string) error {
	for i := range p {
		s := &p[i]
		if s.Key != "$" {
			continue
		}
		if len(vars) == 0 {
			return fmt.Errorf("not enough path variables")
		}
		s.parse(vars[0], true)
		vars = vars[1:]
	}
	if len(vars) > 0 {
		return fmt.Errorf("superflous path segment variables %s", vars)
	}
	return nil
}

func (res *Seg) parse(s string, ignoreSep bool) (rest string) {
	var idx, upper, other bool
	for i, r := range s {
		if !ignoreSep && (r == '.' || r == '/') {
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
		res.Key = ""
		res.Idx, _ = strconv.Atoi(s)
	} else { // empty
		res.Key = ""
		res.Sel += '@' // magic
	}
	return rest
}

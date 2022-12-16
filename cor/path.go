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

func (s Seg) Sep() byte   { return (s.Sel & magic) }
func (s Seg) Empty() bool { return (s.Sel & '@') != 0 /* magic */ }
func (s Seg) IsIdx() bool { return s.Idx != 0 || !s.Empty() && s.Key == "" }
func (s Seg) String() string {
	if s.IsIdx() {
		return strconv.Itoa(s.Idx) + s.Key
	}
	return s.Key
}

// Path consists of non-empty segments separated by dots '.' or slashes '/'.
// Segments starting with a digit or minus sign are idx segments that try to select into an idxr
// literal, otherwise the segment represents a key used to select into a keyr literal. Segments
// starting with a slash signify a selection from a idxr literal.
type Path []Seg

func (p Path) String() string { return p.Suffix("") }
func (p Path) Suffix(suf string) string {
	var b strings.Builder
	for _, s := range p {
		if sep := s.Sep(); sep != 0 {
			b.WriteByte(sep)
		}
		b.WriteString(s.String())
	}
	b.WriteString(suf)
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
	res, err = ParsePath(path)
	if err != nil {
		return nil, err
	}
	return res, res.FillVars(vars)
}
func (p Path) Plain() string {
	if len(p) == 1 && p[0].Sep() == 0 {
		return p[0].Key
	}
	return ""
}
func (p Path) Fst() Seg {
	if len(p) > 0 {
		return p[0]
	}
	return Seg{Sel: 'n'}
}
func (p Path) CountVars() (n int) {
	for i, s := range p {
		if s.Key == "$" && (i > 0 || s.Sel != 0) {
			n++
		}
	}
	return n
}

func (p Path) FillVars(vars []string) error {
	for i := range p {
		s := &p[i]
		if s.Key != "$" || (i == 0 && s.Sel == 0) {
			continue
		}
		if len(vars) == 0 {
			return fmt.Errorf("not enough path variables")
		}

		idx, raw, _ := checkSeg(vars[0], false)
		if s.Key = ""; idx {
			s.Idx, _ = strconv.Atoi(raw)
		} else if raw == "" {
			s.Sel += '@' // magic
		} else {
			s.Key = raw
		}
		vars = vars[1:]
	}
	if len(vars) > 0 {
		return fmt.Errorf("superflous path segment variables %s", vars)
	}
	return nil
}

func addSeg(p Path, s string) (Path, string, error) {
	var res Seg
	if r := s[0]; r == '.' || r == '/' {
		s = s[1:]
		res.Sel = r
	} else if len(p) > 0 {
		return p, "", fmt.Errorf("missing path sep")
	}
	idx, raw, rest := checkSeg(s, true)
	if idx {
		res.Idx, _ = strconv.Atoi(s)
	} else if raw != "" {
		res.Key = strings.ToLower(raw)
	} else { // empty
		res.Key = ""
		res.Sel += '@' // magic
	}
	p = append(p, res)
	return p, rest, nil
}

func checkSeg(s string, split bool) (idx bool, raw, rest string) {
	var other bool
	raw = s
	for i, r := range s {
		if split && (r == '.' || r == '/') {
			raw, rest = s[:i], s[i:]
			break
		} else if r >= '0' && r <= '9' || r == '-' && i == 0 {
			idx = true
		} else {
			other = true
		}
	}
	idx = idx && !other && s != "-"
	return
}

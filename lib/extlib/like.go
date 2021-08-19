package extlib

import (
	"strings"
	"unicode/utf8"
)

func Like(txt, term string, ignCase bool) bool {
	pat, min := parsePat(term, ignCase)
	if len(pat) == 0 || len(txt) < min {
		return false
	}
	return matchPat(txt, pat, ignCase)
}
func matchPat(s string, pat []seg, ign bool) bool {
	rest, pat, ok := shaveLeft(s, pat, ign)
	if !ok {
		return false
	}
	rest, pat, ok = shaveRight(rest, pat, ign)
	if !ok {
		return false
	}
	var last seg
	for i, seg := range pat {
		if seg.Text != "" {
			rest = strings.ToLower(rest)
			for len(rest) >= len(seg.Text) { // tries
				idx := strings.Index(rest, seg.Text)
				if idx < 0 {
					return false
				}
				tmp := rest[idx+len(seg.Text):]
				if i == len(pat)-2 {
					return true
				}
				if !matchPat(tmp, pat[i+1:], false) {
					rest = rest[idx+1:]
					continue
				}
				return true
			}
			return false
		} else {
			rest, ok = skipRunes(rest, seg.Skip)
			if !ok {
				return false
			}
		}
		last = seg
	}
	return len(rest) == 0 || last.Search
}

func shaveLeft(s string, pat []seg, ign bool) (_ string, res []seg, ok bool) {
	res = pat
	for i, seg := range pat {
		s, ok = skipRunes(s, seg.Skip)
		if seg.Skip > 0 {
			pat[i].Skip = 0
		}
		if !ok || seg.Search {
			break
		}
		n := len(seg.Text)
		if ok = n == 0 || len(s) >= n && equal(s[:n], seg.Text, ign); !ok {
			break
		}
		res = pat[i+1:]
		s = s[n:]
	}
	return s, res, ok
}
func shaveRight(s string, pat []seg, ign bool) (_ string, res []seg, ok bool) {
	res = pat
	ok = true
	for i := len(pat) - 1; i >= 0; i-- {
		seg := &pat[i]
		s, ok = skipRunesRight(s, seg.Skip)
		seg.Skip = 0
		if !ok || seg.Search {
			break
		}
		n := len(seg.Text)
		if ok = n == 0 || len(s) >= n && equal(s[len(s)-n:], seg.Text, ign); !ok {
			break
		}
		res = pat[:i]
		s = s[:len(s)-n]
	}
	return s, res, ok
}

func equal(txt, seg string, ign bool) bool {
	return txt == seg || ign && strings.EqualFold(txt, seg)
}

func skipRunes(s string, n int) (string, bool) {
	for i := 0; i < n; i++ {
		if s == "" {
			return s, false
		}
		_, w := utf8.DecodeRuneInString(s)
		s = s[w:]
	}
	return s, true
}
func skipRunesRight(s string, n int) (string, bool) {
	for i := 0; i < n; i++ {
		if s == "" {
			return s, false
		}
		_, w := utf8.DecodeLastRuneInString(s)
		s = s[:len(s)-w]
	}
	return s, true
}

type seg struct {
	Text   string
	Skip   int
	Search bool
}

func parsePat(n string, ign bool) (res []seg, min int) {
	var last int
	var buf *strings.Builder
	for i := 0; i < len(n); i++ {
		c := n[i]
		switch c {
		case '_', '%':
			if last < i {
				var txt string
				if buf != nil {
					txt = buf.String()
				} else {
					txt = n[last:i]
				}
				if ign {
					txt = strings.ToLower(txt)
				}
				min += len(txt)
				res = append(res, seg{Text: txt}, seg{})
			} else if i == 0 {
				res = append(res, seg{})
			}
			seg := &res[len(res)-1]
			if c == '_' {
				min++
				seg.Skip++
			} else {
				seg.Search = true
			}
			last = i + 1
			buf = nil
		case '\\':
			if buf == nil {
				buf = &strings.Builder{}
				buf.WriteString(n[last:i])
			}
			i++
			if i < len(n) {
				buf.WriteByte(n[i])
			}
			continue
		}
		if buf != nil {
			buf.WriteByte(c)
		}
	}
	if last < len(n) {
		var txt string
		if buf != nil {
			txt = buf.String()
		} else {
			txt = n[last:]
		}
		if ign {
			txt = strings.ToLower(txt)
		}
		min += len(txt)
		res = append(res, seg{Text: txt})
	}
	return res, min
}

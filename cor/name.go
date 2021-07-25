package cor

import "strings"

// NameStart tests whether r is ascii letter or underscore.
func NameStart(r rune) bool { return Letter(r) || r == '_' }

// NamePart tests whether r is ascii letter, digit or underscore.
func NamePart(r rune) bool { return NameStart(r) || Digit(r) }

// IsName tests whether s is a valid name.
func IsName(s string) bool {
	if s == "" || !NameStart(rune(s[0])) {
		return false
	}
	for _, r := range s[1:] {
		if !NamePart(r) {
			return false
		}
	}
	return true
}

// Cased returns n starting with uppercase letter.
// This function is especially used for go code gen.
func Cased(s string) string {
	for i, c := range s {
		if NameStart(c) {
			s = s[i:]
			break
		}
	}
	for i, c := range s {
		if !NamePart(c) && c != '.' {
			s = s[:i]
			break
		}
	}
	if s != "" && (s[0] < 'A' || s[0] > 'Z') {
		return strings.ToUpper(s[:1]) + s[1:]
	}
	return s
}

func IsCased(s string) bool {
	if s != "" {
		b := s[0]
		return b >= 'A' && b <= 'Z'
	}
	return false
}

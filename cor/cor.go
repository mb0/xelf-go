/*
Package cor has some utility function for working with numeric and character literal values.

The literal values of raw, uuid, time and span all have a formatter, string parser
and a pointer parser function that returns nil on error.

The primitive go types bool, int64, float64 and string have a pointer helper that returns a
pointer to argument value.
*/
package cor

// Bool returns a pointer to v.
func Bool(v bool) *bool { return &v }

// Int returns a pointer to v.
func Int(v int64) *int64 { return &v }

// Real returns a pointer to v.
func Real(v float64) *float64 { return &v }

// Str returns a pointer to v.
func Str(v string) *string { return &v }

// Any returns a pointer to v.
func Any(v interface{}) *interface{} { return &v }

// Space tests whether r is a space, tab or newline.
func Space(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }

// Digit tests whether r is an ascii digit.
func Digit(r rune) bool { return r >= '0' && r <= '9' }

// Letter tests whether r is an ascii letter.
func Letter(r rune) bool { return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') }

// Punct tests whether r is one of the ascii punctuations allowed in symbols.
func Punct(r rune) bool {
	switch r {
	case '!', '#', '$', '%', '&', '*', '+', '-', '.', '/',
		'=', '?', '@', '^', '|', '~':
		return true
	}
	return false
}

func Ctrl(r rune) int {
	switch r {
	case '"', '\'', '`':
		return 0x1000
	case '(', ')':
		return 0x80
	case ',':
		return 0
	case ':', ';':
		return 0x20
	case '<', '>':
		return 0x08
	case '[', ']':
		return 0x140000
	case '{', '}':
		return 0x180000
	}
	return -1
}

package ast

import (
	"fmt"

	"xelf.org/xelf/knd"
)

// Error in addition to a name has a source position, error code and optional help message.
type Error struct {
	Src  Src
	Code uint
	Name string
	Help string
}

func (e *Error) Error() string {
	if e.Help != "" {
		return fmt.Sprintf("%s: %s E%d\n\t%s", e.Src, e.Name, e.Code, e.Help)
	}
	return fmt.Sprintf("%s: %s E%d", e.Src, e.Name, e.Code)
}

func ErrTokStart(t Tok) *Error {
	return &Error{Src: t.Src, Code: 101, Name: "unexpected token start",
		Help: fmt.Sprintf("at input %q", t.String())}
}
func ErrAdjZero(t Tok) *Error {
	return &Error{Src: t.Src, Code: 102, Name: "adjacent zeros",
		Help: "number zero must be followed by a fraction or whitespace"}
}
func ErrNumFrac(t Tok) *Error { return &Error{Src: t.Src, Code: 103, Name: "expect number fraction"} }
func ErrNumExpo(t Tok) *Error { return &Error{Src: t.Src, Code: 104, Name: "expect number exponent"} }
func ErrStrTerm(t Tok) *Error {
	return &Error{Src: t.Src, Code: 105, Name: "unterminated string",
		Help: fmt.Sprintf("expecting closing %q", t.Raw[0])}
}
func ErrUnquote(t Tok, err error) *Error {
	return &Error{Src: t.Src, Code: 106, Name: fmt.Sprintf("invalid string quoting %v", err)}
}
func ErrTreeTerm(t Tok) *Error {
	_, end := parens(t.Kind)
	return &Error{Src: t.Src, Code: 111, Name: "unterminated tree",
		Help: fmt.Sprintf("expecting closing %q", end)}
}
func ErrInvalidSep(t Tok) *Error { return &Error{Src: t.Src, Code: 112, Name: "invalid separator"} }
func ErrInvalidTag(t Tok) *Error { return &Error{Src: t.Src, Code: 113, Name: "invalid tag"} }

func ErrUnexpected(a Ast) *Error {
	return &Error{Src: a.Src, Code: 201, Name: fmt.Sprintf("unexpected input %s", a)}
}
func ErrExpectSym(a Ast) *Error {
	return &Error{Src: a.Src, Code: 202, Name: fmt.Sprintf("expect sym got %s", a)}
}
func ErrExpectTag(a Ast) *Error {
	return &Error{Src: a.Src, Code: 203, Name: fmt.Sprintf("expect tag got %s", a)}
}
func ErrInvalidType(s Src, raw string) *Error {
	return &Error{Src: s, Code: 301, Name: fmt.Sprintf("invalid type %s", raw)}
}
func ErrInvalidParams(a Ast) *Error {
	return &Error{Src: a.Src, Code: 302, Name: fmt.Sprintf("invalid type parameters %s", a)}
}
func ErrExpect(a Ast, kind knd.Kind) *Error {
	return &Error{Src: a.Src, Code: 400, Name: fmt.Sprintf("expect %s got %s", knd.Name(kind), a)}
}
func ErrInvalidBool(a Ast) *Error {
	return &Error{Src: a.Src, Code: 401, Name: fmt.Sprintf("invalid bool %s", a)}
}
func ErrInvalid(a Ast, kind knd.Kind, err error) *Error {
	return &Error{Src: a.Src, Code: 402, Name: fmt.Sprintf("invalid %s %s: %v", knd.Name(kind), a, err)}
}

func parens(k knd.Kind) (rune, rune) {
	switch k {
	case knd.Typ:
		return '<', '>'
	case knd.Call:
		return '(', ')'
	case knd.List:
		return '[', ']'
	case knd.Dict:
		return '{', '}'
	}
	return 0, 0
}

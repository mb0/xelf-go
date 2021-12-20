package ast

import (
	"fmt"
	"io"
	"strings"

	"xelf.org/xelf/knd"
)

// Error in addition to a name has a source position, error code and optional help message.
type Error struct {
	Src  Src
	Code uint
	Name string
	Help string
	Err  error
}

func (e *Error) Unwrap() error { return e.Err }
func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s E%d", e.Src, e.Name, e.Code)
	if e.Err != nil {
		b.WriteString("\n\t")
		b.WriteString(e.Err.Error())
	}
	if e.Help != "" {
		b.WriteString("\n\t")
		b.WriteString(e.Help)
	}
	return b.String()
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
	return &Error{Src: t.Src, Code: 105, Name: "unterminated string", Err: io.EOF,
		Help: fmt.Sprintf("expecting closing %q", t.Raw[0])}
}
func ErrUnquote(t Tok, err error) *Error {
	return &Error{Src: t.Src, Code: 106, Name: "invalid string quoting", Err: err}
}
func ErrTreeTerm(t Tok) *Error {
	_, end := parens(t.Kind)
	return &Error{Src: t.Src, Code: 111, Name: "unterminated tree", Err: io.EOF,
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
	name := fmt.Sprintf("invalid %s %s", knd.Name(kind), a)
	return &Error{Src: a.Src, Code: 402, Name: name, Err: err}
}
func ErrUnexpectedExp(s Src, e interface{}) *Error {
	return &Error{Src: s, Code: 501, Name: fmt.Sprintf("unexpected exp %T", e)}
}
func ErrReslSym(s Src, sym string, err error) *Error {
	name := fmt.Sprintf("sym %s unresolved", sym)
	return &Error{Src: s, Code: 502, Name: name, Err: err}
}
func ErrReslSpec(s Src, name string, err error) *Error {
	return &Error{Src: s, Code: 503, Name: name, Err: err}
}
func ErrUnify(s Src, name string) *Error {
	return &Error{Src: s, Code: 504, Name: name}
}
func ErrLayout(s Src, t fmt.Stringer, err error) *Error {
	name := fmt.Sprintf("layout %s failed", t)
	return &Error{Src: s, Code: 505, Name: name, Err: err}
}
func ErrEval(s Src, name string, err error) *Error {
	name = fmt.Sprintf("eval %s failed", name)
	return &Error{Src: s, Code: 506, Name: name, Err: err}
}
func ErrUserErr(s Src, name string, err error) *Error {
	return &Error{Src: s, Code: 600, Name: name, Err: err}
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

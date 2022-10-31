package ext

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

func TestFunc(t *testing.T) {
	reg := &lit.Reg{}
	echo, err := NewFunc(reg, "echo", func(s string) string { return s })
	if err != nil {
		t.Fatal("failed to create node spec for test Element")
	}
	env := exp.Builtins(make(lib.Specs).AddMap(lib.Core).Add(echo))
	tests := []struct {
		raw string
		xlf string
	}{
		{`(echo "a")`, `'a'`},
	}
	for _, test := range tests {
		got, err := exp.NewProg(nil, reg, env).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		vs := bfr.String(got)
		if vs != test.xlf {
			t.Errorf("want %s got %s", test.xlf, vs)
		}
	}
}

func TestReflectFunc(t *testing.T) {
	tests := []struct {
		fun   interface{}
		name  string
		names []string
		typ   typ.Type
		err   bool
	}{
		{strings.ToLower, "lower", nil, typ.Func("lower",
			typ.P("", typ.Str),
			typ.P("", typ.Str),
		), false},
		{strings.Split, "", nil, typ.Func("",
			typ.P("", typ.Str),
			typ.P("", typ.Str),
			typ.P("", typ.ListOf(typ.Str)),
		), false},
		{time.Parse, "", nil, typ.Func("",
			typ.P("", typ.Str),
			typ.P("", typ.Str),
			typ.P("", typ.Time),
		), true},
		{time.Time.Format, "", []string{"t?", "format"}, typ.Func("",
			typ.P("t?", typ.Time),
			typ.P("format", typ.Str),
			typ.P("", typ.Str),
		), false},
		{fmt.Sprintf, "", nil, typ.Func("",
			typ.P("", typ.Str),
			typ.P("", typ.ListOf(typ.Any)),
			typ.P("", typ.Str),
		), false},
	}
	for _, test := range tests {
		f, err := NewFunc(nil, test.name, test.fun, test.names...)
		if err != nil {
			t.Errorf("reflect for %+v err: %v", test.fun, err)
			continue
		}
		if !test.typ.Equal(f.Decl) {
			t.Errorf("for %T want param %s got %s", test.fun, test.typ, f.Decl)
		}
		if test.err != f.err {
			t.Errorf("for %T want err %v got %v", test.fun, test.err, f.err)
		}
	}
}

func TestFuncEval(t *testing.T) {
	tests := []struct {
		fun   interface{}
		names []string
		raw   string
		xlf   string
	}{
		{strings.ToLower, nil, `(_ 'HELLO')`, `'hello'`},
		{time.Time.Format, []string{"t?", "format"}, `(_ format:'2006')`, `'0001'`},
		{fmt.Sprintf, nil, `(_ "Hi %s no. %d." "you" 9)`, `'Hi you no. 9.'`},
	}
	for _, test := range tests {
		reg := &lit.Reg{}
		f, err := NewFunc(reg, "_", test.fun, test.names...)
		if err != nil {
			t.Errorf("reflect for %+v err: %v", test.fun, err)
			continue
		}
		env := exp.Builtins(lib.Specs{"_": f}.AddMap(lib.Core))
		got, err := exp.NewProg(nil, reg, env).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		vs := bfr.String(got)
		if vs != test.xlf {
			t.Errorf("want %s got %s", test.xlf, vs)
		}
	}
}

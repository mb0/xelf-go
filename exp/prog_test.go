package exp_test

import (
	"strings"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib/extlib"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

type Point struct{ X, Y int }

func TestProgEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`(@test.point {})`, `{x:0 y:0}`},
		{`(dot {a:[{b:2}]} .a.0.b)`, `2`},
		{`(dot {a:[{b:2}, {b:3}]} .a/b)`, `[2 3]`},
		{`(dot {a:'2021-08-19T15:00:00Z'} (month .a))`, `8`},
		{`(dyn (month $now))`, `8`},
	}
	reg := &lit.Reg{}
	mut := reg.MustProxy(&Point{})
	reg.SetRef("test.point", mut.Type(), mut)
	env := &exp.ArgEnv{Par: extlib.Std, Typ: typ.Dict, Val: &lit.Dict{Reg: reg, Keyed: []lit.KeyVal{
		{Key: "now", Val: lit.Str("2021-08-19T15:00:00Z")},
	}}}
	for _, test := range tests {
		got, err := exp.Eval(nil, reg, env, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		str := got.String()
		if str != test.want {
			t.Errorf("eval %s want res %s got %s", test.raw, test.want, str)
		}
	}
}

func TestProgResl(t *testing.T) {
	tests := []struct {
		raw  string
		want string
		sig  string
	}{
		{`(if true 1 2)`, `<num@1>`,
			`<form if <tupl cond:any act:exp|num@1> else:exp?|num@1 num@1>`},
		{`(if true "one")`, `<char@1>`,
			`<form if <tupl cond:any act:exp|char@1> else:exp?|char@1 char@1>`},
		{`(make @test.point {})`, `<obj exp_test.Point>`, ``},
	}
	reg := &lit.Reg{}
	mut := reg.MustProxy(&Point{})
	reg.SetRef("test.point", mut.Type(), mut)
	for _, test := range tests {
		e, err := exp.Read(reg, strings.NewReader(test.raw), "test")
		if err != nil {
			t.Errorf("read %s failed: %v", test.raw, err)
			continue
		}
		p := exp.NewProg(nil, reg, extlib.Std, e)
		got, err := p.Resl(p.Root, p.Exp, typ.Void)
		if err != nil {
			t.Errorf("resl %s failed: %v", test.raw, err)
			continue
		}
		ts := got.Resl().String()
		if ts != test.want {
			t.Errorf("resl %s want res %s got %s", test.raw, test.want, ts)
		}
		if test.sig == "" {
			continue
		}
		c, ok := got.(*exp.Call)
		if !ok {
			t.Errorf("resl %s want call got %T", test.raw, got)
		}
		ss := c.Sig.String()
		if ss != test.sig {
			t.Errorf("resl %s want sig %s got %s", test.raw, test.sig, ss)
		}
	}
}

package exp_test

import (
	"testing"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lib/extlib"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

func TestProgEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`($now)`, `'2021-08-19T15:00:00Z'`},
		{`($)`, `{now:'2021-08-19T15:00:00Z'}`},
		{`(@test.point {})`, `{x:0 y:0}`},
		{`(with {a:[{b:2}]} .a.0.b)`, `2`},
		{`(with {a:[{b:2}, {b:3}]} .a/b)`, `[2 3]`},
		{`(with {a:'2021-08-19T15:00:00Z'} (month .a))`, `8`},
		{`((month $now))`, `8`},
		{`$test`, `null`},
		{`@`, `<@1>`},
		{`<@>`, `<@1>`},
		{`<@test.point>`, `<obj@test.point>`},
		{`(with make ([]+ (. int) (. str)))`, `[0 '']`},
		{`(fold (list|typ + int str) [] (fn r:list t:typ (mut .r + (.t null))))`, `[0 '']`},
	}
	tval, _ := typ.Parse("<obj@test.point x:int y:int>")
	env := &lib.DotEnv{Par: extlib.Std, Lets: lit.MakeObj(lit.Keyed{
		{Key: "test", Val: &lit.Keyed{{Key: "point", Val: tval}}},
	})}
	arg := &lit.Dict{Keyed: []lit.KeyVal{
		{Key: "now", Val: lit.Char("2021-08-19T15:00:00Z")},
	}}
	for _, test := range tests {
		got, err := exp.NewProg(env).RunStr(test.raw, arg)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		str := bfr.String(got)
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
		{`(if true 1 2)`, `<call|num@1>`,
			`<form@if <tupl cond:any then:exp|num@1> else:exp?|num@1 num@1>`},
		{`(if true "one")`, `<call|char@1>`,
			`<form@if <tupl cond:any then:exp|char@1> else:exp?|char@1 char@1>`},
		{`<@>`, `<lit|typ|@1>`, `<@1>`},
		{`bool`, `<lit|typ|bool>`,
			`<bool>`},
		{`add`, `<lit|spec>`,
			`<form@add num@ tupl?|num _>`},
		{`(if true add sub)`, `<call|spec>`,
			`<form@if <tupl cond:any then:exp|spec> else:exp?|spec spec>`},
		{`(@test.point {})`, `<call|obj@test.point>`, ``},
		{`(add (int 1) 2)`, `<call|int>`, `<form@add int tupl?|num int>`},
		{`(fn 1)`, `<lit|func@fn1 num>`, ``},
		{`(fn (add _ 2))`, `<lit|func@fn1 num@3 num@3>`, ``},
		{`(''+ test)`, `<call|char@1>`, ``},
		{`([]+ test)`, `<call|idxr@1>`, ``},
		{`<@test.point>`, `<lit|typ|obj@test.point>`, `<obj@test.point>`},
	}
	tval, _ := typ.Parse("<obj@test.point x:int y:int>")
	env := &lib.DotEnv{Par: extlib.Std, Lets: lit.MakeObj(lit.Keyed{
		{Key: "test", Val: &lit.Keyed{{Key: "point", Val: tval}}},
	})}
	for _, test := range tests {
		e, err := exp.Parse(test.raw)
		if err != nil {
			t.Errorf("read %s failed: %v", test.raw, err)
			continue
		}
		p := exp.NewProg(env)
		got, err := p.Resl(p, e, typ.Void)
		if err != nil {
			t.Errorf("resl %s failed: %v", test.raw, err)
			continue
		}
		ts := got.Type().String()
		if ts != test.want {
			t.Errorf("resl %s want res %s got %s", test.raw, test.want, ts)
		}
		if test.sig == "" {
			continue
		}
		c, ok := got.(*exp.Call)
		if ok {
			ss := c.Sig.String()
			if ss != test.sig {
				t.Errorf("resl %s want sig %s got %s", test.raw, test.sig, ss)
			}
		} else {
			ss := got.String()
			if ss != test.sig {
				t.Errorf("resl %s want res %s got %s", test.raw, test.sig, ss)
			}
		}
	}
}

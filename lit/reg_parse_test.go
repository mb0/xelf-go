package lit

import (
	"reflect"
	"testing"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

func TestRead(t *testing.T) {
	tests := []struct {
		Val
		str, out, jsn string
	}{
		{typ.Void, `<>`, ``, `"<>"`},
		{typ.ListOf(typ.Opt(typ.Int)), `<list|int?>`, ``, `"<list|int?>"`},
		{Null{}, `null`, ``, ``},
		{Bool(true), `true`, ``, ``},
		{Bool(false), `false`, ``, ``},
		{Num(0), `0`, ``, ``},
		{Num(23), `23`, ``, ``},
		{Num(-23), `-23`, ``, ``},
		{Real(0), `0.0`, `0`, `0`},
		{Real(-0.2), `-0.2`, ``, ``},
		{Char("test"), `"test"`, `'test'`, ``},
		{Char("test"), `'test'`, ``, `"test"`},
		{Char("te\"st"), `'te"st'`, ``, `"te\"st"`},
		{Char("te\"st"), `"te\"st"`, `'te"st'`, ``},
		{Char("te'st"), `'te\'st'`, ``, `"te'st"`},
		{Char("te'st"), `"te'st"`, `'te\'st'`, ``},
		{Char("te\\n\\\"st"), "`" + `te\n\"st` + "`", `'te\\n\\"st'`, `"te\\n\\\"st"`},
		{Char("♥♥"), `'\u2665\u2665'`, `'♥♥'`, `"♥♥"`},
		{Char("😎"), `'\ud83d\ude0e'`, `'😎'`, `"😎"`},
		{Char("2019-01-17"), `'2019-01-17'`, ``, `"2019-01-17"`},
		{&Vals{Num(1), Num(2), Num(3)}, `[1,2,3]`, `[1 2 3]`, ``},
		{&Vals{Num(1), Num(2), Num(3)}, `[1,2,3,]`, `[1 2 3]`, `[1,2,3]`},
		{&Vals{Num(1), Num(2), Num(3)}, `[1 2 3]`, ``, `[1,2,3]`},
		{&Keyed{{"a", Num(1)}, {"b", Num(2)}, {"c", Num(3)}},
			`{"a":1,"b":2,"c":3}`,
			`{a:1 b:2 c:3}`, ``,
		},
		{&Keyed{{"a", Num(1)}, {"b", Num(2)}, {"c", Num(3)}},
			`{"a":1,"b":2,"c":3,}`,
			`{a:1 b:2 c:3}`,
			`{"a":1,"b":2,"c":3}`,
		},
		{&Keyed{{"a", Num(1)}, {"b", Num(2)}, {"c", Num(3)}},
			`{"a":1 "b":2 "c":3}`,
			`{a:1 b:2 c:3}`,
			`{"a":1,"b":2,"c":3}`,
		},
		{&Keyed{{"a", Num(1)}, {"b c", Num(2)}, {"+foo", Char("bar")}},
			`{a:1, "b c":2 '+foo':'bar'}`,
			`{a:1 'b c':2 '+foo':'bar'}`,
			`{"a":1,"b c":2,"+foo":"bar"}`,
		},
	}
	reg := &Reg{Cache: &Cache{}}
	for _, test := range tests {
		v, err := Parse(reg, test.str)
		if err != nil {
			t.Errorf("read %s err %v", test.str, err)
			continue
		}
		if wreg, ok := test.Val.(interface{ WithReg(*Reg) }); ok {
			wreg.WithReg(reg)
		}
		if !reflect.DeepEqual(test.Val, v) {
			t.Errorf("%s want %+v got %+v", test.str, test.Val, v)
		}
		want := strOr(test.out, test.str)
		if got := bfr.String(v); want != got {
			t.Errorf("want xelf %s got %s", want, got)
		}
		buf, err := v.MarshalJSON()
		if err != nil {
			t.Errorf("marshal %s err %v", test.str, err)
			continue
		}
		want = strOr(test.jsn, test.str)
		if got := string(buf); want != got {
			t.Errorf("want xelf %s got %s", want, got)
		}
	}
}

func strOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

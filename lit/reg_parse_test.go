package lit

import (
	"reflect"
	"testing"

	"xelf.org/xelf/bfr"
)

func TestRead(t *testing.T) {
	tests := []struct {
		Val
		str, out, jsn string
	}{
		{Null{}, `null`, ``, ``},
		{Bool(true), `true`, ``, ``},
		{Bool(false), `false`, ``, ``},
		{Int(0), `0`, ``, ``},
		{Int(23), `23`, ``, ``},
		{Int(-23), `-23`, ``, ``},
		{Real(0), `0.0`, `0`, `0`},
		{Real(-0.2), `-0.2`, ``, ``},
		{Str("test"), `"test"`, `'test'`, ``},
		{Str("test"), `'test'`, ``, `"test"`},
		{Str("te\"st"), `'te"st'`, ``, `"te\"st"`},
		{Str("te\"st"), `"te\"st"`, `'te"st'`, ``},
		{Str("te'st"), `'te\'st'`, ``, `"te'st"`},
		{Str("te'st"), `"te'st"`, `'te\'st'`, ``},
		{Str("te\\n\\\"st"), "`" + `te\n\"st` + "`", `'te\\n\\"st'`, `"te\\n\\\"st"`},
		{Str("♥♥"), `'\u2665\u2665'`, `'♥♥'`, `"♥♥"`},
		{Str("😎"), `'\ud83d\ude0e'`, `'😎'`, `"😎"`},
		{Str("2019-01-17"), `'2019-01-17'`, ``, `"2019-01-17"`},
		{&List{Vals: []Val{Int(1), Int(2), Int(3)}}, `[1,2,3]`, `[1 2 3]`, ``},
		{&List{Vals: []Val{Int(1), Int(2), Int(3)}}, `[1,2,3,]`, `[1 2 3]`, `[1,2,3]`},
		{&List{Vals: []Val{Int(1), Int(2), Int(3)}}, `[1 2 3]`, ``, `[1,2,3]`},
		{&Dict{Keyed: []KeyVal{{"a", Int(1)}, {"b", Int(2)}, {"c", Int(3)}}},
			`{"a":1,"b":2,"c":3}`,
			`{a:1 b:2 c:3}`, ``,
		},
		{&Dict{Keyed: []KeyVal{{"a", Int(1)}, {"b", Int(2)}, {"c", Int(3)}}},
			`{"a":1,"b":2,"c":3,}`,
			`{a:1 b:2 c:3}`,
			`{"a":1,"b":2,"c":3}`,
		},
		{&Dict{Keyed: []KeyVal{{"a", Int(1)}, {"b", Int(2)}, {"c", Int(3)}}},
			`{"a":1 "b":2 "c":3}`,
			`{a:1 b:2 c:3}`,
			`{"a":1,"b":2,"c":3}`,
		},
		{&Dict{Keyed: []KeyVal{{"a", Int(1)}, {"b c", Int(2)}, {"+foo", Str("bar")}}},
			`{a:1, "b c":2 '+foo':'bar'}`,
			`{a:1 'b c':2 '+foo':'bar'}`,
			`{"a":1,"b c":2,"+foo":"bar"}`,
		},
	}
	reg := &Reg{}
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

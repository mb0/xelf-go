package lit

import "testing"

func TestDiff(t *testing.T) {
	tests := []struct {
		a, b string
		want string
	}{
		{`[1]`, `[1]`, `{}`},
		{`[1]`, `[1 2]`, `{'.+':[2]}`},
		{`[1]`, `[1 2 3]`, `{'.+':[2 3]}`},
		{`[1 2 3]`, `[1 4 3]`, `{'.1':4}`},
		{`[1]`, `[2 1]`, `{'.*':[[2]]}`},
		{`[1]`, `[2 3 1]`, `{'.*':[[2 3]]}`},
		{`[1 2]`, `[1 3 2]`, `{'.*':[1 [3]]}`},
		{`[1 2]`, `[1 3 4 2]`, `{'.*':[1 [3 4]]}`},
		{`null`, `[1 2]`, `{'.':[1 2]}`},
		{`[1 2]`, `null`, `{'.':null}`},
		{`{a:1}`, `{a:2}`, `{a:2}`},
		{`{a:1 b:2}`, `{a:2}`, `{a:2 '.b-':null}`},
		{`{a:[1 2]}`, `{a:[1 3 2]}`, `{'.a*':[1 [3]]}`},
	}
	reg := &Reg{}
	for _, test := range tests {
		a, err := Parse(reg, test.a)
		if err != nil {
			t.Errorf("parse a %s: %v", test.a, err)
			continue
		}
		b, err := Parse(reg, test.b)
		if err != nil {
			t.Errorf("parse b %s: %v", test.b, err)
			continue
		}
		d, err := Delta(a, b)
		if err != nil {
			t.Errorf("delta failed %s %s: %v", test.a, test.b, err)
			continue
		}
		dict := &Dict{Keyed: d}
		got := dict.String()
		if got != test.want {
			t.Errorf("for %s and %s want %s got %s", test.a, test.b, test.want, got)
		}
	}
}

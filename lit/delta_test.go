package lit

import (
	"testing"

	"xelf.org/xelf/bfr"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		a, b string
		want string
	}{
		{`1`, `1`, `{}`},
		{`1`, `2`, `{.:2}`},
		{`'1'`, `'12'`, `{.+:'2'}`},
		{`'12'`, `'1342'`, `{.*:[1 '34']}`},
		{`[1]`, `[1]`, `{}`},
		{`[1]`, `[1 2]`, `{.+:[2]}`},
		{`[1]`, `[1 2 3]`, `{.+:[2 3]}`},
		{`[1 2 3]`, `[1 4 3]`, `{.1:4}`},
		{`[1]`, `[2 1]`, `{.*:[[2]]}`},
		{`[1]`, `[2 3 1]`, `{.*:[[2 3]]}`},
		{`[1 2]`, `[1 3 2]`, `{.*:[1 [3]]}`},
		{`[1 2]`, `[1 3 4 2]`, `{.*:[1 [3 4]]}`},
		{`null`, `[1 2]`, `{.:[1 2]}`},
		{`[1 2]`, `null`, `{.;}`},
		{`{a:1}`, `{a:2}`, `{a:2}`},
		{`{}`, `{a.b:2}`, `{$:['a.b' 2]}`},
		{`{}`, `{' ':2}`, `{$:[' ' 2]}`},
		{`{}`, `{$:2}`, `{$:['$' 2]}`},
		{`{a:1 b:2}`, `{a:2}`, `{a:2 b-;}`},
		{`{a:1 b:2}`, `{b:2}`, `{a-;}`},
		{`{a:[1 2]}`, `{a:[1 3 2]}`, `{a*:[1 [3]]}`},
		{`{' ':[1 2]}`, `{' ':[1 3 2]}`, `{$*:[' ' [1 [3]]]}`},
		{`{a:[[1 2]]}`, `{a:[[1 3]]}`, `{a.0.1:3}`},
	}
	for _, test := range tests {
		a, err := Parse(test.a)
		if err != nil {
			t.Errorf("parse a %s: %v", test.a, err)
			continue
		}
		b, err := Parse(test.b)
		if err != nil {
			t.Errorf("parse b %s: %v", test.b, err)
			continue
		}
		d, err := Diff(a, b)
		if err != nil {
			t.Errorf("delta failed %s %s: %v", test.a, test.b, err)
			continue
		}
		got := d.String()
		if got != test.want {
			t.Errorf("for %s and %s want %s got %s", test.a, test.b, test.want, got)
			continue
		}
		mut := a.Mut()
		mut, err = Apply(mut, d)
		if err != nil {
			t.Errorf("apply failed %s %s: %v", test.a, got, err)
			continue
		}
		bstr := bfr.String(mut)
		if bstr != test.b {
			t.Errorf("apply %s to %s want %s got %s", d, test.a, test.b, bstr)
			continue
		}
	}
}

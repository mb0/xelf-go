package lit

import (
	"fmt"
	"testing"
)

func TestEqual(t *testing.T) {
	reg := &Reg{Cache: &Cache{}}
	tests := []struct {
		a, b string
		want bool
	}{
		{`<void>`, `<void>`, true},
		{`<void>`, `null`, false},
		{`<list|int>`, `<list|int>`, true},
		{`<list|num>`, `<list|int>`, false},
		{`1`, `1`, true},
		{`1`, `2`, false},
		{`0`, `null`, false},
		{`0`, `0.0`, true},
		{`[1]`, `[1]`, true},
		{`[1]`, `[2]`, false},
		{`{a:1}`, `{a:1}`, true},
		{`{a:[1]}`, `{a:[1]}`, true},
		{`[{a:[1]}]`, `[{a:[1]}]`, true},
		{`{a:1}`, `{}`, false},
		{`{a:1}`, `{a:2}`, false},
	}
	for _, test := range tests {
		a, b, err := parsePair(reg, test.a, test.b)
		if err != nil {
			t.Errorf("parse %v", err)
			continue
		}
		got := Equal(a, b)
		if test.want != got {
			t.Errorf("equal %s %s want %v", a, b, test.want)
		}
	}
}

func parsePair(reg *Reg, astr, bstr string) (a, b Val, err error) {
	a, err = Parse(reg, astr)
	if err != nil {
		return nil, nil, fmt.Errorf("parse a %s: %w", astr, err)
	}
	b, err = Parse(reg, bstr)
	if err != nil {
		return nil, nil, fmt.Errorf("parse b %s: %w", bstr, err)
	}
	return
}

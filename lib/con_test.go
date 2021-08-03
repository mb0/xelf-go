package lib

import (
	"testing"

	"xelf.org/xelf/exp"
)

func TestConEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
		typ  string
	}{
		{`(con real 1)`, `1`, "<real>"},
		{`(con str 'ab')`, `ab`, "<str>"},
		{`(con raw "test")`, `test`, "<raw>"},
		{`(con list 1 2)`, `[1 2]`, "<list>"},
		{`(con dict a:1 b:2)`, `{a:1 b:2}`, "<dict>"},
	}
	for _, test := range tests {
		got, err := exp.Eval(nil, Core, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		str := got.String()
		tstr := got.Val.Type().String()
		if str != test.want || tstr != test.typ {
			t.Errorf("eval %s want %s %s got %s %s",
				test.raw, test.want, test.typ, str, tstr)
		}
	}
}

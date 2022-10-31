package lib

import (
	"testing"

	"xelf.org/xelf/exp"
)

func TestMakeEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
		typ  string
	}{
		{`(make real 1)`, `1`, "<real>"},
		{`(make str 'ab')`, `ab`, "<str>"},
		{`(make raw "test")`, `test`, "<raw>"},
		{`(make list 1 2)`, `[1 2]`, "<list|any>"},
		{`(make dict a:1 b:2)`, `{a:1 b:2}`, "<dict|any>"},
	}
	for _, test := range tests {
		got, err := exp.NewProg(nil, nil, Core).RunStr(test.raw, nil)
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

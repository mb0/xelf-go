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
		{`(make int)`, `0`, "<int>"},
		{`(make int 1)`, `1`, "<int>"},
		{`(make int null)`, `0`, "<int>"},
		{`(make real 1)`, `1`, "<real>"},
		{`(make str 'ab')`, `ab`, "<str>"},
		{`(make raw "test")`, `test`, "<raw>"},
		{`(make list + 1 2)`, `[1 2]`, "<list>"},
		{`(make list|int + 1 2)`, `[1 2]`, "<list|int>"},
		{`(make dict a:1 b:2)`, `{a:1 b:2}`, "<dict>"},
		{`(make dict|int .:{a:1} b:2)`, `{a:1 b:2}`, "<dict|int>"},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Core).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		str := got.String()
		tstr := got.Type().String()
		if str != test.want || tstr != test.typ {
			t.Errorf("eval %s want %s %s got %s %s",
				test.raw, test.want, test.typ, str, tstr)
		}
	}
}

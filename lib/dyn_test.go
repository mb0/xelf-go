package lib

import (
	"testing"

	"xelf.org/xelf/exp"
)

func TestDynEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
		typ  string
	}{
		{`(1 2 3)`, `6`, "<num>"},
		{`(real 1)`, `1`, "<real>"},
		{`(raw 'test')`, `test`, "<raw>"},
		{`('a' 'b' 'c')`, `abc`, "<str>"},
		{`('a' (json ['b']) (xelf 'c'))`, "a[\"b\"]'c'", "<str>"},
		{`(let addone:(fn n:int (add .n 1)) (addone 2))`, `3`, "<int>"},
		{`(let addone:(fn (_ 1)) (addone 2))`, `3`, "<num>"},
		{`((if false add sub) 1 2)`, `-1`, "<num>"},
	}
	for _, test := range tests {
		got, err := exp.NewProg(nil, nil, Std).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed\n\t%v", test.raw, err)
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

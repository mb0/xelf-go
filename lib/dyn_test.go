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
		{`(1+ 2 3)`, `6`, "<num>"},
		{`(real 1)`, `1`, "<real>"},
		{`(real)`, `0`, "<real>"},
		{`(raw 'test')`, `test`, "<raw>"},
		{`(cat '1' '2' '3')`, `123`, "<str>"},
		{`('a'+ 'b' 'c')`, `abc`, "<char>"},
		{`('a'+ (json ['b']) (xelf 'c'))`, "a[\"b\"]'c'", "<char>"},
		{`(with (fn n:int (add .n 1)) (. 2))`, `3`, "<int>"},
		{`(with addone:(fn (add _ 1)) (addone 2))`, `3`, "<num>"},
		{`(if false add sub)`, `<form@sub num@2 tupl|num num@2>`, "<form@sub num@2 tupl|num num@2>"},
		{`((if false add sub) 1 2)`, `-1`, "<num>"},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Std).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed\n\t%v", test.raw, err)
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

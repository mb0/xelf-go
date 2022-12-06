package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestSelEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Val
	}{
		{`(with {a:1 b:2} (sel .$ 'a'))`, lit.Num(1)},
		{`(with {a:1 b:2} (sel $))`, &lit.Vals{lit.Str("arg")}},
		{`(with {a:1 b:2} (sel $.-1))`, lit.Str("arg")},
		{`(with {a:1 b:2} (sel $.$ -1))`, lit.Str("arg")},
		{`(with {a:1 b:2} (sel '.$' 'b'))`, lit.Num(2)},
		{`(with [1 2] (sel .$ 0))`, lit.Num(1)},
		{`(with [1 2] (sel .$ -1))`, lit.Num(2)},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Std).RunStr(test.raw, &lit.Vals{lit.Str("arg")})
		if err != nil {
			t.Errorf("eval %s failed:\n%v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Value(), test.want.Value()) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s",
				test.raw, test.want.Value(), got.Value())
		}
	}
}

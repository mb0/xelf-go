package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestDeclEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Val
	}{
		{`(with 1 .)`, lit.Num(1)},
		{`(with 1 (add 2 .))`, lit.Num(3)},
		{`(with {a:1 b:2} (add .a .b))`, lit.Num(3)},
		{`(let a:1 b:2 (add a b))`, lit.Num(3)},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Std).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got.Val)
		}
	}
}

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
		{`(dot 1 .)`, lit.Int(1)},
		{`(dot 1 (add 2 .))`, lit.Int(3)},
		{`(dot {a:1 b:2} (add .a .b))`, lit.Int(3)},
		{`(let a:1 b:2 (add a b))`, lit.Int(3)},
	}
	for _, test := range tests {
		got, err := exp.Eval(nil, Std, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got.Val)
		}
	}
}

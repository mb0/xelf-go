package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestFuncEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Val
	}{
		{`((fn 1))`, lit.Int(1)},
		{`((fn (add _ 1)) 2)`, lit.Int(3)},
		{`((fn n:int (add .n 1)) 2)`, lit.Int(3)},
		{`((fn a:int b:int (sub .a .b)) 1 2)`, lit.Int(-1)},
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

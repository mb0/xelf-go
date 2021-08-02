package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestMathEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Val
	}{
		{`1`, lit.Int(1)},
		{`-1.2`, lit.Real(-1.2)},
		{`(neg 1)`, lit.Int(-1)},
		{`(neg 1.2)`, lit.Real(-1.2)},
		{`(neg -1.2)`, lit.Real(1.2)},
		{`(abs 1.2)`, lit.Real(1.2)},
		{`(abs -1.2)`, lit.Real(1.2)},
		{`(add 1 2 3)`, lit.Int(6)},
		{`(sub 1 2 3)`, lit.Int(-4)},
		{`(sub 3 2 1)`, lit.Int(0)},
		{`(mul 1 2 3)`, lit.Int(6)},
		{`(div 5 2)`, lit.Real(2.5)},
		{`(div 6 2)`, lit.Int(3)},
		{`(add 1 2.1 3)`, lit.Real(6.1)},
		{`(rem 6 2)`, lit.Int(0)},
		{`(rem 5 2)`, lit.Int(1)},
		{`(rem 5 3)`, lit.Int(2)},
		{`(rem -5 -3)`, lit.Int(-2)},
		{`(rem -5 3)`, lit.Int(-2)},
		{`(min 1 2 3)`, lit.Int(1)},
		{`(max 1 2 3)`, lit.Int(3)},
	}
	for _, test := range tests {
		got, err := exp.Eval(nil, Core, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got.Val)
		}
	}
}

package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/lit"
)

func TestMathEval(t *testing.T) {
	tests := []struct {
		raw  string
		kind knd.Kind
		want lit.Val
	}{
		{`1`, knd.Num, lit.Num(1)},
		{`-1.2`, knd.Real, lit.Real(-1.2)},
		{`(neg 1)`, knd.Num, lit.Num(-1)},
		{`(neg 1.2)`, knd.Real, lit.Real(-1.2)},
		{`(neg -1.2)`, knd.Real, lit.Real(1.2)},
		{`(abs 1.2)`, knd.Real, lit.Real(1.2)},
		{`(abs -1.2)`, knd.Real, lit.Real(1.2)},
		{`(add 1 2 3)`, knd.Num, lit.Num(6)},
		{`(add (make int 1) 2 3)`, knd.Int, lit.Int(6)},
		{`(add (make real 1) 2 3)`, knd.Real, lit.Real(6)},
		{`(sub 1 2 3)`, knd.Num, lit.Num(-4)},
		{`(sub 3 2 1)`, knd.Num, lit.Num(0)},
		{`(mul 1 2 3)`, knd.Num, lit.Num(6)},
		{`(div 5 2)`, knd.Num, lit.Real(2.5)},
		{`(div 6 2)`, knd.Num, lit.Num(3)},
		{`(add 1 2.1 3)`, knd.Num, lit.Real(6.1)},
		{`(rem 6 2)`, knd.Int, lit.Int(0)},
		{`(rem 5 2)`, knd.Int, lit.Int(1)},
		{`(rem 5 3)`, knd.Int, lit.Int(2)},
		{`(rem -5 -3)`, knd.Int, lit.Int(-2)},
		{`(rem -5 3)`, knd.Int, lit.Int(-2)},
		{`(min 1 2 3)`, knd.Num, lit.Num(1)},
		{`(max 1 2 3)`, knd.Num, lit.Num(3)},
	}
	for _, test := range tests {
		got, err := exp.NewProg(nil, nil, Core).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if got.Res.Kind&^knd.Var != test.kind {
			t.Errorf("eval %s want kind %s got %s", test.raw, knd.Name(test.kind), knd.Name(got.Res.Kind))
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got.Val)
		}
	}
}

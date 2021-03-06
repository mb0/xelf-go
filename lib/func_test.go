package lib

import (
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestFuncEval(t *testing.T) {
	reg := &lit.Reg{}
	tests := []struct {
		raw  string
		typ  string
		want string
	}{
		{`((fn 1))`, `<num@3>`, `1`},
		{`((fn (add _ 1)) 2)`, `<num@3>`, `3`},
		{`((fn n:int (add .n 1)) 2)`, `<num@4>`, `3`},
		{`((fn a:int b:int (sub .a .b)) 1 2)`, `<num@4>`, `-1`},
		{`((fn n:int (if (le _ 2) 1 (add (recur (sub _ 1)) (recur (sub _ 2))))) 12)`,
			`<num@5>`, `144`},
		{`(fold (range 12 (fn (sub 12 _))) [1 1]
			(fn a:list|int n:int (if (le .n 2) .a (list (add .a.0 .a.1) .a.0)))
		)`, `<list|int>`, `[144 89]`},
	}
	for _, test := range tests {
		res, err := exp.Eval(nil, reg, Std, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		gott := res.Res.String()
		if gott != test.typ {
			t.Errorf("eval %s want typ %s got %s", test.raw, test.typ, gott)
		}
		got := res.String()
		if got != test.want {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got)
		}
	}
}

package lib

import (
	"testing"

	"xelf.org/xelf/exp"
)

func TestContEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`(len null)`, "0"},
		{`(len "test")`, "4"},
		{`(len [1 2 3])`, "3"},
		{`(fold [1 2 3] "" (fn (cat _ .1)))`, "123"},
		{`(foldr [1 2 3] "" (fn (cat _ .1)))`, "321"},
		// map
		{`(fold [1 2 3] [] (fn
			(mut _ + (mul .1 2))
		))`, "[2 4 6]"},
		// filter
		{`(fold [1 2 3] [] (fn
			(if (ne 0 (rem .1 2)) (mut .0 + .1) .0)
		))`, "[1 3]"},
		{`(range 4)`, "[0 1 2 3]"},
		{`(range 4 (fn (add _ 1)))`, "[1 2 3 4]"},
		{`(range 4)`, "[0 1 2 3]"},
		{`(range 4 (fn (''+ _)))`, "['0' '1' '2' '3']"},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Std).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		str := got.String()
		if str != test.want {
			t.Errorf("eval %s want %s got %s", test.raw, test.want, str)
		}
	}
}

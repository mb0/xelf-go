package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestCallEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Val
	}{
		{`(call add 1 2 3)`, lit.Num(6)},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Std).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s",
				test.raw, test.want, got)
		}
	}
}

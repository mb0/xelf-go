package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestCharEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Val
	}{
		{`(cat 'Hallo' 'Welt' '!')`, lit.Str("HalloWelt!")},
		{`(sep ' ' 'Hallo' 'Welt' '!')`, lit.Str("Hallo Welt !")},
		{`(json 'Hallo')`, lit.Raw("\"Hallo\"")},
		{`(xelf 'Hallo')`, lit.Raw("'Hallo'")},
	}
	for _, test := range tests {
		got, err := exp.Eval(nil, Core, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s",
				test.raw, test.want, got.Val)
		}
	}
}

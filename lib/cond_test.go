package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestCondEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Val
	}{
		{`(if true 1 2)`, lit.Num(1)},
		{`(if false 1 2)`, lit.Num(2)},
		{`(if false 1 false (err) 2)`, lit.Num(2)},
		{`(if false (err) true 1 2)`, lit.Num(1)},
		{`(if true 1)`, lit.Num(1)},
		{`(if true 1 (err))`, lit.Num(1)},
		{`(if false 1)`, lit.Num(0)},
		{`(if 0 "zero")`, lit.Char("")},
		{`(if 0 "zero" 1 "one" "err")`, lit.Char("one")},
		{`(if "" "some" "none")`, lit.Char("none")},
		{`(swt 1 1 "one")`, lit.Char("one")},
		{`(swt 1 1 "one" 2 "two")`, lit.Char("one")},
		{`(swt 0 1 "one")`, lit.Char("")},
		{`(swt 0 1 "one" 2 "two")`, lit.Char("")},
		{`(df 0 1 2)`, lit.Num(1)},
		{`(df "" "none")`, lit.Char("none")},
		{`(df null "none")`, lit.Char("none")},
		{`(df "some" (err))`, lit.Char("some")},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Core).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got.Val)
		}
	}
}

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
		{`(if true 1 2)`, lit.Int(1)},
		{`(if false 1 2)`, lit.Int(2)},
		{`(if false 1 false (err) 2)`, lit.Int(2)},
		{`(if false (err) true 1 2)`, lit.Int(1)},
		{`(if true 1)`, lit.Int(1)},
		{`(if true 1 (err))`, lit.Int(1)},
		{`(if false 1)`, lit.Int(0)},
		{`(if 0 "zero")`, lit.Str("")},
		{`(if 0 "zero" 1 "one" "err")`, lit.Str("one")},
		{`(if "" "some" "none")`, lit.Str("none")},
		{`(swt 1 1 "one")`, lit.Str("one")},
		{`(swt 1 1 "one" 2 "two")`, lit.Str("one")},
		{`(swt 0 1 "one")`, lit.Str("")},
		{`(swt 0 1 "one" 2 "two")`, lit.Str("")},
		{`(df 0 1 2)`, lit.Int(1)},
		{`(df "" "none")`, lit.Str("none")},
		{`(df null "none")`, lit.Str("none")},
		{`(df "some" (err))`, lit.Str("some")},
	}
	for _, test := range tests {
		got, err := exp.Eval(exp.BG, nil, Core, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got.Val)
		}
	}
}

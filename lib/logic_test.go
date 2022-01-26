package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestLogicEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Bool
	}{
		{`true`, true},
		{`(and)`, true},
		{`(and true)`, true},
		{`(and false)`, false},
		{`(and true true)`, true},
		{`(and true false)`, false},
		{`(and false true)`, false},
		{`(and false false)`, false},
		{`(and false (err))`, false},
		{`(or)`, false},
		{`(or true)`, true},
		{`(or false)`, false},
		{`(or true true)`, true},
		{`(or true false)`, true},
		{`(or false true)`, true},
		{`(or false false)`, false},
		{`(or true (err))`, true},
		{`(ok)`, false},
		{`(ok true)`, true},
		{`(ok false)`, false},
		{`(ok true true)`, true},
		{`(ok true false)`, false},
		{`(ok false true)`, false},
		{`(ok false false)`, false},
		{`(ok false (err))`, false},
		{`(not)`, true},
		{`(not true)`, false},
		{`(not false)`, true},
		{`(not true true)`, false},
		{`(not true false)`, false},
		{`(not false true)`, false},
		{`(not false false)`, true},
		{`(not true (err))`, false},
	}
	for _, test := range tests {
		got, err := exp.Eval(nil, nil, Core, test.raw)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got.Val, test.want) {
			t.Errorf("eval %s want %[2]T %[2]s got %[3]T %[3]s", test.raw, test.want, got.Val)
		}
	}
}

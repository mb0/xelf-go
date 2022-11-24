package lib

import (
	"reflect"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lit"
)

func TestTestEval(t *testing.T) {
	tests := []struct {
		raw  string
		want lit.Bool
	}{
		{`(eq 1 2)`, false},
		{`(eq 2 2)`, true},
		{`(eq 1 2 2)`, false},
		{`(eq 2 2 1)`, false},
		{`(eq 2 2 2)`, true},
		{`(eq 2 2)`, true},
		{`(eq 1 2)`, false},
		{`(eq 2 2)`, true},
		{`(lt 1 2 3)`, true},
		{`(lt 2 3 1)`, false},
		{`(lt 3 1 2)`, false},
		{`(lt 3 2 1)`, false},
		{`(lt 1 2 2)`, false},
		{`(lt 1 1 2)`, false},
		{`(le 1 2 3)`, true},
		{`(le 2 3 1)`, false},
		{`(le 3 1 2)`, false},
		{`(le 3 2 2)`, false},
		{`(le 1 2 2)`, true},
		{`(le 1 1 2)`, true},
		{`(gt 1 2 3)`, false},
		{`(gt 2 3 1)`, false},
		{`(gt 3 1 2)`, false},
		{`(gt 3 2 1)`, true},
		{`(gt 1 2 2)`, false},
		{`(gt 1 1 2)`, false},
		{`(ge 1 2 3)`, false},
		{`(ge 2 3 1)`, false},
		{`(ge 3 1 2)`, false},
		{`(ge 3 2 1)`, true},
		{`(ge 3 2 2)`, true},
		{`(ge 1 2 2)`, false},
		{`(ge 1 1 2)`, false},
		{`(in 1 [1 2 3])`, true},
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

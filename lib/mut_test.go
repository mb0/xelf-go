package lib

import (
	"testing"

	"xelf.org/xelf/exp"
)

func TestMutEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`(mut {} a:1 b:2)`, "{a:1 b:2}"},
		{`(mut {a:1} b:2)`, "{a:1 b:2}"},
		{`(mut {a:1} a:2)`, "{a:2}"},
		{`(mut [3 4 1] 1:2)`, "[3 2 1]"},
		{`(mut {a:{}} a.b:1)`, "{a:{b:1}}"},
		{`(mut [3 2]+ 1)`, "[3 2 1]"},
		{`(mut [3]+ 2 1)`, "[3 2 1]"},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Std).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if gstr := got.String(); gstr != test.want {
			t.Errorf("eval %s\n\twant %s\n\tgot  %s %#v", test.raw, test.want, gstr, got)
		}
	}
}

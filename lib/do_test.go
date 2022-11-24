package lib

import (
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/typ"
)

func TestDoEval(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`(do 1 2)`, "2"},
		{`(do 2 1)`, "1"},
	}
	for _, test := range tests {
		got, err := exp.NewProg(Std).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		if gots := got.String(); gots != test.want {
			t.Errorf("eval %s want %s got %s", test.raw, test.want, gots)
		}
	}
}

func TestDoResl(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{`(do 1)`, "1"},
		{`(do 1 2)`, "(do 1 2)"},
	}
	for _, test := range tests {
		x, err := exp.Parse(test.raw)
		if err != nil {
			t.Errorf("parse %s failed: %v", test.raw, err)
			continue
		}
		p := exp.NewProg(Std)
		x, err = p.Resl(p, x, typ.Void)
		if err != nil {
			t.Errorf("resl %s failed: %v", test.raw, err)
			continue
		}
		if got := x.String(); got != test.want {
			t.Errorf("resl %s want %s got %s", test.raw, test.want, got)
		}
	}
}

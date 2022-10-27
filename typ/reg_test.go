package typ

import (
	"fmt"
	"testing"
)

type MapLookup map[string]Type

func (r MapLookup) LookupType(ref string) (Type, error) {
	t, ok := r[ref]
	if !ok {
		return t, fmt.Errorf("unable to resolve ref %q", ref)
	}
	return t, nil
}

func TestReg(t *testing.T) {
	reg := MapLookup{
		"Test":    mustParse(`<obj@Test ID:int Name:str>`),
		"foo.Bar": mustParse(`<obj@foo.Bar ID:int Name:str>`),
		"Bar":     mustParse(`<obj@foo.Bar ID:int Name:str>`),
	}
	tests := []struct {
		raw  string
		want string
	}{
		{"<@Test>", "<obj@Test>"},
		{"<@Test.ID>", "<int@Test.ID>"},
		{"<@foo.Bar>", "<obj@foo.Bar>"},
		{"<@foo.Bar.ID>", "<int@foo.Bar.ID>"},
		{"<@Bar.ID>", "<int@foo.Bar.ID>"},
	}
	for i, test := range tests {
		sys := NewSys()
		raw, err := Parse(test.raw)
		if err != nil {
			t.Errorf("read %s error: %v", test.raw, err)
			continue
		}
		res, err := sys.Inst(reg.LookupType, raw)
		if err != nil {
			t.Errorf("inst %s error: %v", test.raw, err)
			continue
		}
		if got := res.String(); got != test.want {
			t.Errorf("failed test %d\ngot:  %s\nwant: %s", i, got, test.want)
		}
	}
}

func mustParse(raw string) Type {
	t, err := Parse(raw)
	if err != nil {
		panic(err)
	}
	return t
}

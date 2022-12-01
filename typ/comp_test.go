package typ

import "testing"

const (
	wass = 1 << iota
	wconv
	wresl

	wantNone = 0
	wantAss  = wass | wconv | wresl
	wantConv = wconv | wresl
	wantResl = wresl
)

func TestTypeCompatibility(t *testing.T) {
	tests := []struct {
		src, dst string
		want     uint
	}{
		{"void", "void", wantNone},
		{"void", "int", wantNone},
		{"int", "void", wantNone},
		{"int", "int", wantAss},
		{"int", "any", wantAss},
		{"int", "str", wantNone},
		{"int", "bool", wantNone},
		{"list|int", "list|int", wantAss},
		{"list|int", "list|str", wantNone},
		{"list|int", "list|num", wantAss},
		{"list|num", "list|int", wantConv},
		{"list|int", "list", wantAss},
		{"list|int", "any", wantAss},
		{"idxr", "any", wantAss},
		{"list", "list|int", wantConv},
		{"@123", "@123", wantAss},
		{"int", "@1", wantAss},
		{"int", "char@1", wantNone},
		{"int?", "int", wantConv},
		{"int", "int?", wantAss},
		{"int", "num", wantAss},
		{"num", "int", wantConv},
		{"int", "span", wantNone},
		{"time", "char", wantAss},
		{"time", "str", wantNone},
		{"char", "time", wantConv},
		{"str", "time", wantNone},
		{"keyr@1|@2", "dict|int", wantConv},
		{"keyr", "dict", wantConv},
		{"dict", "keyr", wantAss},
		{"dict", "keyr@1", wantAss},
		{"dict", "keyr@1|@2", wantAss},
		{"dict|int", "keyr@1", wantAss},
		{"dict|int", "keyr@1|@2", wantAss},
		{"typ|int", "typ", wantAss},
		{"typ|int", "typ|num", wantAss},
		{"typ|int", "typ|char", wantNone},
		{"typ", "typ|int", wantConv},
		{"<obj x:int y:int>", "<obj x:int y:int z:int>", wantConv},
		{"<obj x:int y:int z:int>", "<obj x:int y:int>", wantAss},
		{"<obj x:int y:int>", "<obj x:int y:int z?:int>", wantAss},
		{"int", "call|int", wantResl},
		{"call|int", "int", wantResl},
		{"call", "sym", wantResl},
		{"call|int", "sym|int", wantResl},
		{"call|sym|typ|int", "sym|typ", wantResl},
	}
	for _, test := range tests {
		src, err := Parse(test.src)
		if err != nil {
			t.Errorf("failed to parse %s: %v", test.src, err)
		}
		dst, err := Parse(test.dst)
		if err != nil {
			t.Errorf("failed to parse %s: %v", test.dst, err)
		}
		got := src.AssignableTo(dst)
		if want := test.want&wass != 0; got != want {
			t.Errorf("assign %s to %s: want %v got %v", test.src, test.dst, want, got)
		}
		got = src.ConvertibleTo(dst)
		if want := test.want&wconv != 0; got != want {
			t.Errorf("convert %s to %s: want %v got %v", test.src, test.dst, want, got)
		}
		got = src.ResolvableTo(dst)
		if want := test.want&wresl != 0; got != want {
			t.Errorf("resolve %s to %s: want %v got %v", test.src, test.dst, want, got)
		}
	}
}

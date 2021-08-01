package typ

import "testing"

func TestAssignable(t *testing.T) {
	tests := []struct {
		src, dst string
		assign   bool
		convert  bool
	}{
		{"void", "void", false, false},
		{"void", "int", false, false},
		{"int", "void", false, false},
		{"int", "int", true, true},
		{"int", "call|int", false, false},
		{"int", "str", false, false},
		{"int", "bool", false, false},
		{"list|int", "list|int", true, true},
		{"list|int", "list|str", false, false},
		{"list|int", "list|num", true, true},
		{"list|int", "list", true, true},
		{"list", "list|int", false, true},
		{"@123", "@123", true, true},
		{"int", "@1", true, true},
		{"int?", "int", true, true},
		{"int", "int?", true, true},
		{"int", "num", true, true},
		{"num", "int", false, true},
		{"int", "span", false, false},
		{"time", "char", true, true},
		{"time", "str", false, false},
		{"char", "time", false, true},
		{"str", "time", false, false},
		{"keyr@1|@2", "dict|int", false, true},
		{"keyr", "dict", false, true},
		{"dict", "keyr", true, true},
		{"dict", "keyr@1", true, true},
		{"dict", "keyr@1|@2", true, true},
		{"dict|int", "keyr@1", true, true},
		{"dict|int", "keyr@1|@2", true, true},
		{"typ|int", "typ", true, true},
		{"typ|int", "typ|num", true, true},
		{"typ|int", "typ|char", false, false},
		{"<rec x:int y:int z:int>", "<rec x:int y:int>", true, true},
		{"<rec x:int y:int>", "<rec x:int y:int z:int>", false, true},
		{"<rec x:int y:int>", "<rec x:int y:int z?:int>", true, true},
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
		if got != test.assign {
			t.Errorf("assign %s to %s: want %v got %v", test.src, test.dst, test.assign, got)
		}
		got = src.ConvertibleTo(dst)
		if got != test.convert {
			t.Errorf("convert %s to %s: want %v got %v", test.src, test.dst, test.convert, got)
		}
	}
}

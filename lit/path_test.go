package lit

import (
	"testing"
)

func TestSelect(t *testing.T) {
	tests := []struct {
		raw  string
		path string
		want string
	}{
		{"null", ".", "null"},
		{"{a:1}", ".a", "1"},
		{"[1 2 3]", "0", "1"},
		{"[1 2 3]", "1", "2"},
		{"[1 2 3]", "-1", "3"},
		{"[1 2 3]", "-2", "2"},
		{"[{a:1} {a:2}]", "0.a", "1"},
		{"[{a:1} {a:2}]", "/a", "[1 2]"},
		{"{a:[{b:1},{b:2}]}", "a/b", "[1 2]"},
	}
	for _, test := range tests {
		on, err := Parse(test.raw)
		if err != nil {
			t.Errorf("parse error: %v", err)
			continue
		}
		r, err := Select(on, test.path)
		if err != nil {
			t.Errorf("parse error: %v", err)
			continue
		}
		got := r.String()
		if got != test.want {
			t.Errorf("%s want %s got %s", test.raw, test.want, got)
		}
	}
}

package typ

import (
	"strings"
	"testing"
)

func TestSelect(t *testing.T) {
	tests := []struct {
		raw  string
		path string
		want string
	}{
		{"<>", ".", "<>"},
		{"<dict>", ".a", "<any>"},
		{"<obj a:str>", "a", "<str>"},
		{"<list|int>", "0", "<int>"},
		{"<list|dict|int>", "0.a", "<int>"},
		{"<list|dict|int>", "/a", "<list|int>"},
		{"<dict|list|dict|int>", "a/b", "<list|int>"},
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
func TestSelectErr(t *testing.T) {
	tests := []struct {
		raw  string
		path string
		want string
	}{
		{"<rect a:str>", "a", "invalid type rect"},
		{"<obj a:string>", "a", "invalid type string"},
		{"<obj a:str>", "b", "key b not found"},
		{"<dict|dict|int>", "/a", "want idxr got <dict|dict|int>"},
	}
	for _, test := range tests {
		on, err := Parse(test.raw)
		if err == nil {
			_, err = Select(on, test.path)
		}
		if err == nil {
			t.Errorf("expect error %s got nil", test.want)
			continue
		}
		got := err.Error()
		if strings.Index(got, test.want) < 0 {
			t.Errorf("%s error want %s got %s", test.raw, test.want, got)
		}
	}
}

package lit

import (
	"strings"
	"testing"

	"xelf.org/xelf/typ"
)

func TestAs(t *testing.T) {
	tests := []struct {
		raw  string
		typ  string
		want string
	}{
		{"5", "", "<num>"},
		{"5", "int", "<int>"},
		{"5", "real", "<real>"},
		{"5", "any", "<any>"},
		{"5", "str", "err:cannot convert *lit.NumMut from <num> to <str>"},
		{"''", "str@a", "<str@a>"},
		{"''", "time", "<time>"},
		{"'hi'", "time", "err:cannot"},
		{"[]", "", "<idxr>"},
		{"[]", "any", "<any>"},
		{"[]", "list|int", "<list|int>"},
		{"[]", "int", "err:cannot"},
		{"['']", "list|int", "err:cannot"},
	}
	for _, test := range tests {
		v, err := Parse(test.raw)
		if err != nil {
			t.Errorf("err parsing %s: %v", test.raw, err)
			continue
		}
		if test.typ != "" {
			tt, err := typ.Parse(test.typ)
			if err != nil {
				t.Errorf("err parsing typ %s for %s: %v", test.typ, test.raw, err)
				continue
			}
			v, err = v.As(tt)
			if err != nil {
				if strings.HasPrefix(test.want, "err:") {
					if !strings.Contains(err.Error(), test.want[4:]) {
						t.Errorf("%s want %s got %v", test.raw, test.want, err)
					}
				} else {
					t.Errorf("err converting %s: %v", test.raw, err)
				}
				continue
			}
		}
		got := v.Type().String()
		if got != test.want {
			t.Errorf("err for %s got %s want %s", test.raw, got, test.want)
		}
	}
}

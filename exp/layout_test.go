package exp

import (
	"strings"
	"testing"

	"xelf.org/xelf/typ"
)

func TestLayout(t *testing.T) {
	tests := []struct {
		sig  string
		args string
		want string
		err  string
	}{
		{"<form _ tupl ?>", "(_ a b x:c)", "(a b x:c)", ""},
		{"<form _ tupl? ?>", "(_)", "()", ""},
		{"<form _ tupl ?>", "(_)", "", "missing argument 0 <exp>"},
		{"<form _ ? tupl ?>", "(_ a b x:c)", "(a) (b x:c)", ""},
		{"<form _ tupl|all tupl|tag tupl ?>", "(_ a b x:c d y:e)", "(a b) (x:c) (d y:e)", ""},
		{"<form _ tupl|any tupl|tag ?>", "(_ x:c)", "", "missing argument 0 <any>"},
		{"<form _ tupl|any tupl|tag ?>", "(_ a b c)", "", "missing argument 1 <tag>"},
		{"<form _ tupl|all tupl?|tag ?>", "(_ a b c)", "(a b c) ()", ""},
		{"<form _ tupl?|all tupl|tag ?>", "(_ x:c)", "() (x:c)", ""},
		{"<func ? list|? ?>", "(_ a b c d)", "(a) (b c d)", ""},
	}
	for _, test := range tests {
		s, err := typ.Parse(test.sig)
		if err != nil {
			t.Errorf("failed to parse typ %s: %v", test.sig, err)
			continue
		}
		e, err := Parse(nil, test.args)
		if err != nil {
			t.Errorf("failed to parse args %s: %v", test.args, err)
			continue
		}
		d := e.(*Call)
		res, err := LayoutSpec(s, d.Args[1:])
		if err != nil {
			if test.err == "" {
				t.Errorf("failed to layout %s %s: %v", test.sig, test.args, err)
				continue
			}
			got := err.Error()
			if !strings.Contains(got, test.err) {
				t.Errorf("expect layout %s err %s got %s", test.sig, test.err, got)
			}
			continue
		}
		var b strings.Builder
		for i, r := range res {
			if i > 0 {
				b.WriteByte(' ')
			}
			b.WriteByte('(')
			b.WriteString(r.String())
			b.WriteByte(')')
		}

		got := b.String()
		if got != test.want {
			t.Errorf("layout %s %s want %s got %s", test.sig, test.args, test.want, got)
		}
	}
}

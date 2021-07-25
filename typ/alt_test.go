package typ

import "testing"

func TestAlt(t *testing.T) {
	tests := []struct {
		a, b Type
		want Type
	}{
		{Void, Void, Void},
		{Int, Void, Int},
		{Int, Int, Int},
		{Data, Int, Data},
		{Num, Int, Num},
		{Opt(Num), Int, Opt(Num)},
		{Opt(Int), Num, Opt(Num)},
		{Opt(Int), Num, Opt(Num)},
		{Alt(Num, Str), Int, Alt(Num, Str)},
	}
	for _, test := range tests {
		got := Alt(test.a, test.b)
		if !test.want.Equal(got) {
			t.Errorf("for %s and %s want %#v got %#v", test.a, test.b, test.want, got)
		}
		rev := Alt(test.b, test.a)
		if !test.want.Equal(rev) {
			t.Errorf("rev %s and %s want %#v got %#v", test.a, test.b, test.want, rev)
		}
	}
}

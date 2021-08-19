package extlib

import "testing"

func TestLike(t *testing.T) {
	tests := []struct {
		s, n string
		want bool
		ign  bool
	}{
		{"test", "test", true, false},
		{"[test]", "_test_", true, false},
		{"test", "%test%", true, false},
		{"[test]", "%test%", true, false},
		{"abcdtestabcd", "%test%", true, false},
		{"dtesta", "%_test_%", true, false},
		{"testab", "%_test_%", false, false},
		{"cdtest", "%_test_%", false, false},
		{"testtest", "%test", true, false},
		{"testatesttesttest", "%test____test%", true, false},
		{"Test", "test", false, false},
		{"test", "Test", false, false},
		{"Test", "test", true, true},
		{"test", "Test", true, true},
		{"Test", "%test%", true, true},
		{"%Test", `\%test%`, true, true},
		{`\%Test`, `\\\%test%`, true, true},
		{`%Test`, `\\\%test%`, false, true},
		// This looks odd, but it is how at least postgres treats unknown escapes
		{`%Test`, `\%\test%`, true, true},
		{`\abcTest`, `\\%\test%`, true, true},
		{"\\abcTest", "\\\\%\\test%", true, true},
		{"\\\nabcTest", "\\\\\n%\\test%", true, true},
		{`abcTest`, `\\%\test%`, false, true},
	}
	for _, test := range tests {
		got := Like(test.s, test.n, test.ign)
		if got != test.want {
			t.Errorf("want like(%s, %s) %v", test.s, test.n, test.want)
		}
	}
}

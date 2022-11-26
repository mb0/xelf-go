package ast

import (
	"strings"
	"testing"
)

var tryCache = `
(try (cache.get $id)
 err:(let res:(calc $)
	(if (not (eq err 'not found')) res
	 else:(try (cache.set $id res)
		res)))
)`[1:]
var minTryCache = "(try(cache.get $id)err:(let res:(calc $)(if(not(eq err'not found'))res else:(try(cache.set $id res)res))))"
var stdTryCache = `(try (cache.get $id) err:(let res:(calc $) (if (not (eq err 'not found')) res else:(try (cache.set $id res) res))))`
var maxTryCache = "(try\n\t(cache.get\n\t\t$id\n\t)\n\terr:(let\n\t\tres:(calc\n\t\t\t$\n\t\t)\n\t\t(if\n\t\t\t(not\n\t\t\t\t(eq\n\t\t\t\t\terr\n\t\t\t\t\t'not found'\n\t\t\t\t)\n\t\t\t)\n\t\t\tres\n\t\t\telse:(try\n\t\t\t\t(cache.set\n\t\t\t\t\t$id\n\t\t\t\t\tres\n\t\t\t\t)\n\t\t\t\tres\n\t\t\t)\n\t\t)\n\t)\n)"

// funny idea: only indet the last calls and see how far it gets us
var lstTryCache = `
(try (cache.get $id) err:(let
	res:(calc $)
	(if (not (eq err 'not found')) res else:(try
		(cache.set $id res)
		res
	))
))`[1:]

var (
	std = &SimpleFormat{}
	min = &SimpleFormat{Flags: FmtMin}
	max = &SimpleFormat{Flags: FmtMax}
	lst = &SimpleFormat{Flags: FmtLst}
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		fmtr Formatter
		raw  string
		want string
	}{
		{name: "std try", fmtr: std, raw: tryCache, want: stdTryCache},
		{name: "min try", fmtr: min, raw: tryCache, want: minTryCache},
		{name: "max try", fmtr: max, raw: tryCache, want: maxTryCache},
		{name: "lst try", fmtr: lst, raw: tryCache, want: lstTryCache},
	}
	for _, test := range tests {
		a, err := Read(strings.NewReader(test.raw), "test")
		if err != nil {
			t.Errorf("%s read err: %v", test.name, err)
			continue
		}
		var b strings.Builder
		err = test.fmtr.Format(&b, a)
		if err != nil {
			t.Errorf("%s fmt err: %v", test.name, err)
			continue
		}
		if got := b.String(); got != test.want {
			t.Errorf("%s fmt got:\n%s%d\n\twant:\n%s%d", test.name,
				got, len(got), test.want, len(test.want))
		}
	}
}

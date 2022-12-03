package mod

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
)

func TestFSMods(t *testing.T) {
	fsmods := FileMods("testdata/lib")
	env := NewLoaderEnv(exp.Builtins(lib.Std), fsmods)
	var reads string
	fsmods.Log = func(root, path string) {
		reads += fmt.Sprintf("\n\tread %s %s", root, path)
	}
	tests := []struct {
		name  string
		raw   string
		local []string
		want  string
	}{
		{"none", "1", nil, "1"},
		{"simple", "(module big a:'A') big.a", []string{"testdata/#big"}, `"A"`},
		{"import foo", "(import './foo') foo.b", []string{"file:testdata/foo.xelf#foo"}, "2"},
		{"import multi", "(import './multi') ([] bar.c spam.e egg.name)", []string{
			"file:testdata/multi.xelf#bar",
			"file:testdata/multi.xelf#egg",
			"file:testdata/multi.xelf#spam",
		}, `[3,5,"ehh"]`},
		{"import multi frag", "(import './multi#bar') bar.c", []string{
			"file:testdata/multi.xelf#bar",
		}, `3`},
		{"import multi frag alias", "(import foo:'./multi#bar') foo.c", []string{
			"file:testdata/multi.xelf#bar",
		}, `3`},
		{"import multi plain alias", "(import bar:'./multi') bar.c", []string{
			"file:testdata/multi.xelf#bar",
		}, `3`},
		{"import liba", "(import 'name.org/liba') liba.name", []string{
			"file:testdata/lib/name.org/liba.xelf#liba",
		}, `"liba"`},
		{"import libb", "(import 'name.org/libb') libb.name", []string{
			"file:testdata/lib/name.org/libb.xelf#libb",
		}, `"libb using liba"`},
		{"import name.org", "(import 'name.org') prod.name", []string{
			"file:testdata/lib/name.org/liba.xelf#liba",
			"file:testdata/lib/name.org/libb.xelf#libb",
			"file:testdata/lib/name.org/mod.xelf#prod",
		}, `"my product with liba and libb using liba"`},
	}
	for _, test := range tests {
		x, err := exp.Parse(test.raw)
		if err != nil {
			t.Errorf("%s parse failed: %v", test.name, err)
			continue
		}
		p := exp.NewProg(env)
		p.File.URL = "testdata/"
		res, err := p.Run(x, nil)
		if err != nil {
			t.Errorf("%s run failed: %v", test.name, err)
			continue
		}
		var local []string
		for _, m := range p.File.Refs {
			local = append(local, m.File.URL+"#"+m.Name)
		}
		sort.Strings(local)
		if !reflect.DeepEqual(local, test.local) {
			t.Errorf("%s got file mods %d %s want %s", test.name, len(local), local, test.local)
		}
		got, _ := res.MarshalJSON()
		if string(got) != test.want {
			t.Errorf("%s got result %s want %s", test.name, got, test.want)
		}
	}
	want := `
	read testdata foo.xelf
	read testdata multi.xelf
	read testdata/lib name.org/liba.xelf
	read testdata/lib name.org/libb.xelf
	read testdata/lib name.org/mod.xelf`
	if reads != want {
		t.Errorf("reads got %s want %s", reads, want)
	}
}

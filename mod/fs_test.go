package mod

import (
	"fmt"
	"reflect"
	"sort"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

func TestFSMods(t *testing.T) {
	fsmods := FileMods("testdata/lib")
	sysmods := &SysMods{files: map[string]*File{
		"sys": {Decls: []ModRef{{Mod: &Mod{Name: "sys"}}}},
	}}
	lenv := NewLoaderEnv(exp.Builtins(lib.Std), sysmods, fsmods)
	var reads string
	fsmods.log = func(root, path string) {
		reads += fmt.Sprintf("\n\tread %s %s", root, path)
	}
	tests := []struct {
		name  string
		raw   string
		local []string
		want  string
	}{
		{"none", "1", nil, "1"},
		{"simple", "(mod big a:'A') big.a", []string{"testdata/#big"}, `"A"`},
		{"use foo", "(use './foo') foo.b", []string{"testdata/foo.xelf#foo"}, "2"},
		{"use multi", "(use './multi') ([] bar.c spam.e egg.name)", []string{
			"testdata/multi.xelf#bar",
			"testdata/multi.xelf#egg",
			"testdata/multi.xelf#spam",
		}, `[3,5,"ehh"]`},
		{"use liba", "(use 'name.org/liba') liba.name", []string{
			"testdata/lib/name.org/liba.xelf#liba",
		}, `"liba"`},
		{"use libb", "(use 'name.org/libb') libb.name", []string{
			"testdata/lib/name.org/libb.xelf#libb",
		}, `"libb using liba"`},
		{"use name.org", "(use 'name.org') prod.name", []string{
			"testdata/lib/name.org/liba.xelf#liba",
			"testdata/lib/name.org/libb.xelf#libb",
			"testdata/lib/name.org/mod.xelf#prod",
		}, `"my product with liba and libb using liba"`},
	}
	for _, test := range tests {
		reg := &lit.Reg{Cache: &lit.Cache{}}
		x, err := exp.Parse(test.raw)
		if err != nil {
			t.Errorf("%s parse failed: %v", test.name, err)
			continue
		}
		p := exp.NewProg(nil, reg, lenv)
		p.File.URL = "testdata/"
		x, err = p.Resl(p, x, typ.Void)
		if err != nil {
			t.Errorf("%s resl failed: %v", test.name, err)
			continue
		}
		res, err := p.Eval(p, x)
		if err != nil {
			t.Errorf("%s eval failed: %v", test.name, err)
			continue
		}
		var local []string
		for _, m := range p.File.Uses {
			local = append(local, m.File.URL+"#"+m.Name)
		}
		sort.Strings(local)
		if !reflect.DeepEqual(local, test.local) {
			t.Errorf("%s got file mods %d %s want %s", test.name, len(local), local, test.local)
		}
		got, _ := res.Val.MarshalJSON()
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

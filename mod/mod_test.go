package mod

import (
	"sort"
	"strings"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
)

func TestSysMods(t *testing.T) {
	setup := func(prog *exp.Prog, s *Src) (*File, error) {
		f := &exp.File{URL: s.URL}
		m := &exp.Mod{File: f, Name: "foo", Decl: lit.MakeObj(lit.Keyed{
			{Key: "b", Val: new(lit.Int)},
		})}
		return f, f.AddRefs(exp.ModRef{Pub: true, Mod: m})
	}
	mods := new(SysMods)
	mods.Register(&Src{
		Rel:   "test/foo",
		Loc:   Loc{URL: "xelf:test/foo"},
		Setup: setup,
	})
	env := NewLoaderEnv(exp.Builtins(lib.Std), mods)
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"import foo", "(import 'test/foo') foo.b", "0"},
		{"mutate foo", "(import 'test/foo') (mut foo.b 2) foo.b", "2"},
		{"import foo as spam", "(import spam:'test/foo') spam.b", "0"},
	}
	for _, test := range tests {
		x, err := exp.Parse(test.raw)
		if err != nil {
			t.Errorf("%s parse failed: %v", test.name, err)
			continue
		}
		p := exp.NewProg(env)
		res, err := p.Run(x, nil)
		if err != nil {
			t.Errorf("%s resl failed: %v", test.name, err)
			continue
		}
		var local []string
		for _, m := range p.File.Refs {
			local = append(local, m.File.URL+"#"+m.Name)
		}
		sort.Strings(local)
		got, _ := res.Val.MarshalJSON()
		if string(got) != test.want {
			t.Errorf("%s got result %s want %s", test.name, got, test.want)
		}
	}
}

func TestFailMods(t *testing.T) {
	fsmods := FileMods()
	env := NewLoaderEnv(exp.Builtins(lib.Std), fsmods)
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"recurse 1", `(import './rec1')`,
			"module load recursion detected for file:testdata/rec1.xelf"},
		{"recurse 2", `(import './rec3')`,
			"sym rec2.Foo unresolved"},
		{"invalid name", `(import './foo')(module Foo)`,
			`invalid module name "Foo"`},
		{"invalid name", `(import './foo')(module foo)`,
			`the module name "foo" is already in use`},
	}
	for _, test := range tests {
		p := exp.NewProg(env)
		p.File.URL = "testdata/"
		_, err := p.RunStr(test.raw, nil)
		if err == nil {
			t.Errorf("%s expect error %s got none", test.name, test.want)
			continue
		}
		got := err.Error()
		if !strings.Contains(got, test.want) {
			t.Errorf("%s error does not contain %s:\n%s", test.name, test.want, got)
		}
	}
}

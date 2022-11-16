package mod

import (
	"sort"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
)

func TestSysMods(t *testing.T) {
	setup := func(prog *exp.Prog, s *Src) (*File, error) {
		f := &exp.File{URL: s.URL}
		decl := exp.LitVal(lit.MakeObj(lit.Keyed{
			{Key: "b", Val: new(lit.Int)},
		}))
		f.Refs = []exp.ModRef{
			{Pub: true, Mod: &exp.Mod{File: f, Name: "foo", Decl: decl}},
		}
		return f, nil
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
		{"use foo", "(use 'test/foo') foo.b", "0"},
		{"mut foo", "(use 'test/foo') (mut foo.b 2) foo.b", "2"},
		{"use foo as spam", "(use spam:'test/foo') spam.b", "0"},
	}
	for _, test := range tests {
		reg := &lit.Reg{Cache: &lit.Cache{}}
		x, err := exp.Parse(test.raw)
		if err != nil {
			t.Errorf("%s parse failed: %v", test.name, err)
			continue
		}
		p := exp.NewProg(nil, reg, env)
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

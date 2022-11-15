package mod

import (
	"log"
	"sort"
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
)

func TestSysMods(t *testing.T) {
	f := &exp.File{URL: "xelf:test/foo"}
	o := lit.MakeObj(lit.Keyed{
		{Key: "b", Val: new(lit.Int)},
	})
	m := &exp.Mod{
		File: f,
		Name: "foo",
		Decl: exp.LitVal(o),
		Setup: func(p *exp.Prog, m *exp.Mod) error {
			log.Printf("setup mod %s", f.URL)
			return nil
		},
	}
	f.Decls = append(f.Decls, exp.ModRef{Mod: m})
	mods := new(SysMods)
	err := mods.Register(f)
	if err != nil {
		t.Errorf("failed to register mod: %v", err)
		return
	}
	env := NewLoaderEnv(exp.Builtins(lib.Std), mods)
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"use foo", "(use 'test/foo') foo.b", "0"},
		{"mut foo", "(use 'test/foo') (mut foo.b 2) foo.b", "2"},
		{"reuse foo", "(use 'test/foo') foo.b", "0"},
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
		for _, m := range p.File.Uses {
			local = append(local, m.File.URL+"#"+m.Name)
		}
		sort.Strings(local)
		got, _ := res.Val.MarshalJSON()
		if string(got) != test.want {
			t.Errorf("%s got result %s want %s", test.name, got, test.want)
		}
	}
}

package ext

import (
	"testing"

	"xelf.org/xelf/exp"
	"xelf.org/xelf/lib"
	"xelf.org/xelf/lit"
)

func TestNode(t *testing.T) {
	c := &lit.PrxReg{}
	def := &Element{Kind: "el", Box: Box{Dim: Dim{W: 3}}}
	spec, err := NodeSpecName(c, "el", def, Rules{IdxKeyer: ZeroKeyer})
	if err != nil {
		t.Fatalf("parse sig error: %v", err)
	}
	nt := spec.Node.Type()
	want := `<obj@ext.Element>`
	if ts := nt.String(); ts != want {
		t.Errorf("want %s got %s", want, ts)
	}
	nt.Ref = ""
	want = `<obj kind:str x?:real y?:real w?:real h?:real font:<obj@ext.Font?> list?:list|.? data?:str>`
	if ts := nt.String(); ts != want {
		t.Errorf("want %s got %s", want, ts)
	}
	env := exp.Builtins(lib.Specs{"el": spec}.AddMap(lib.Std))
	tests := []struct {
		raw  string
		want string
	}{
		{`(el)`, `{kind:'el' w:3 font:null}`},
		{`(el 'test')`, `{kind:'test' w:3 font:null}`},
		{`(el 'test' 1 2)`, `{kind:'test' x:1 y:2 w:3 font:null}`},
		{`(el 'test' 4 3 2 1)`, `{kind:'test' x:4 y:3 w:2 h:1 font:null}`},
		{`(el)`, `{kind:'el' w:3 font:null}`},
		{`(el h:4)`, `{kind:'el' w:3 h:4 font:null}`},
		{`(el font.size:4)`, `{kind:'el' w:3 font:{size:4}}`},
		{`(fold [0 1] [] (fn (append .0 (el font.size:(4 .1)))))`,
			`[{kind:'el' w:3 font:{size:4}} {kind:'el' w:3 font:{size:5}}]`},
	}
	for _, test := range tests {
		got, err := exp.NewProg(env, c).RunStr(test.raw, nil)
		if err != nil {
			t.Errorf("eval %s failed: %v", test.raw, err)
			continue
		}
		vs := got.String()
		if vs != test.want {
			t.Errorf("want %s got %s", test.want, got)
		}
	}
}

type Pos struct {
	X float64 `json:"x,omitempty"`
	Y float64 `json:"y,omitempty"`
}

type Dim struct {
	W float64 `json:"w,omitempty"`
	H float64 `json:"h,omitempty"`
}

type Box struct {
	Pos
	Dim
}

type Font struct {
	Name string  `json:"name,omitempty"`
	Size float64 `json:"size,omitempty"`
}

type Element struct {
	Kind string `json:"kind"`
	Box
	Font *Font      `json:"font"`
	List []*Element `json:"list,omitempty"`
	Data string     `json:"data,omitempty"`
	Calc Box        `json:"-"`
}

package ast

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"xelf.org/xelf/knd"
)

func TestScan(t *testing.T) {
	var noast Ast
	tests := []struct {
		raw  string
		want Ast
		err  string
	}{
		{"0.12", Ast{Tok{Kind: knd.Real, Src: src(0, 4), Raw: "0.12"}, nil}, ""},
		{"[0 0]", Ast{Tok{Kind: knd.List, Rune: '[', Src: src(0, 5)}, []Ast{
			{Tok{Kind: knd.Int, Src: src(1, 1), Raw: "0"}, nil},
			{Tok{Kind: knd.Int, Src: src(3, 1), Raw: "0"}, nil},
		}}, ""},
		{"[1 00]", noast, "test2:1:3: adjacent zeros"},
		{":", Ast{Tok{Kind: knd.Tag, Rune: ':', Src: src(0, 1)}, nil}, ""},
		{";", Ast{Tok{Kind: knd.Tag, Rune: ';', Src: src(0, 1)}, nil}, ""},
		{"{:0}", noast, "test5:1:1: invalid tag"},
		{"{::0}", noast, "test6:1:1: invalid tag"},
		{"{a;}", Ast{Tok{Kind: knd.Dict, Rune: '{', Src: src(0, 4)}, []Ast{
			{Tok{Kind: knd.Tag, Src: src(1, 2), Rune: ';'}, []Ast{
				{Tok{Kind: knd.Sym, Src: src(1, 1), Raw: "a"}, nil},
			}},
		}}, ""},
		{"{a::}", noast, "test8:1:3: invalid tag"},
		{"{a:0}", Ast{Tok{Kind: knd.Dict, Rune: '{', Src: src(0, 5)}, []Ast{
			{Tok{Kind: knd.Tag, Src: src(1, 3), Rune: ':'}, []Ast{
				{Tok{Kind: knd.Sym, Src: src(1, 1), Raw: "a"}, nil},
				{Tok{Kind: knd.Int, Src: src(3, 1), Raw: "0"}, nil},
			}},
		}}, ""},
		{"{\"a\":0}", Ast{Tok{Kind: knd.Dict, Rune: '{', Src: src(0, 7)}, []Ast{
			{Tok{Kind: knd.Tag, Src: src(1, 5), Rune: ':'}, []Ast{
				{Tok{Kind: knd.Str, Src: src(1, 3), Raw: "\"a\""}, nil},
				{Tok{Kind: knd.Int, Src: src(5, 1), Raw: "0"}, nil},
			}},
		}}, ""},
		{"{a:0:}", noast, "test11:1:4: invalid tag"},
	}
	for i, test := range tests {
		doc := fmt.Sprintf("test%d", i)
		got, err := Read(strings.NewReader(test.raw), doc)
		if test.err != "" {
			if err == nil {
				t.Errorf("expect error %s got nil", test.err)
				continue
			}
			if !strings.HasPrefix(err.Error(), test.err) {
				t.Errorf("want error %s got %v", test.err, err)
			}
		} else {
			if err != nil {
				t.Errorf("scan %s: %v", test.raw, err)
				continue
			}
			fixDoc(&test.want, got.Src.Doc)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("want ast %s got %s\n\t%#[1]v\n\t%#[2]v", test.want, got)
			}
		}
	}
}

func src(o, l uint32) Src {
	return Src{
		nil,
		Pos{Line: 1, Byte: int32(o)},
		Pos{Line: 1, Byte: int32(o + l)},
	}
}

func fixDoc(ast *Ast, doc *Doc) {
	ast.Src.Doc = doc
	for i := range ast.Seq {
		fixDoc(&ast.Seq[i], doc)
	}
}

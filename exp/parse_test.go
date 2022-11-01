package exp

import (
	"reflect"
	"strings"
	"testing"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/lit"
	"xelf.org/xelf/typ"
)

func TestParse(t *testing.T) {
	tests := []struct {
		raw  string
		want Exp
	}{
		{`()`, &Call{Src: src(0, 2)}},
		{`<>`, &Lit{typ.Typ, typ.Void, src(0, 2)}},
		{`(() 1 2 3 'things')`, &Call{Src: src(0, 19)}},
		{`(<> 1 2 3 'things')`, &Call{Src: src(0, 19)}},
		{`null`, &Lit{typ.None, lit.Null{}, src(0, 4)}},
		{`1`, &Lit{typ.Num, lit.Num(1), src(0, 1)}},
		{`1.0`, &Lit{typ.Real, lit.Real(1), src(0, 3)}},
		{`bool`, &Sym{Sym: "bool", Src: src(0, 4)}},
		{`name`, &Sym{Sym: "name", Src: src(0, 4)}},
		{`(false)`, call(src(0, 7),
			&Lit{typ.Bool, lit.Bool(false), src(1, 6)},
		)},
		{`(int 1)`, call(src(0, 7),
			&Sym{Sym: "int", Src: src(1, 4)},
			&Lit{typ.Num, lit.Num(1), src(5, 6)},
		)},
		{`(bool 1)`, call(src(0, 8),
			&Sym{Sym: "bool", Src: src(1, 5)},
			&Lit{typ.Num, lit.Num(1), src(6, 7)},
		)},
		{`(bool (() comment) 1)`, call(src(0, 21),
			&Sym{Sym: "bool", Src: src(1, 5)},
			&Lit{typ.Num, lit.Num(1), src(19, 20)},
		)},
		{`<obj x:int y:int>`, &Lit{typ.Typ, typ.Obj("",
			typ.P("x", typ.Int),
			typ.P("y", typ.Int),
		), src(0, 17)}},
		{`('Hello ' $Name '!')`, call(src(0, 20),
			&Lit{typ.Char, lit.Char("Hello "), src(1, 9)},
			&Sym{Sym: "$Name", Src: src(10, 15)},
			&Lit{typ.Char, lit.Char("!"), src(16, 19)},
		)},
		{`(a b; d)`, call(src(0, 8),
			&Sym{Sym: "a", Src: src(1, 2)},
			&Tag{Tag: "b", Src: src(3, 5)},
			&Sym{Sym: "d", Src: src(6, 7)},
		)},
		{`((1 2) 1 2)`, call(src(0, 11),
			call(src(1, 6),
				&Lit{typ.Num, lit.Num(1), src(2, 3)},
				&Lit{typ.Num, lit.Num(2), src(4, 5)},
			),
			&Lit{typ.Num, lit.Num(1), src(7, 8)},
			&Lit{typ.Num, lit.Num(2), src(9, 10)},
		)},
		{`(1 z:(3 4))`, call(src(0, 11),
			&Lit{typ.Num, lit.Num(1), src(1, 2)},
			&Tag{Tag: "z", Src: src(3, 10), Exp: call(src(5, 10),
				&Lit{typ.Num, lit.Num(3), src(6, 7)},
				&Lit{typ.Num, lit.Num(4), src(8, 9)},
			)},
		)},
		{`(s m:(a:(u t;)))`, call(src(0, 16),
			&Sym{Sym: "s", Src: src(1, 2)},
			&Tag{Tag: "m", Src: src(3, 15), Exp: call(src(5, 15),
				&Tag{Tag: "a", Src: src(6, 14), Exp: call(src(8, 14),
					&Sym{Sym: "u", Src: src(9, 10)},
					&Tag{Tag: "t", Src: src(11, 13)},
				)},
			)},
		)},
		{`1 2`, call(ast.Src{},
			&Sym{Sym: "do", Src: ast.Src{}},
			&Lit{typ.Num, lit.Num(1), src(0, 1)},
			&Lit{typ.Num, lit.Num(2), src(2, 3)},
		)},
	}
	for _, test := range tests {
		a, err := ast.ReadAll(strings.NewReader(test.raw), "")
		if err != nil {
			t.Errorf("%s read err: %v", test.raw, err)
		}
		for i := range a {
			fixDoc(&a[i], nil)
		}
		got, err := ParseAll(&lit.Reg{}, a)
		if err != nil {
			t.Errorf("%s parse err: %v", test.raw, err)
			continue
		}
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("%s want:\n%s\n\tgot:\n%s", test.raw, test.want, got)
		}
	}
}

func src(p, e int32) ast.Src {
	return ast.Src{Pos: ast.Pos{Line: 1, Byte: p}, End: ast.Pos{Line: 1, Byte: e}}
}

func call(s ast.Src, xs ...Exp) *Call { return &Call{Src: s, Args: xs} }

func fixDoc(ast *ast.Ast, doc *ast.Doc) {
	ast.Src.Doc = doc
	for i := range ast.Seq {
		fixDoc(&ast.Seq[i], doc)
	}
}

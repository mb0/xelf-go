package typ

import (
	"encoding/json"
	"strings"
	"testing"

	"xelf.org/xelf/knd"
)

func TestParse(t *testing.T) {
	rec1 := Obj("")
	pb := rec1.Body.(*ParamBody)
	pb.Params = append(pb.Params, P("child", ListOf(rec1)))

	rec2 := Obj("")
	pb = rec2.Body.(*ParamBody)
	pb.Params = append(pb.Params, P("body", Obj("", P("child", ListOf(rec2)))))

	tests := []struct {
		typ Type
		raw string
		std string
	}{
		{Void, `void`, `<>`},
		{Void, `<void>`, `<>`},
		{Void, `<>`, ``},
		{None, `none`, `<none>`},
		{None, `<none>`, ``},
		{Data, `data`, `<data>`},
		{Any, `any`, `<any>`},
		{Any, `all?`, `<any>`},
		{Num, `num`, `<num>`},
		{Opt(Num), `num?`, `<num?>`},
		{Typ, `typ`, `<typ>`},
		{TypOf(Num), `typ|num`, `<typ|num>`},
		{VarTyp, `typ@`, `<typ@>`},
		{LitOf(Num), `lit|num`, `<lit|num>`},
		{CallOf(Num), `<call|num>`, ``},
		{Opt(CallOf(Num)), `<call?|num>`, ``},
		{TypOf(ListOf(Int)), `<typ|list|int>`, ``},
		{List, `<list>`, ``},
		{ListOf(Void), `<list>`, ``},
		{ListOf(TypOf(Int)), `<list|typ|int>`, ``},
		{ListOf(TypOf(Int)), `<list <typ int>>`, `<list|typ|int>`},
		{LitOf(Opt(Int)), `<lit|int?>`, ``},
		{ElemTupl(Int), `<tupl int>`, `<tupl|int>`},
		{ElemTupl(Int), `<tupl|int>`, ``},
		{Obj("", P("Name", Str)), `<obj Name:str>`, ``},
		{Type{Kind: knd.Mod}, `<mod>`, ``},
		{Type{Kind: knd.Mod | knd.Ref, Ref: "prod"}, `<mod@prod>`, ``},

		{Var(-1, Void), `@`, `<@>`},
		{Var(1, Void), `<@1>`, ``},
		{Var(1, Num), `<num@1>`, ``},
		{Var(1, Alt(Num, Str)), `<alt@1 num str>`, ``},
		{Ref(`a`), `@a`, `<@a>`},
		{Ref(`a.b`), `<@a.b>`, ``},
		{WithRef(`a.b`, Num), `<num@a.b>`, ``},
		{Sel(`..a`), `<..a>`, ``},
		{ListOf(Var(1, Alt(Num, Str))), `<list|alt@1 num str>`, ``},
		{Opt(Ref(`b`)), `<@b?>`, ``},
		{List, `<list>`, ``},
		{ListOf(Int), `<list|int>`, ``},
		{SymOf(TypOf(ListOf(Int))), `<sym|typ|list|int>`, ``},
		{Opt(Obj("", P(`Name`, Str))), `<obj? Name:str>`, ``},
		{ListOf(Obj("", P(`Name`, Str))), `<list|obj Name:str>`, ``},
		{Obj("", P(`x`, Int), P(`y`, Int)), `<obj x:int y:int>`, ``},
		{Obj("", P("", Ref(`Other`)), P(`Name`, Str)), `<obj @Other Name:str>`, ``},
		{Func("", P(`text`, Str), P(`sub`, Str), P("", Int)),
			`<func text:str sub:str int>`, ``},
		{Form("_", P(`a`, Int), P(`b`, Int), P("", Int)),
			`<form@_ a:int b:int int>`, ``},
		{Form("abs", P(``, Num), P("", Sel(`.0`))),
			`<form@abs num _>`, ``},
		{Form("abs", P(``, Var(-1, Num)), P("", Sel(`.0`))),
			`<form@abs num@ _>`, ``},
		{Obj("", P("child", ListOf(Sel(`.`)))),
			`<obj child:list|.>`, ``},
		{rec1, `<obj child:list|.>`, `-`},
		{rec2, `<obj body:<obj child:list|..>>`, `-`},
	}
	for _, test := range tests {
		raw := test.typ.String()
		want := test.std
		if want == "" || want == "-" {
			want = test.raw
		}
		if raw != want {
			t.Errorf("%s string got %s want %s", test.raw, raw, want)
		}
		typ, err := Read(strings.NewReader(test.raw), "test")
		if err != nil {
			t.Errorf("%s parse error: %v", test.raw, err)
			continue
		}
		if test.std != "-" && !typ.Equal(test.typ) {
			t.Errorf("%s parse\n\twant %v\n\t got %v", test.raw, test.typ, typ)
		}
		rawb, err := json.Marshal(test.typ)
		if err != nil {
			t.Errorf("%s marshal error: %v", test.raw, err)
			continue
		}
		b, err := json.Marshal(want)
		want = string(b)
		if got := string(rawb); got != want {
			t.Errorf("%s marshal got %v", want, got)
		}
		err = json.Unmarshal([]byte(want), &typ)
		if err != nil {
			t.Errorf("%s unmarshal error: %v", want, err)
			continue
		}
		if test.std != "-" && !typ.Equal(test.typ) {
			t.Errorf("%s unmarshal want %v got %v", want, test.typ, typ)
		}
	}
}

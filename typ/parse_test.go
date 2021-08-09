package typ

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
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
		{Obj("test.foo", nil), `<obj test.foo>`, ``},

		{Var(-1, Void), `<@>`, ``},
		{Ref(`a`), `<@a>`, ``},
		{Ref(`a.b`), `<@a.b>`, ``},
		{Sel(`..a`), `<..a>`, ``},
		{Var(1, Void), `<@1>`, ``},
		{Var(1, Num), `<num@1>`, ``},
		{Var(1, Alt(Num, Str)), `<alt@1 num str>`, ``},
		{ListOf(Var(1, Alt(Num, Str))), `<list|alt@1 num str>`, ``},
		{Opt(Ref(`b`)), `<@b?>`, ``},
		{List, `<list>`, ``},
		{ListOf(Int), `<list|int>`, ``},
		{SymOf(TypOf(ListOf(Int))), `<sym|typ|list|int>`, ``},
		{Opt(Rec(P(`Name`, Str))), `<rec? Name:str>`, ``},
		{ListOf(Rec(P(`Name`, Str))), `<list|rec Name:str>`, ``},
		{Rec(P(`x`, Int), P(`y`, Int)), `<rec x:int y:int>`, ``},
		{Rec(P("", Ref(`Other`)), P(`Name`, Str)), `<rec @Other Name:str>`, ``},
		{Func("", P(`text`, Str), P(`sub`, Str), P("", Int)),
			`<func text:str sub:str int>`, ``},
		{Form("_", P(`a`, Int), P(`b`, Int), P("", Int)),
			`<form _ a:int b:int int>`, ``},
	}
	for _, test := range tests {
		raw := test.typ.String()
		want := test.std
		if want == "" {
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
		if !typ.Equal(test.typ) {
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
		if !typ.Equal(test.typ) {
			t.Errorf("%s unmarshal want %v got %v", want, test.typ, typ)
		}
	}
}

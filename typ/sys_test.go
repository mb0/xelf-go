package typ

import (
	"fmt"
	"strings"
	"testing"
)

func TestSys(t *testing.T) {
	s := Func("",
		P("", Func("", P("", Var(1, Void)), P("", Bool))),
		P("", ListOf(Var(1, Void))),
		P("", Sel(".1")),
	)
	want := `<func <func @1 bool> list|@1 .1>`
	if got := s.String(); got != want {
		t.Errorf("want %s\ngot %s", want, got)
	}
	sys := NewSys()
	sys.MaxID = 5
	s, err := sys.Inst(nil, s)
	if err != nil {
		t.Errorf("inst want %s got err %v", want, err)
	}
	want = `<func <func @6 bool> list|@6 list|@6>`
	if got := s.String(); got != want {
		t.Errorf("want inst %s\ngot %s", want, got)
	}
	free := sys.Free(s, nil)
	if len(free) != 1 || free[0].ID != 6 {
		t.Errorf("want free [@6] got %s", free)
	}
	sys.Bind(WithID(6, Int))
	free = sys.Free(s, nil)
	if len(free) != 0 {
		t.Errorf("want free [] got %s", free)
	}
	a, err := sys.Update(s)
	want = `<func <func int bool> list|int list|int>`
	if err != nil {
		t.Errorf("update want %s got err %v", want, err)
	}
	if got := a.String(); got != want {
		t.Errorf("want %s\ngot %s", want, got)
	}
}

func TestInst(t *testing.T) {
	tests := []struct {
		a string
		w string
	}{
		{"<form@a num@ _>", "<form@a num@2 num@2>"},
		// The self selection in child is resolved and then printed as selection again
		{"<obj id:int@ par:.id child:list|.>", "<obj id:int@2 par:int@2 child:list|.>"},
		{"<obj id:int@ body:<obj child:list|..>>", "<obj id:int@2 body:<obj child:list|..>>"},
		{"<form@make typ@ lit|_>", "<form@make typ@2 lit|typ@2>"},
		{"<@decl>", "<str>"},
		{"<@.field>", "<str>"},
		{"<@named>", "<str@qualified.Named>"},
		{"<enum@abc a; b; c;>", "<enum@abc>"},
		{"<obj@info id:int label:str>", "<obj@info>"},
	}
	var lup Lookup = func(key string) (Type, error) {
		switch key {
		case "decl", ".field":
			return Str, nil
		case "named": // essentially the lookup determines the resolved name
			return WithRef("qualified.Named", Str), nil
		}
		return Void, fmt.Errorf("not found")
	}
	for _, test := range tests {
		a, err := Parse(test.a)
		if err != nil {
			t.Errorf("read %s error: %v", test.a, err)
			continue
		}
		sys := NewSys()
		sys.MaxID = 1
		a, err = sys.Inst(lup, a)
		if err != nil {
			t.Errorf("inst error for %s: %v", a, err)
			continue
		}
		if got := a.String(); got != test.w {
			t.Errorf("inst got %s want %s", got, test.w)
		}
	}
}

func TestUnify(t *testing.T) {
	tests := []struct {
		a, b string
		err  string
		w    Type
	}{
		{"int", "int", "", Int},
		{"typ", "typ", "", Typ},
		{"num", "int", "", Int},
		{"int", "num", "", Int},
		{"num", "num", "", Num},
		{"num", "any", "", Num},
		{"int", "real", "cannot", Void},
		{"num", "str", "cannot", Void},
		{"num", "<alt int str>", "", Int},
		{"<alt int str>", "<alt int str>", "", Alt(Int, Str)},
		{"<alt num str>", "<alt int char>", "", Alt(Int, Str)},
		{"<alt real bits>", "<alt int str>", "cannot", Void},
		{"<alt real bits str>", "<alt int str>", "", Str},
		{"<alt@ real bits str>", "<alt int str>", "", WithID(1, Str)},
		{"char", "<alt str cont>", "", Str},
		{"@", "num", "", Var(1, Num)},
		{"@", "int", "", WithID(1, Int)},
		{"int", "@", "", WithID(1, Int)},
		{"@1", "@2", "", Var(1, Void)},
		{"idxr", "list|int", "", ListOf(Int)},
		{"list", "list|int", "", ListOf(Int)},
		{"list|@", "list", "", ListOf(Var(1, Void))},
		{"list|@", "list|int", "", ListOf(WithID(1, Int))},
		{"list|str", "list|int", "cannot", Void},
		{"<obj@foo x:int y:int>", "<obj@foo x:int y:int>", "", Obj("foo", P("x", Int), P("y", Int))},
		{"<obj@foo x:int y:int>", "<obj@bar x:int y:int>", "", Obj("", P("x", Int), P("y", Int))},
		{"<obj@foo x:int y:int>", "any", "", Obj("foo", P("x", Int), P("y", Int))},
		{"<obj@foo x:int y:int>", "idxr", "", Obj("foo", P("x", Int), P("y", Int))},
		{"num@", "exp", "", Var(1, Num)},
		{"num", "exp|@", "", Var(1, Num)},
		{"num", "@", "", Var(1, Num)},
		{"tupl|int", "tupl?", "", ElemTupl(Int)},
		{"tupl?|int", "tupl?", "", Opt(ElemTupl(Int))},
		{"<form@a int any>", "<form@b int any>", "", Form("", P("", Int), P("", Any))},
		{"<form@a int@1 any>", "<form@b int@2? any>", "", Form("", P("", Opt(Int)), P("", Any))},
	}
	for _, test := range tests {
		a, err := Parse(test.a)
		if err != nil {
			t.Errorf("read %s error: %v", test.a, err)
			continue
		}
		b, err := Parse(test.b)
		if err != nil {
			t.Errorf("read %s error: %v", test.b, err)
			continue
		}
		sys := NewSys()
		a, _ = sys.Inst(nil, a)
		b, _ = sys.Inst(nil, b)
		r, err := sys.Unify(a, b)
		if err != nil {
			if test.err == "" {
				t.Errorf("unify ab error for %s %s: %v", r, test.b, err)
				continue
			} else if !strings.Contains(err.Error(), test.err) {
				t.Errorf("unify ab want error %s for %s %s got: %v",
					test.err, test.a, test.b, err)
			}
		} else if test.err != "" {
			t.Errorf("unify ab want error %s for %s %s got none",
				test.err, test.a, test.b)
		}
		if test.err == "" && !r.Equal(test.w) {
			t.Errorf("unify ab for %s %s want %s got %s\n",
				a, b, test.w, r)
		}
	}
}

package typ

import (
	"strings"
	"testing"
)

func TestCtx(t *testing.T) {
	s := Func("",
		P("", Func("", P("", Var(1, Void)), P("", Bool))),
		P("", ListOf(Var(1, Void))),
		P("", Sel(".1")),
	)
	want := `<func <func @1 bool> list|@1 .1>`
	if got := s.String(); got != want {
		t.Errorf("want %s\ngot %s", want, got)
	}
	sys := NewSys(nil)
	sys.MaxID = 5
	s, err := sys.Inst(s)
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
	a := sys.Update(s)
	want = `<func <func int bool> list|int list|int>`
	if got := a.String(); got != want {
		t.Errorf("want %s\ngot %s", want, got)
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
		{"<alt real bits>", "<alt int str>", "cannot", Void},
		{"<alt real bits str>", "<alt int str>", "", Str},
		{"<alt@ real bits str>", "<alt int str>", "", WithID(1, Str)},
		{"char", "<alt str cont>", "", Str},
		{"@", "num", "", Var(1, Num)},
		{"@", "int", "", WithID(1, Int)},
		{"int", "@", "", WithID(1, Int)},
		{"@", "@", "", Var(1, Void)},
		{"list", "list|int", "", ListOf(Int)},
		{"list|@", "list", "", ListOf(Var(1, Void))},
		{"list|@", "list|int", "", ListOf(WithID(1, Int))},
		{"list|str", "list|int", "cannot", Void},
		{"<rec x:int y:int>", "<rec x:int y:int>", "", Rec(P("x", Int), P("y", Int))},
		{"<rec x:int y:int>", "any", "", Rec(P("x", Int), P("y", Int))},
		{"num@", "exp", "", Var(1, Num)},
		{"num", "exp|@", "", Var(1, Num)},
		{"num", "@", "", Var(1, Num)},
		{"tupl|int", "tupl?", "", ElemTupl(Int)},
		{"<form a int any>", "<form b int any>", "", Form("_", P("", Int), P("", Any))},
		{"<form a int any>", "<form b int? any>", "", Form("_", P("", Int), P("", Any))},
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
		sys := NewSys(nil)
		a, _ = sys.Inst(a)
		b, _ = sys.Inst(b)
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

func TestUnifyError(t *testing.T) {
	tests := []struct {
		a, b Type
	}{
		{Num, Char},
		{Var(1, Char), Int},
		{Int, Var(1, Char)},
		{Alt(Num, Int), Char},
		{ListOf(Alt(Num)), ListOf(Char)},
	}
	for _, test := range tests {
		sys := NewSys(nil)
		m := make(map[int32]Type)
		a, _ := sys.inst(test.a, m)
		b, _ := sys.inst(test.b, m)
		r := sys.Bind(Var(0, Void))
		var err error
		r, err = sys.Unify(r, a)
		if err != nil {
			t.Errorf("unify a error for %s %s: %+v", a, b, err)
			continue
		}
		r, err = sys.Unify(r, b)
		if err == nil {
			t.Errorf("unify b want error for %s %s got %s", a, b, r)
		}
	}
}

package lit

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"xelf.org/xelf/bfr"
	"xelf.org/xelf/cor"
	"xelf.org/xelf/typ"
)

type Point struct {
	X, Y int
}

type POI struct {
	Name  string
	Point *Point `json:"point,omitempty"`
}

type Embed struct {
	ID int
	POI
}

func TestEmbed(t *testing.T) {
	reg := &Reg{}
	o := reg.MustProxy(new(Embed)).(Keyr)
	ot := o.Type()
	ot.Ref = ""
	if got, want := ot.String(), "<obj id:int name:str point?:<obj@lit.Point?>>"; got != want {
		t.Errorf("embed want type %s got %s", want, got)
	}
	keys := o.Keys()
	if got, want := strings.Join(keys, " "), "id name point"; got != want {
		t.Errorf("embed want keys %s got %s", want, got)
	}
}

func TestProxy(t *testing.T) {
	reg := &Reg{}
	poi := reg.MustProxy(new(POI))
	tests := []struct {
		val  interface{}
		typ  string
		set  Val
		want string
	}{
		{new(bool), "<bool>", nil, "false"},
		{new(*bool), "<bool?>", nil, "null"},
		{new(*bool), "<bool?>", Bool(true), "true"},
		{new(*typ.Type), "<typ?>", typ.Int, "<int>"},
		{new(int64), "<int>", Int(23), "23"},
		{new(*int64), "<int?>", nil, "null"},
		{new(int32), "<int>", Int(23), "23"},
		{new(*int32), "<int?>", Int(23), "23"},
		{new(*int32), "<int?>", nil, "null"},
		{new([]byte), "<raw>", Raw("foo"), "'foo'"},
		{new(time.Time), "<time>", Null{}, "'0001-01-01T00:00:00Z'"},
		{new([16]byte), "<uuid>",
			UUID(cor.MustParseUUID("b024a8d9-b7bd-4e79-a3d2-2dc255d6da24")),
			"'b024a8d9-b7bd-4e79-a3d2-2dc255d6da24'",
		},
		{new(*time.Time), "<time?>", Null{}, "null"},
		{new(time.Duration), "<span>", Span(time.Hour), "'1:00:00'"},
		{new(Point), "<obj@lit.Point>", Null{}, "{x:0 y:0}"},
		{new(Point), "<obj@lit.Point>", &Dict{Keyed: []KeyVal{{"y", Int(5)}}}, "{x:0 y:5}"},
		{new([]Point), "<list|obj@lit.Point>", Null{}, "[]"},
		{new([]Point), "<list|obj@lit.Point>", &List{Vals: []Val{}}, "[]"},
		{new(*[]Point), "<list?|obj@lit.Point>", Null{}, "null"},
		{new(*[]Point), "<list?|obj@lit.Point>", &List{Vals: []Val{}}, "[]"},
		{new([]Point), "<list|obj@lit.Point>", &List{Vals: []Val{
			&Dict{Keyed: []KeyVal{{"y", Int(5)}}},
		}}, "[{x:0 y:5}]"},
		{new(*Point), "<obj@lit.Point?>", Null{}, "null"},
		{new(*Point), "<obj@lit.Point?>", &Dict{Keyed: []KeyVal{{"y", Int(5)}}}, "{x:0 y:5}"},
		{new(*POI), "<obj@lit.POI?>", &Dict{Keyed: []KeyVal{{"name", Char("foo")}}}, "{name:'foo'}"},
		{new(*POI), "<obj@lit.POI?>", poi, "{name:''}"},
		{new(map[string]Point), "<dict|obj@lit.Point>", Null{}, "{}"},
		{new(map[string]Point), "<dict|obj@lit.Point>",
			&Dict{Keyed: []KeyVal{{"a", &Dict{Keyed: []KeyVal{{"y", Int(5)}}}}}},
			"{a:{x:0 y:5}}"},
	}
	for _, test := range tests {
		p, err := reg.Proxy(test.val)
		if err != nil {
			t.Errorf("proxy %T error: %v", test.val, err)
			continue
		}
		gt := p.Type().String()
		if gt != test.typ {
			t.Errorf("typ want %s got %s", test.typ, gt)
		}
		err = p.Assign(test.set)
		if err != nil {
			t.Errorf("assign %s %T error: %v", test.typ, test.val, err)
			continue
		}
		gp := bfr.String(p)
		if gp != test.want {
			t.Errorf("str want %s got %s", test.want, gp)
		}
		gj, err := bfr.JSON(p)
		if err != nil {
			t.Errorf("marshal %T error: %v", p, err)
			continue
		}
		rj, _ := p.New()
		err = json.Unmarshal(gj, rj.Ptr())
		if err != nil {
			t.Errorf("unmarshal %s %#v error: %v", string(gj), p, err)
			continue
		}
	}
}

func TestProxyAll(t *testing.T) {
	reg := &Reg{}
	reg.MustProxy(new(POI))
	tests := []struct {
		val interface{}
		typ string
		def string
		nzv string
	}{
		{new(bool), "<bool>", "false", "true"},
		{new(typ.Type), "<typ>", "<>", "<bool>"},
		{new(int64), "<int>", "0", "1"},
		{new(int32), "<int>", "0", "1"},
		{new([]byte), "<raw>", "''", "'a'"},
		{new(time.Time), "<time>", "'0001-01-01T00:00:00Z'", "'2006-01-02T15:04:05Z'"},
		{new([16]byte), "<uuid>", "'00000000-0000-0000-0000-000000000000'", "'19f0a4d8-c728-43ec-aca0-1a1f33e2de49'"},
		{new(time.Duration), "<span>", "'0'", "'1:00'"},
		{new(Point), "<obj@lit.Point>", "{x:0 y:0}", "{x:1 y:2}"},
		{new([]Point), "<list|obj@lit.Point>", "[]", "[{}]"},
		{new(POI), "<obj@lit.POI>", "{name:''}", "{name:'a'}"},
		{new(map[string]Point), "<dict|obj@lit.Point>", "{}", "{a:{}}"},
	}
	for _, test := range tests {
		// test value
		p, err := reg.Proxy(test.val)
		if err != nil {
			t.Errorf("proxy %T error: %v", test.val, err)
			continue
		}
		gt := p.Type().String()
		if gt != test.typ {
			t.Errorf("typ want typ %s got %s", test.typ, gt)
		}
		gp := bfr.String(p)
		if gp != test.def {
			t.Errorf("str want def %s got %s", test.def, gp)
		}
		testDefault(t, reg, p, false)
	}
}

func testDefault(t *testing.T, reg *Reg, mut Mut, ptr bool) {
	if ptr {
		gotstr := bfr.String(mut)
		if gotstr != "null" {
			t.Errorf("ptr string %T want null got %s", mut, gotstr)
		}
	}
	jsn, err := bfr.JSON(mut)
	if err != nil {
		t.Errorf("marshal %T error: %v", mut, err)
		return
	}
	nmut, err := mut.New()
	if err != nil {
		t.Errorf("new %T error: %v", mut, err)
		return
	}
	err = json.Unmarshal(jsn, nmut.Ptr())
	if err != nil {
		t.Errorf("unmarshal %T error: %v", mut, err)
		return
	}
	err = nmut.Assign(nil)
	if err != nil {
		t.Errorf("assign nil %T error: %v", nmut, err)
		return
	}
	err = nmut.Assign(Null{})
	if err != nil {
		t.Errorf("assign null %T error: %v", nmut, err)
		return
	}
	if !nmut.Zero() {
		t.Errorf("want zero %T", nmut)
	}
	if ptr && !nmut.Nil() {
		t.Errorf("want nil %T", nmut)
	}
	if ptr {
		ppt := reflect.New(reflect.ValueOf(mut.Ptr()).Type())
		pmut, err := reg.ProxyValue(ppt)
		if err != nil {
			t.Errorf("proxy %s error: %v", ppt.Elem(), err)
			return
		}
		testDefault(t, reg, pmut, true)
	}
}

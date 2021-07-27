package lit

import (
	"encoding/json"
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
		{new(int32), "<int>", Int(23), "23"},
		{new(*int32), "<int?>", Int(23), "23"},
		{new([]byte), "<raw>", Raw("foo"), "'foo'"},
		{new(time.Time), "<time>", Null{}, "'0001-01-01T00:00:00Z'"},
		{new([16]byte), "<uuid>",
			UUID(cor.MustParseUUID("b024a8d9-b7bd-4e79-a3d2-2dc255d6da24")),
			"'b024a8d9-b7bd-4e79-a3d2-2dc255d6da24'",
		},
		{new(*time.Time), "<time?>", Null{}, "null"},
		{new(time.Duration), "<span>", Span(time.Hour), "'1:00:00'"},
		{new(Point), "<obj lit.Point>", Null{}, "{x:0 y:0}"},
		{new(Point), "<obj lit.Point>", &Dict{Keyed: []KeyVal{{"y", Int(5)}}}, "{x:0 y:5}"},
		{new([]Point), "<list|obj lit.Point>", Null{}, "[]"},
		{new([]Point), "<list|obj lit.Point>", &List{Vals: []Val{
			&Dict{Keyed: []KeyVal{{"y", Int(5)}}},
		}}, "[{x:0 y:5}]"},
		{new(*Point), "<obj? lit.Point>", Null{}, "null"},
		{new(*Point), "<obj? lit.Point>", &Dict{Keyed: []KeyVal{{"y", Int(5)}}}, "{x:0 y:5}"},
		{new(*POI), "<obj? lit.POI>", &Dict{Keyed: []KeyVal{{"name", Str("foo")}}}, "{name:'foo'}"},
		{new(*POI), "<obj? lit.POI>", poi, "{name:''}"},
		{new(map[string]Point), "<dict|obj lit.Point>", Null{}, "{}"},
		{new(map[string]Point), "<dict|obj lit.Point>",
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

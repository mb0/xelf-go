package lit

import (
	"reflect"
	"testing"
)

func TestProxyAny(t *testing.T) {
	reg := &Reg{Cache: &Cache{}}
	var one interface{}
	err := ParseInto(reg, `1`, &one)
	if err != nil {
		t.Errorf("parse one %#v", err)
	}
	var want interface{} = Num(1)
	if !reflect.DeepEqual(one, want) {
		t.Errorf("want %#v got %#v", want, one)
	}
	var any []interface{}
	err = ParseInto(reg, `[null 1 'test' []]`, &any)
	if err != nil {
		t.Errorf("parse %#v", err)
	}
	wanta := []interface{}{
		Null{}, Num(1), Char("test"), &List{Reg: reg, Vals: []Val{}},
	}
	if !reflect.DeepEqual(any, wanta) {
		t.Errorf("want %v got %v", wanta, any)
	}
}

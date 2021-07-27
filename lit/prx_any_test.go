package lit

import (
	"reflect"
	"testing"
)

func TestProxyAny(t *testing.T) {
	reg := &Reg{}
	var one interface{}
	err := reg.ParseInto(`1`, &one)
	if err != nil {
		t.Errorf("parse one %#v", err)
	}
	var want interface{} = Int(1)
	if !reflect.DeepEqual(one, want) {
		t.Errorf("want %#v got %#v", want, one)
	}
	var any []interface{}
	err = reg.ParseInto(`[null 1 'test' []]`, &any)
	if err != nil {
		t.Errorf("parse %#v", err)
	}
	wanta := []interface{}{
		Null{}, Int(1), Str("test"), &List{Reg: reg, Vals: []Val{}},
	}
	if !reflect.DeepEqual(any, wanta) {
		t.Errorf("want %v got %v", wanta, any)
	}

}

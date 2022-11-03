package lit

import (
	"reflect"
	"testing"
)

func TestProxyAny(t *testing.T) {
	reg := &Reg{Cache: &Cache{}}
	var one interface{}
	mut, err := reg.Proxy(&one)
	if err != nil {
		t.Fatalf("parse one %#v", err)
	}
	err = ParseInto(`1`, mut)
	if err != nil {
		t.Fatalf("parse one %#v", err)
	}
	var want interface{} = Num(1)
	if !reflect.DeepEqual(one, want) {
		t.Errorf("want %#v got %#v", want, one)
	}
	var any []interface{}
	mut, err = reg.Proxy(&any)
	if err != nil {
		t.Fatalf("parse one %#v", err)
	}
	err = ParseInto(`[null 1 'test' []]`, mut)
	if err != nil {
		t.Errorf("parse %#v", err)
	}
	wanta := []interface{}{
		Null{}, Num(1), Char("test"), &Vals{},
	}
	if !reflect.DeepEqual(any, wanta) {
		t.Errorf("want %v got %v", wanta, any)
	}
}

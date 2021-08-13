package lit

import (
	"fmt"
	"reflect"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/typ"
)

type AnyPrx struct {
	proxy
	val Val
}

func newAnyPrx(reg *Reg, t typ.Type) *AnyPrx {
	var any interface{}
	return &AnyPrx{proxy{reg, t, reflect.ValueOf(&any)}, Null{}}
}

func anyVal(v reflect.Value) Val {
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() || v.Elem().Kind() != reflect.Interface {
		panic(fmt.Errorf("invalid anyprx value %s", v.Type()))
	}
	ve := v.Elem()
	if ve.IsNil() {
		return Null{}
	}
	val, ok := ve.Interface().(Val)
	if !ok {
		panic(fmt.Errorf("proxy any failed to get proxy value %v", ve.Interface()))
	}
	return val
}

func (x *AnyPrx) NewWith(v reflect.Value) (Mut, error) {
	return &AnyPrx{x.with(v), anyVal(v)}, nil
}
func (x *AnyPrx) New() (Mut, error) { return x.NewWith(x.new()) }

func (x *AnyPrx) Zero() bool { return x.Nil() || x.val.Zero() }
func (x *AnyPrx) Value() Val { return x.val.Value() }
func (x *AnyPrx) Parse(a ast.Ast) (err error) {
	if isNull(a) {
		x.val = Null{}
	} else {
		x.val, err = x.Reg.ParseVal(a)
		if err != nil {
			return err
		}
	}
	x.elem().Set(reflect.ValueOf(x.val))
	return nil
}

func (x *AnyPrx) Assign(v Val) (err error) {
	if v.Nil() {
		x.val = Null{}
	} else {
		x.val = v
	}
	x.elem().Set(reflect.ValueOf(x.val))
	return nil
}
func (x *AnyPrx) String() string               { return x.val.String() }
func (x *AnyPrx) MarshalJSON() ([]byte, error) { return x.val.MarshalJSON() }
func (x *AnyPrx) UnmarshalJSON(b []byte) error { return unmarshal(b, x) }
func (x *AnyPrx) Print(b *bfr.P) error         { return x.val.Print(b) }

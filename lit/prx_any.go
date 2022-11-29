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

func newAnyPrx(reg *PrxReg, t typ.Type) *AnyPrx {
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

func (x *AnyPrx) NewWith(v reflect.Value) Mut { return &AnyPrx{x.with(v), anyVal(v)} }
func (x *AnyPrx) Unwrap() Val {
	if x.Nil() {
		return Null{}
	}
	return x.val
}
func (x *AnyPrx) New() Mut   { return x.NewWith(x.new()) }
func (x *AnyPrx) Zero() bool { return x.Nil() || x.val.Zero() }
func (x *AnyPrx) Mut() Mut   { return x }
func (x *AnyPrx) Value() Val { return x.Unwrap().Value() }
func (x *AnyPrx) As(t typ.Type) (Val, error) {
	if x.typ == t {
		return x, nil
	}
	return &AnyPrx{x.typed(t), x.val}, nil
}
func (x *AnyPrx) Parse(a ast.Ast) (err error) {
	if isNull(a) {
		x.val = Null{}
	} else {
		x.val, err = ParseVal(a)
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
func (x *AnyPrx) String() string               { return x.Unwrap().String() }
func (x *AnyPrx) MarshalJSON() ([]byte, error) { return x.Unwrap().MarshalJSON() }
func (x *AnyPrx) UnmarshalJSON(b []byte) error { return unmarshal(b, x) }
func (x *AnyPrx) Print(b *bfr.P) error         { return x.Unwrap().Print(b) }

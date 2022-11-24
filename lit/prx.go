package lit

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"

	"xelf.org/xelf/ast"
	"xelf.org/xelf/bfr"
	"xelf.org/xelf/knd"
	"xelf.org/xelf/typ"
)

type proxy struct {
	Reg *PrxReg
	typ typ.Type
	val reflect.Value
}

func newProxy(reg *PrxReg, t typ.Type, ptr reflect.Value) proxy {
	x := proxy{reg, t, ptr}
	if x.isptr() {
		x.typ = typ.Opt(t)
	}
	return x
}
func (x *proxy) Nil() bool {
	switch x.val.Type().Elem().Kind() {
	case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
		return x.val.Elem().IsNil()
	}
	return false
}
func (x *proxy) Type() typ.Type               { return x.typ }
func (x *proxy) Ptr() interface{}             { return x.val.Interface() }
func (x *proxy) Reflect() reflect.Value       { return x.val.Elem() }
func (x *proxy) new() reflect.Value           { return reflect.New(x.val.Type().Elem()) }
func (x *proxy) with(ptr reflect.Value) proxy { return newProxy(x.Reg, x.typ, ptr) }
func (x *proxy) unmarshal(b []byte, mut Mut) error {
	return ReadInto(bytes.NewReader(b), "", mut)
}
func (x *proxy) isptr() bool { return x.val.Type().Elem().Kind() == reflect.Ptr }
func (x *proxy) setNull() error {
	if x.isptr() {
		val := reflect.Zero(x.val.Type().Elem())
		x.val.Elem().Set(val)
	} else {
		x.val.Elem().Set(reflect.Zero(x.val.Type().Elem()))
	}
	return nil
}
func (x *proxy) elem() reflect.Value {
	e := x.val.Elem()
	if !x.isptr() {
		if e.Kind() == reflect.Map && e.IsNil() {
			e.Set(reflect.MakeMap(e.Type()))
		}
		return e
	}
	if !e.IsNil() {
		return e.Elem()
	}
	e = reflect.New(e.Type().Elem())
	if e.Kind() == reflect.Map {
		e.Set(reflect.MakeMap(e.Type()))
	}
	x.val.Elem().Set(e)
	return e.Elem()
}

type IntPrx struct{ proxy }

func (x *IntPrx) NewWith(v reflect.Value) Mut { return &IntPrx{x.with(v)} }

func (x *IntPrx) New() Mut   { return x.NewWith(x.new()) }
func (x *IntPrx) Zero() bool { return x.Nil() || x.value() == 0 }
func (x *IntPrx) Value() Val {
	if x.Nil() {
		return Null{}
	}
	return Int(x.value())
}
func (x *IntPrx) Parse(a ast.Ast) error {
	if isNull(a) {
		return x.setNull()
	}
	if a.Kind != knd.Int {
		return ast.ErrExpect(a, knd.Int)
	}
	n, err := strconv.ParseInt(a.Raw, 10, 64)
	if err != nil {
		return err
	}
	x.elem().SetInt(n)
	return nil
}
func (x *IntPrx) Assign(v Val) error {
	if v == nil || v.Nil() {
		return x.setNull()
	}
	n, err := ToInt(v)
	if err != nil {
		return err
	}
	switch e := x.elem(); e.Kind() {
	case reflect.Int64, reflect.Int, reflect.Int32, reflect.Int16:
		e.SetInt(int64(n))
	case reflect.Uint64, reflect.Uint, reflect.Uint32, reflect.Uint16:
		e.SetUint(uint64(n))
	default:
		return fmt.Errorf("unexpected int proxy target %s", e.Type())
	}
	return nil
}
func (x *IntPrx) value() int64 {
	switch e := x.elem(); e.Kind() {
	case reflect.Int64, reflect.Int, reflect.Int32:
		return e.Int()
	case reflect.Uint64, reflect.Uint, reflect.Uint32:
		return int64(e.Uint())
	default:
		panic(fmt.Errorf("unexpected int proxy target %s", e.Type()))
	}
}
func (x *IntPrx) String() string {
	if x.Nil() {
		return "null"
	}
	return fmt.Sprintf("%d", x.value())
}
func (x *IntPrx) MarshalJSON() ([]byte, error) { return []byte(x.String()), nil }
func (x *IntPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *IntPrx) Print(p *bfr.P) error         { return p.Fmt(x.String()) }

type RealPrx struct{ proxy }

func (x *RealPrx) NewWith(v reflect.Value) Mut { return &RealPrx{x.with(v)} }

func (x *RealPrx) New() Mut   { return x.NewWith(x.new()) }
func (x *RealPrx) Zero() bool { return x.Nil() || x.value() == 0 }
func (x *RealPrx) Value() Val {
	if x.Nil() {
		return Null{}
	}
	return Real(x.value())
}
func (x *RealPrx) Parse(a ast.Ast) error {
	if isNull(a) {
		return x.setNull()
	}
	if a.Kind != knd.Real && a.Kind != knd.Int {
		return ast.ErrExpect(a, knd.Num)
	}
	n, err := strconv.ParseFloat(a.Raw, 64)
	if err != nil {
		return err
	}
	x.elem().SetFloat(n)
	return nil
}
func (x *RealPrx) Assign(v Val) error {
	if v == nil || v.Nil() {
		return x.setNull()
	}
	n, err := ToReal(v)
	if err != nil {
		return err
	}
	switch e := x.elem(); e.Kind() {
	case reflect.Float64, reflect.Float32:
		e.SetFloat(float64(n))
	default:
		return fmt.Errorf("unexpected real proxy target %s", e.Type())
	}
	return nil
}
func (x *RealPrx) value() float64 {
	switch e := x.elem(); e.Kind() {
	case reflect.Float64, reflect.Float32:
		return e.Float()
	default:
		panic(fmt.Errorf("unexpected real proxy target %s", e.Type()))
	}
}
func (x *RealPrx) String() string {
	if x.Nil() {
		return "null"
	}
	return fmt.Sprintf("%g", x.value())
}
func (x *RealPrx) MarshalJSON() ([]byte, error) { return []byte(x.String()), nil }
func (x *RealPrx) UnmarshalJSON(b []byte) error { return x.unmarshal(b, x) }
func (x *RealPrx) Print(p *bfr.P) error         { return p.Fmt(x.String()) }
